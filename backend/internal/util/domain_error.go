package util

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
