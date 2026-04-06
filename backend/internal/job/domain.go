package job

import (
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

var (
	ErrInvalidBatchStatusCode = core.NewDomainError("job.invalid_batch_status_code", "invalid batch status code", core.StatusBadRequest)

	batchStatusCodeMap = []struct {
		code BatchStatusCode
		str  string
	}{
		{code: 0, str: "pending"},
		{code: 1, str: "running"},
		{code: 2, str: "completed"},
		{code: 3, str: "failed"},
	}
)

type BatchStatusCode int

func NewBatchStatusCode(str string) (_ BatchStatusCode, err error) {
	defer util.Wrap(&err, "job.NewBatchStatusCode")

	for _, c := range batchStatusCodeMap {
		if str == c.str {
			return c.code, nil
		}
	}

	return -1, ErrInvalidBatchStatusCode
}

func (b BatchStatusCode) String() string {
	for _, c := range batchStatusCodeMap {
		if c.code == b {
			return c.str
		}
	}

	return "pending"
}
