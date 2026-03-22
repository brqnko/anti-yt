package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidAuthorizationIDProcessed  = errors.New("invalid authorization id: currently processed")
	ErrInvalidAuthorizationIDRegistered = errors.New("invalid authorization id: already registered")
)

type Service struct {
	db         *pgxpool.Pool
	jwtService jwt_d.JWTService
}

func NewService(db *pgxpool.Pool, jwtService jwt_d.JWTService) (*Service, error) {
	return &Service{
		db:         db,
		jwtService: jwtService,
	}, nil
}

func (s *Service) CreateNewUser(ctx context.Context, dailyScreenLimit *int, screenLimits []struct{ Start, End int }, displayName string, languageCode string) (User, error) {
	// 登録用アクセストークン取得
	accessToken, ok := util.AccessTokenFromContext(ctx)
	if !ok {
		return User{}, errors.New("access token not found in context")
	}
	authorizationID, err := s.jwtService.VerifyRegisterToken(accessToken)
	if err != nil {
		return User{}, err
	}

	// Entityの検証
	domainDailyScreenLimitDuration, err := NewDailyScreenTimeLimit(dailyScreenLimit)
	if err != nil {
		return User{}, err
	}
	domainDisplayName, err := NewDisplayName(displayName)
	if err != nil {
		return User{}, err
	}
	domainLanguageCode, err := NewLanguageCode(languageCode)
	if err != nil {
		return User{}, err
	}
	domainDailyScreenTimeLimitRangeSet, err := NewDailyScreenTimeLimitRangeSet(screenLimits)
	if err != nil {
		return User{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	// 勧告ロック
	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, util.Sha256Int64(authorizationID[:]))
	if err != nil {
		return User{}, fmt.Errorf("tryAcquireAdvisoryXactLock: %w", err)
	}
	if !acquired {
		return User{}, ErrInvalidAuthorizationIDProcessed
	}

	// すでに登録しているか
	authorizationIDCount, err := q.CountUsersByAuthorization(ctx, authorizationID)
	if err != nil {
		return User{}, fmt.Errorf("countUsersByAuthorization: %w", err)
	}
	if authorizationIDCount >= 1 {
		return User{}, ErrInvalidAuthorizationIDRegistered
	}

	// Userの保存
	saveUser, err := q.SaveUser(ctx, sqlc.SaveUserParams{
		DisplayName:  domainDisplayName.String(),
		LanguageCode: domainLanguageCode.String(),
		// NOTE: domainDailyScreenLimitDurationはNewDailyScreenTimeLimitである、非nil
		DailyScreenTimeSeconds:    domainDailyScreenLimitDuration.ToInt(),
		UserAuthorizationPublicID: authorizationID,
	})
	if err != nil {
		return User{}, fmt.Errorf("saveUser: %w", err)
	}

	// rangesの保存
	saveUserScreenTimeRangesParams := make([]sqlc.SaveUserScreenTimeRangesParams, len(domainDailyScreenTimeLimitRangeSet))
	for i, domainRange := range domainDailyScreenTimeLimitRangeSet {
		saveUserScreenTimeRangesParams[i] = sqlc.SaveUserScreenTimeRangesParams{
			MUserID:              saveUser.MUserID,
			ScreenTimeRangeStart: database_d.Seconds(domainRange.StartTimeSeconds),
			ScreenTimeRangeEnd:   database_d.Seconds(domainRange.EndTimeSeconds),
		}
	}
	if _, err := q.SaveUserScreenTimeRanges(ctx, saveUserScreenTimeRangesParams); err != nil {
		return User{}, fmt.Errorf("saveUserScreenTimeRanges: %w", err)
	}

	// publicIdを取得するため、selectで取得する。(:copyFromはRETURNING使えないっぽい)
	screenTimeLimitRanges, err := q.GetUserScreenTimeRanges(ctx, saveUser.PublicID)
	if err != nil {
		return User{}, fmt.Errorf("getUserScreenTimeRanges: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit: %w", err)
	}

	remainingSeconds := -1
	if !domainDailyScreenLimitDuration.IsInfinity() {
		remainingSeconds = domainDailyScreenLimitDuration.ToInt()
	}

	screenTimeLimitRangesDTO := make([]struct {
		ID                       uuid.UUID
		StartSeconds, EndSeconds int
	}, len(screenTimeLimitRanges))
	for i, pg := range screenTimeLimitRanges {
		screenTimeLimitRangesDTO[i] = struct {
			ID           uuid.UUID
			StartSeconds int
			EndSeconds   int
		}{
			ID:           pg.PublicID,
			StartSeconds: (int)(pg.ScreenTimeRangeStart),
			EndSeconds:   (int)(pg.ScreenTimeRangeEnd),
		}
	}

	return NewUser(
		saveUser.PublicID,
		domainDisplayName.String(),
		domainLanguageCode.String(),
		saveUser.JoinedAt.UTC(),
		screenTimeLimitRangesDTO,
		remainingSeconds,
		remainingSeconds,
	), nil
}

func (s *Service) EditUser(ctx context.Context, newDisplayName, newLanguageCode *string, newDailyScreenLimit *int, newScreenLimits *[]struct{ Start, End int }) (User, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return User{}, err
	}

	// Entityの検証
	var newDisplayNameStr *string
	if newDisplayName != nil {
		d, err := NewDisplayName(*newDisplayName)
		if err != nil {
			return User{}, err
		}
		s := d.String()
		newDisplayNameStr = &s
	}
	var newScreenTimeInt *int
	if newDailyScreenLimit != nil {
		d, err := NewDailyScreenTimeLimit(newDailyScreenLimit)
		if err != nil {
			return User{}, err
		}
		v := d.ToInt()
		newScreenTimeInt = &v
	}
	var newLanguageCodeStr *string
	if newLanguageCode != nil {
		d, err := NewLanguageCode(*newLanguageCode)
		if err != nil {
			return User{}, err
		}
		s := d.String()
		newLanguageCodeStr = &s
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	watchStats, err := q.GetTotalWatchSeconds(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("getTotalWatchSeconds: %w", err)
	}

	updateUserProfile, err := q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		NewDisplayName:            newDisplayNameStr,
		NewDailyScreenTimeSeconds: newScreenTimeInt,
		NewLanguageCode:           newLanguageCodeStr,
		UserPublicID:              userID,
	})
	if err != nil {
		return User{}, fmt.Errorf("updateUserProfile: %w", err)
	}

	// screen limitsの更新
	if newScreenLimits != nil {
		// domainの検証
		domainDailyScreenTimeLimitRangeSet, err := NewDailyScreenTimeLimitRangeSet(*newScreenLimits)
		if err != nil {
			return User{}, err
		}

		// sqlcのparamに詰め替え
		saveUserScreenTimeRangesParams := make([]sqlc.SaveUserScreenTimeRangesParams, len(domainDailyScreenTimeLimitRangeSet))
		for i, domainRange := range domainDailyScreenTimeLimitRangeSet {
			saveUserScreenTimeRangesParams[i] = sqlc.SaveUserScreenTimeRangesParams{
				MUserID:              updateUserProfile.MUserID,
				ScreenTimeRangeStart: database_d.Seconds(domainRange.StartTimeSeconds),
				ScreenTimeRangeEnd:   database_d.Seconds(domainRange.EndTimeSeconds),
			}
		}

		// sql発行
		if err := q.RemoveScreenTimeRangesByUserId(ctx, updateUserProfile.MUserID); err != nil {
			return User{}, fmt.Errorf("removeScreenTimeRangesByUserId: %w", err)
		}
		if _, err := q.SaveUserScreenTimeRanges(ctx, saveUserScreenTimeRangesParams); err != nil {
			return User{}, fmt.Errorf("saveUserScreenTimeRanges: %w", err)
		}
	}

	// publicIdを取得するため、再度selectを実行
	screenTimeLimitRanges, err := q.GetUserScreenTimeRanges(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("getUserScreenTimeRanges: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit: %w", err)
	}

	screenTimeLimitRangesDTO := make([]struct {
		ID                       uuid.UUID
		StartSeconds, EndSeconds int
	}, len(screenTimeLimitRanges))
	for i, pg := range screenTimeLimitRanges {
		screenTimeLimitRangesDTO[i] = struct {
			ID           uuid.UUID
			StartSeconds int
			EndSeconds   int
		}{
			ID:           pg.PublicID,
			StartSeconds: (int)(pg.ScreenTimeRangeStart),
			EndSeconds:   (int)(pg.ScreenTimeRangeEnd),
		}
	}

	remainingSeconds := CalcRemainingSeconds(updateUserProfile.DailyScreenTimeSeconds, watchStats.TodayWatchTotal)
	dailyScreenLimit := -1
	if remainingSeconds != -1 {
		dailyScreenLimit = updateUserProfile.DailyScreenTimeSeconds
	}

	return NewUser(
		userID,
		updateUserProfile.DisplayName,
		updateUserProfile.LanguageCode,
		updateUserProfile.JoinedAt.UTC(),
		screenTimeLimitRangesDTO,
		dailyScreenLimit,
		remainingSeconds,
	), nil
}

