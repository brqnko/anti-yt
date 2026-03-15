package video

import "github.com/jackc/pgx/v5/pgxpool"

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) (*Service, error) {
	return &Service{
		db: db,
	}, nil
}
