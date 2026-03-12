package user

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
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
	ErrAuthorizationIdIsProcessed  = errors.New("the authorization id is currently processed")
	ErrAuthorizationIdIsRegistered = errors.New("the authorization id is already registered")

	ErrUserIdNotFoundInContext = errors.New("user id not found in context")
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

func (s *Service) CreateNewUser(ctx context.Context, dailyScreenLimit *int, screenLimits []struct{ Start, End int }, displayName string, languageCode string) (*User, error) {
	// 登録用アクセストークン取得
	accessToken, ok := util.AccessTokenFromContext(ctx)
	if !ok {
		return nil, errors.New("access token not found in context")
	}
	authorizationId, err := s.jwtService.VerifyRegisterToken(accessToken)
	if err != nil {
		return nil, err
	}

	// Entityの検証
	domainDailyScreenLimitDuration, err := NewDailyScreenTimeLimit(dailyScreenLimit)
	if err != nil {
		return nil, err
	}
	domainDisplayName, err := NewDisplayName(displayName)
	if err != nil {
		return nil, err
	}
	domainLanguageCode, err := NewLanguageCode(languageCode)
	if err != nil {
		return nil, err
	}
	domainDailyScreenTimeLimitRangeSet, err := NewDailyScreenTimeLimitRangeSet(screenLimits)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	// 勧告ロック
	authorizationIdHash := sha256.Sum256(authorizationId[:])
	acquired, err := q.TryAcquireAdvisoryXactLock(ctx, int64(binary.BigEndian.Uint64(authorizationIdHash[:8])))
	if err != nil {
		return nil, err
	}
	if !acquired {
		return nil, ErrAuthorizationIdIsProcessed
	}

	// すでに登録しているか
	authorizationIdCount, err := q.CountUsersByAuthorization(ctx, authorizationId)
	if err != nil {
		return nil, err
	}
	if authorizationIdCount >= 1 {
		return nil, ErrAuthorizationIdIsRegistered
	}

	// Userの保存
	saveUser, err := q.SaveUser(ctx, sqlc.SaveUserParams{
		DisplayName:  domainDisplayName.ToString(),
		LanguageCode: domainLanguageCode.ToString(),
		// NOTE: domainDailyScreenLimitDurationはNewDailyScreenTimeLimitである、非nil
		DailyScreenTimeSeconds:    *domainDailyScreenLimitDuration.ToInt(),
		UserAuthorizationPublicID: authorizationId,
	})
	if err != nil {
		return nil, err
	}

	// rangesの保存
	saveUserScreenTimeRangesParams := make([]sqlc.SaveUserScreenTimeRangesParams, len(*domainDailyScreenTimeLimitRangeSet))
	for i, domainRange := range *domainDailyScreenTimeLimitRangeSet {
		saveUserScreenTimeRangesParams[i] = sqlc.SaveUserScreenTimeRangesParams{
			MUserID:              saveUser.MUserID,
			ScreenTimeRangeStart: database_d.Seconds(domainRange.StartTimeSeconds),
			ScreenTimeRangeEnd:   database_d.Seconds(domainRange.EndTimeSeconds),
		}
	}
	if _, err := q.SaveUserScreenTimeRanges(ctx, saveUserScreenTimeRangesParams); err != nil {
		return nil, err
	}

	// publicIdを取得するため、selectで取得する。(:copyFromはRETURNING使えないっぽい)
	screenTimeLimitRanges, err := q.GetUserScreenTimeRanges(ctx, saveUser.PublicID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	var remainingSeconds *int
	if !domainDailyScreenLimitDuration.IsInfinity() {
		remainingSeconds = domainDailyScreenLimitDuration.ToInt()
	}

	screenTimeLimitRangesDTO := make([]struct {
		Id                       uuid.UUID
		StartSeconds, EndSeconds int
	}, len(screenTimeLimitRanges))
	for i, pg := range screenTimeLimitRanges {
		screenTimeLimitRangesDTO[i] = struct {
			Id           uuid.UUID
			StartSeconds int
			EndSeconds   int
		}{
			Id:           pg.PublicID,
			StartSeconds: (int)(pg.ScreenTimeRangeStart),
			EndSeconds:   (int)(pg.ScreenTimeRangeEnd),
		}
	}

	return NewUser(
		saveUser.PublicID,
		domainDisplayName.ToString(),
		domainLanguageCode.ToString(),
		saveUser.JoinedAt.UTC(),
		screenTimeLimitRangesDTO,
		remainingSeconds,
		remainingSeconds,
	), nil
}

func (s *Service) EditUser(ctx context.Context, newDisplayName, newLanguageCode *string, newDailyScreenLimit *int, newScreenLimits *[]struct{ Start, End int }) (*User, error) {
	userId, ok := util.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrUserIdNotFoundInContext
	}

	// Entityの検証
	var domainNewDisplayName *DisplayName
	if newDisplayName != nil {
		d, err := NewDisplayName(*newDisplayName)
		if err != nil {
			return nil, err
		}
		domainNewDisplayName = d
	}
	var domainNewScreenTime *DailyScreenTimeLimit
	if newDailyScreenLimit != nil {
		d, err := NewDailyScreenTimeLimit(newDailyScreenLimit)
		if err != nil {
			return nil, err
		}
		domainNewScreenTime = d
	}
	var domainNewLanguageCode *LanguageCode
	if newLanguageCode != nil {
		d, err := NewLanguageCode(*newLanguageCode)
		if err != nil {
			return nil, err
		}
		domainNewLanguageCode = d
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	totalWatchSecondsToday, err := q.GetTotalWatchSeconds(ctx, userId)
	if err != nil {
		return nil, err
	}

	updateUserProfile, err := q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		NewDisplayName:            (*string)(domainNewDisplayName),
		NewDailyScreenTimeSeconds: domainNewScreenTime.ToInt(),
		NewLanguageCode:           (*string)(domainNewLanguageCode),
		UserPublicID:              userId,
	})
	if err != nil {
		return nil, err
	}

	// screen limitsの更新
	if newScreenLimits != nil {
		// domainの検証
		domainDailyScreenTimeLimitRangeSet, err := NewDailyScreenTimeLimitRangeSet(*newScreenLimits)
		if err != nil {
			return nil, err
		}

		// sqlcのparamに詰め替え
		saveUserScreenTimeRangesParams := make([]sqlc.SaveUserScreenTimeRangesParams, len(*domainDailyScreenTimeLimitRangeSet))
		for i, domainRange := range *domainDailyScreenTimeLimitRangeSet {
			saveUserScreenTimeRangesParams[i] = sqlc.SaveUserScreenTimeRangesParams{
				MUserID:              updateUserProfile.MUserID,
				ScreenTimeRangeStart: database_d.Seconds(domainRange.StartTimeSeconds),
				ScreenTimeRangeEnd:   database_d.Seconds(domainRange.EndTimeSeconds),
			}
		}

		// sql発行
		if err := q.RemoveScreenTimeRangesByUserId(ctx, updateUserProfile.MUserID); err != nil {
			return nil, err
		}
		if _, err := q.SaveUserScreenTimeRanges(ctx, saveUserScreenTimeRangesParams); err != nil {
			return nil, err
		}
	}

	// publicIdを取得するため、再度selectを実行
	screenTimeLimitRanges, err := q.GetUserScreenTimeRanges(ctx, userId)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	screenTimeLimitRangesDTO := make([]struct {
		Id                       uuid.UUID
		StartSeconds, EndSeconds int
	}, len(screenTimeLimitRanges))
	for i, pg := range screenTimeLimitRanges {
		screenTimeLimitRangesDTO[i] = struct {
			Id           uuid.UUID
			StartSeconds int
			EndSeconds   int
		}{
			Id:           pg.PublicID,
			StartSeconds: (int)(pg.ScreenTimeRangeStart),
			EndSeconds:   (int)(pg.ScreenTimeRangeEnd),
		}
	}

	var dailyScreenLimit, remainingSeconds *int
	if updateUserProfile.DailyScreenTimeSeconds < 24*60*60 {
		dailyScreenLimit = &updateUserProfile.DailyScreenTimeSeconds
		rem := max(updateUserProfile.DailyScreenTimeSeconds-totalWatchSecondsToday, 0)
		remainingSeconds = &rem
	}

	return NewUser(
		userId,
		updateUserProfile.DisplayName,
		updateUserProfile.LanguageCode,
		updateUserProfile.JoinedAt.UTC(),
		screenTimeLimitRangesDTO,
		dailyScreenLimit,
		remainingSeconds,
	), nil
}

