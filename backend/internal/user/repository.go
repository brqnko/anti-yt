package user

import (
	"context"
	"errors"
	"fmt"


	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
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
}

type userRepositoryImpl struct {
	q sqlc.Querier
}

func NewUserRepository(q sqlc.Querier) UserRepository {
	return &userRepositoryImpl{
		q: q,
	}
}

func (r *userRepositoryImpl) FindForUpdate(ctx context.Context, userID uuid.UUID) (*User, error) {
	row, err := r.q.GetUserForUpdate(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to getUserForUpdate(userRepository.FindForUpdate): %w", err)
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
		return nil, fmt.Errorf("failed to newUser(userRepository.FindForUpdate): %w", err)
	}

	return u, nil
}

func (r *userRepositoryImpl) FindScreenTimeRanges(ctx context.Context, userID uuid.UUID) (*DailyScreenTimeLimitRangeSet, error) {
	rows, err := r.q.ListScreenTimeRanges(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to getUserScreenTimeRanges(userRepository.FindScreenTimeRanges): %w", err)
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

func (r *userRepositoryImpl) Save(ctx context.Context, u *User, authorizationID uuid.UUID) (int64, error) {
	mUserID, err := r.q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               u.DisplayName.String(),
		LanguageCode:              u.LanguageCode.String(),
		DailyScreenTimeSeconds:    u.ScreenTimeLimit.Seconds(),
		JoinedAt:                  u.JoinedAt,
		PublicID:                  u.ID,
		UserAuthorizationPublicID: authorizationID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to insertUser(userRepository.Save): %w", err)
	}
	return mUserID, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, u *User) (int64, error) {
	mUserID, err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		DisplayName:            u.DisplayName.String(),
		LanguageCode:           u.LanguageCode.String(),
		DailyScreenTimeSeconds: u.ScreenTimeLimit.Seconds(),
		UserPublicID:           u.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to updateUser(userRepository.Update): %w", err)
	}
	return mUserID, nil
}

func (r *userRepositoryImpl) SaveScreenTimeRanges(ctx context.Context, mUserID int64, rangeSet *DailyScreenTimeLimitRangeSet) error {
	params := make([]sqlc.BulkInsertScreenTimeRangesParams, len(rangeSet.Ranges))
	for i, r := range rangeSet.Ranges {
		params[i] = sqlc.BulkInsertScreenTimeRangesParams{
			MUserID:              mUserID,
			ScreenTimeRangeStart: database_d.Seconds(r.StartTimeSeconds),
			ScreenTimeRangeEnd:   database_d.Seconds(r.EndTimeSeconds),
		}
	}
	if _, err := r.q.BulkInsertScreenTimeRanges(ctx, params); err != nil {
		return fmt.Errorf("failed to saveUserScreenTimeRanges(userRepository.SaveScreenTimeRanges): %w", err)
	}
	return nil
}

func (r *userRepositoryImpl) Remove(ctx context.Context, userID uuid.UUID, leaveReasonCode LeaveReasonCode) error {
	if err := r.q.ArchiveUser(ctx, sqlc.ArchiveUserParams{
		LeaveReasonCode: int(leaveReasonCode),
		UserPublicID:    userID,
	}); err != nil {
		return fmt.Errorf("failed to removeUser(userRepository.Remove): %w", err)
	}

	return nil
}

func (r *userRepositoryImpl) CountByAuthorization(ctx context.Context, authorizationID uuid.UUID) (int32, error) {
	count, err := r.q.CountUsersByAuthorization(ctx, authorizationID)
	if err != nil {
		return 0, fmt.Errorf("failed to countUsersByAuthorization(userRepository.CountByAuthorization): %w", err)
	}
	return count, nil
}

var _ UserRepository = (*userRepositoryImpl)(nil)
