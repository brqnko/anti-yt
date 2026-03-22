package auth

import (
	"context"
	"fmt"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
)

type AuthorizationRepository interface {
	Save(ctx context.Context, authorization Authorization) (int64, error)
}

func NewAuthorizationRepository(q sqlc.Querier) AuthorizationRepository {
	return &authorizationRepositoryImpl{
		q: q,
	}
}

type authorizationRepositoryImpl struct {
	q sqlc.Querier
}

func (a *authorizationRepositoryImpl) Save(ctx context.Context, authorization Authorization) (int64, error) {
	saveAuthorization, err := a.q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         authorization.Issuer,
		Sub:            authorization.Sub,
		LastLoggedInAt: authorization.LastLoggedInAt,
		PublicID:       authorization.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to authorizationRepository.Save: %w", err)
	}

	return saveAuthorization.MUserAuthorizationID, nil
}

var _ AuthorizationRepository = (*authorizationRepositoryImpl)(nil)
