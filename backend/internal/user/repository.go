package user

import (
	"context"
	"errors"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/dbtype"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	FindForUpdate(ctx context.Context, userID uuid.UUID) (*User, error)
	FindScreenTimeRanges(ctx context.Context, userID uuid.UUID) (*DailyScreenTimeLimitRangeSet, error)
	Save(ctx context.Context, u *User, authorizationID uuid.UUID) (int64, error)
	Update(ctx context.Context, u *User) (int64, error)
	SaveScreenTimeRanges(ctx context.Context, mUserID int64, rangeSet *DailyScreenTimeLimitRangeSet) error
	Remove(ctx context.Context, userID uuid.UUID, leaveReasonCode LeaveReasonCode) error
	CountByAuthorization(ctx context.Context, authorizationID uuid.UUID) (int32, error)
	DeleteLeftByAuthorization(ctx context.Context, authorizationID uuid.UUID) (int64, error)
}

type userRepositoryImpl struct {
	q sqlc.Querier
}

func NewUserRepository(q sqlc.Querier) UserRepository {
	return &userRepositoryImpl{
		q: q,
	}
}

func (r *userRepositoryImpl) FindForUpdate(ctx context.Context, userID uuid.UUID) (_ *User, err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).FindForUpdate(userID=%s)", userID)

	row, err := r.q.GetUserForUpdate(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}

	var dailyScreenLimit *int
	if !IsUnlimitedScreenTimeSeconds(row.DailyScreenTimeSeconds) {
		dailyScreenLimit = &row.DailyScreenTimeSeconds
	}

	u, err := NewUser(
		row.DisplayName,
		row.LanguageCode,
		dailyScreenLimit,
		WithUserID(row.PublicID),
		WithUserJoinedAt(row.JoinedAt.UTC()),
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *userRepositoryImpl) FindScreenTimeRanges(ctx context.Context, userID uuid.UUID) (_ *DailyScreenTimeLimitRangeSet, err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).FindScreenTimeRanges(userID=%s)", userID)

	rows, err := r.q.ListScreenTimeRanges(ctx, userID)
	if err != nil {
		return nil, err
	}

	ranges := make([]DailyScreenTimeLimitRange, len(rows))
	for i, row := range rows {
		ranges[i] = DailyScreenTimeLimitRange{
			StartTimeSeconds: int(row.ScreenTimeRangeStart),
			EndTimeSeconds:   int(row.ScreenTimeRangeEnd),
		}
	}
	return &DailyScreenTimeLimitRangeSet{Ranges: ranges}, nil
}

func (r *userRepositoryImpl) Save(ctx context.Context, u *User, authorizationID uuid.UUID) (_ int64, err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).Save(authorizationID=%s)", authorizationID)

	mUserID, err := r.q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               u.DisplayName.String(),
		LanguageCode:              u.LanguageCode.String(),
		DailyScreenTimeSeconds:    u.ScreenTimeLimit.Seconds(),
		JoinedAt:                  u.JoinedAt,
		PublicID:                  u.ID,
		UserAuthorizationPublicID: authorizationID,
	})
	if err != nil {
		return 0, err
	}
	return mUserID, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, u *User) (_ int64, err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).Update")

	mUserID, err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		DisplayName:            u.DisplayName.String(),
		LanguageCode:           u.LanguageCode.String(),
		DailyScreenTimeSeconds: u.ScreenTimeLimit.Seconds(),
		UserPublicID:           u.ID,
	})
	if err != nil {
		return 0, err
	}
	return mUserID, nil
}

func (r *userRepositoryImpl) SaveScreenTimeRanges(ctx context.Context, mUserID int64, rangeSet *DailyScreenTimeLimitRangeSet) (err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).SaveScreenTimeRanges(mUserID=%d)", mUserID)

	params := make([]sqlc.BulkInsertScreenTimeRangesParams, len(rangeSet.Ranges))
	for i, r := range rangeSet.Ranges {
		params[i] = sqlc.BulkInsertScreenTimeRangesParams{
			MUserID:              mUserID,
			ScreenTimeRangeStart: dbtype.Seconds(r.StartTimeSeconds),
			ScreenTimeRangeEnd:   dbtype.Seconds(r.EndTimeSeconds),
		}
	}
	if _, err := r.q.BulkInsertScreenTimeRanges(ctx, params); err != nil {
		return err
	}
	return nil
}

func (r *userRepositoryImpl) Remove(ctx context.Context, userID uuid.UUID, leaveReasonCode LeaveReasonCode) (err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).Remove(userID=%s)", userID)

	if err := r.q.ArchiveUser(ctx, sqlc.ArchiveUserParams{
		LeaveReasonCode: int(leaveReasonCode),
		UserPublicID:    userID,
	}); err != nil {
		return err
	}

	return nil
}

func (r *userRepositoryImpl) CountByAuthorization(ctx context.Context, authorizationID uuid.UUID) (_ int32, err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).CountByAuthorization(authorizationID=%s)", authorizationID)

	count, err := r.q.CountUsersByAuthorization(ctx, authorizationID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepositoryImpl) DeleteLeftByAuthorization(ctx context.Context, authorizationID uuid.UUID) (_ int64, err error) {
	defer util.Wrap(&err, "user.(*userRepositoryImpl).DeleteLeftByAuthorization(authorizationID=%s)", authorizationID)

	hUserID, err := r.q.DeleteLeftUserByAuthorization(ctx, authorizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, core.ErrNotFound
		}
		return 0, err
	}
	return hUserID, nil
}

var _ UserRepository = (*userRepositoryImpl)(nil)