func (s *Service) GetUserStatus(ctx context.Context) (User, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return User{}, err
	}

	q := sqlc.New(s.db)

	userProfile, err := q.GetUserProfile(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("getUserProfile: %w", err)
	}

	screenLimitRanges, err := q.GetUserScreenTimeRanges(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("getUserScreenTimeRanges: %w", err)
	}
	screenTimeLimitRanges := make([]struct {
		ID                       uuid.UUID
		StartSeconds, EndSeconds int
	}, len(screenLimitRanges))
	for i, screenLimitRange := range screenLimitRanges {
		screenTimeLimitRanges[i] = struct {
			ID                       uuid.UUID
			StartSeconds, EndSeconds int
		}{
			ID:           screenLimitRange.PublicID,
			StartSeconds: int(screenLimitRange.ScreenTimeRangeStart),
			EndSeconds:   int(screenLimitRange.ScreenTimeRangeEnd),
		}
	}

	watchStats, err := q.GetTotalWatchSeconds(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("getTotalWatchSeconds: %w", err)
	}

	remainingSeconds := CalcRemainingSeconds(watchStats.DailyLimitSeconds, watchStats.TodayWatchTotal)
	dailyScreenTimeLimit := -1
	if remainingSeconds != -1 {
		dailyScreenTimeLimit = watchStats.DailyLimitSeconds
	}

	return NewUser(
		userID,
		userProfile.DisplayName,
		userProfile.LanguageCode,
		userProfile.JoinedAt,
		screenTimeLimitRanges,
		dailyScreenTimeLimit,
		remainingSeconds,
	), nil
}

func (s *Service) RemoveUser(ctx context.Context) error {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	q := sqlc.New(s.db)
	if err := q.RemoveUser(ctx, sqlc.RemoveUserParams{
		LeaveReasonCode: 0,
		UserPublicID:    userID,
	}); err != nil {
		return fmt.Errorf("removeUser: %w", err)
	}

	return nil
}
