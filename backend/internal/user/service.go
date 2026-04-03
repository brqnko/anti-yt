package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidAuthorizationIDRegistered = core.NewDomainError("user.already_registered", "invalid authorization id: already registered")
)

type Service struct {
	db         *pgxpool.Pool
	jwtService jwt_d.Service
	serverURL  string

	userQS UserQueryService
}

func NewService(db *pgxpool.Pool, jwtService jwt_d.Service, serverURL string) *Service {
	return &Service{
		db:         db,
		jwtService: jwtService,
		serverURL:  serverURL,
		userQS:     NewUserQueryService(db),
	}
}

func (s *Service) CreateNewUser(ctx context.Context, accessToken string, dailyScreenLimit *int, screenLimits []struct{ Start, End int }, displayName string, languageCode string, loc *time.Location) (_ *User, _ string, _ time.Time, err error) {
	defer util.Wrap(&err, "user.(*Service).CreateNewUser")

	authorizationID, jti, err := s.jwtService.VerifyRegisterToken(accessToken)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user, err := NewUser(displayName, languageCode, dailyScreenLimit)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	rangeSet, err := NewDailyScreenTimeLimitRangeSet(screenLimits, loc)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// 勧告ロック
	if err := database_d.TryAdLock(ctx, q, authorizationID[:]); err != nil {
		return nil, "", time.Time{}, err
	}

	// すでに登録しているか
	authorizationIDCount, err := NewUserRepository(q).CountByAuthorization(ctx, authorizationID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	if authorizationIDCount >= 1 {
		return nil, "", time.Time{}, ErrInvalidAuthorizationIDRegistered
	}

	// Userの保存
	mUserID, err := NewUserRepository(q).Save(ctx, user, authorizationID)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// rangesの保存
	if err := NewUserRepository(q).SaveScreenTimeRanges(ctx, mUserID, rangeSet); err != nil {
		return nil, "", time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", time.Time{}, err
	}

	// ユーザー作成成功後、UserAccessTokenを発行する（RegisterTokenからの切り替え）
	newAccessToken, accessTokenExpiresAt, err := s.jwtService.SignUserAccessToken(user.ID, jti, s.serverURL)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return user, newAccessToken, accessTokenExpiresAt, nil
}

func (s *Service) EditUser(ctx context.Context, userID uuid.UUID, newDisplayName, newLanguageCode *string, newDailyScreenLimit *int, newScreenLimits *[]struct{ Start, End int }, loc *time.Location) (_ *User, err error) {
	defer util.Wrap(&err, "user.(*Service).EditUser(userID=%s)", userID)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	u, err := NewUserRepository(q).FindForUpdate(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := u.SetDisplayName(newDisplayName); err != nil {
		return nil, err
	}
	if err := u.SetLanguageCode(newLanguageCode); err != nil {
		return nil, err
	}
	if err := u.SetScreenTimeLimit(newDailyScreenLimit); err != nil {
		return nil, err
	}

	mUserID, err := NewUserRepository(q).Update(ctx, u)
	if err != nil {
		return nil, err
	}

	if newScreenLimits != nil {
		rangeSet, err := NewDailyScreenTimeLimitRangeSet(*newScreenLimits, loc)
		if err != nil {
			return nil, err
		}

		if err := q.DeleteScreenTimeRangesByUserID(ctx, mUserID); err != nil {
			return nil, err
		}
		if err := NewUserRepository(q).SaveScreenTimeRanges(ctx, mUserID, rangeSet); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) GetUserStatus(ctx context.Context, userID uuid.UUID) (_ UserStatusView, err error) {
	defer util.Wrap(&err, "user.(*Service).GetUserStatus")

	view, err := s.userQS.Find(ctx, userID)
	if err != nil {
		return UserStatusView{}, err
	}
	return view, nil
}

func (s *Service) RemoveUser(ctx context.Context, userID uuid.UUID) (err error) {
	defer util.Wrap(&err, "user.(*Service).RemoveUser")

	if err := NewUserRepository(sqlc.New(s.db)).Remove(ctx, userID, LeaveReasonCode(0)); err != nil {
		return err
	}

	return nil
}