func (s *Service) GetUserStatus(ctx context.Context) (*User, error) {
	userId, ok := util.UserIDFromContext(ctx)
	if !ok {
		return nil, ErrUserIdNotFoundInContext
	}

	q := sqlc.New(s.db)

	userProfile, err := q.GetUserProfile(ctx, userId)
	if err != nil {
		return nil, err
	}

	screenLimitRanges, err := q.GetUserScreenTimeRanges(ctx, userId)
	if err != nil {
		return nil, err
	}
	screenTimeLimitRanges := make([]struct {
		Id                       uuid.UUID
		StartSeconds, EndSeconds int
	}, len(screenLimitRanges))
	for i, screenLimitRange := range screenLimitRanges {
		screenTimeLimitRanges[i] = struct {
			Id                       uuid.UUID
			StartSeconds, EndSeconds int
		}{
			Id:           screenLimitRange.PublicID,
			StartSeconds: int(screenLimitRange.ScreenTimeRangeStart),
			EndSeconds:   int(screenLimitRange.ScreenTimeRangeEnd),
		}
	}

	totalWatchSecondsToday, err := q.GetTotalWatchSeconds(ctx, userId)
	if err != nil {
		return nil, err
	}

	var remainingSeconds, dailyScreenTimeLimit *int
	if userProfile.DailyScreenTimeSeconds < 24*60*60 {
		dailyScreenTimeLimit = &userProfile.DailyScreenTimeSeconds
		rem := max(0, userProfile.DailyScreenTimeSeconds-totalWatchSecondsToday)
		remainingSeconds = &rem
	}

	return NewUser(
		userId,
		userProfile.DisplayName,
		userProfile.LanguageCode,
		userProfile.JoinedAt,
		screenTimeLimitRanges,
		dailyScreenTimeLimit,
		remainingSeconds,
	), nil
}

func (s *Service) RemoveUser(ctx context.Context) error {
	userId, ok := util.UserIDFromContext(ctx)
	if !ok {
		return ErrUserIdNotFoundInContext
	}

	q := sqlc.New(s.db)
	if err := q.RemoveUser(ctx, sqlc.RemoveUserParams{
		LeaveReasonCode: 0,
		UserPublicID:    userId,
	}); err != nil {
		return err
	}

	return nil
}
