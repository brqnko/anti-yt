package core

var (
	ErrNotFound       = NewDomainError("not_found", "resource not found")
	ErrJTIBlacklisted = NewDomainError("jti_blacklisted", "jti blacklisted")
)

type DomainError struct {
	code string
	msg  string
}

func NewDomainError(code, msg string) *DomainError {
	return &DomainError{code: code, msg: msg}
}

func (e *DomainError) Code() string {
	return e.code
}

func (e *DomainError) Error() string {
	return e.msg
}

func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.code == t.code
}
