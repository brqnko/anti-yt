package dbtype

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Seconds int

func (s *Seconds) ScanTime(v pgtype.Time) error {
	if !v.Valid {
		*s = 0
		return nil
	}
	*s = Seconds(v.Microseconds / 1_000_000)
	return nil
}

func (s Seconds) TimeValue() (pgtype.Time, error) {
	return pgtype.Time{
		Microseconds: int64(s) * 1_000_000,
		Valid:        true,
	}, nil
}
