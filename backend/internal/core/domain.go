package core

import "net/http"

type HTTPErrorStatusCode struct {
	code int
}

func (c HTTPErrorStatusCode) Int() int {
	return c.code
}

var (
	StatusBadRequest          = HTTPErrorStatusCode{code: http.StatusBadRequest}
	StatusUnauthorized        = HTTPErrorStatusCode{code: http.StatusUnauthorized}
	StatusForbidden           = HTTPErrorStatusCode{code: http.StatusForbidden}
	StatusNotFound            = HTTPErrorStatusCode{code: http.StatusNotFound}
	StatusPayloadTooLarge     = HTTPErrorStatusCode{code: http.StatusRequestEntityTooLarge}
	StatusTooManyRequests     = HTTPErrorStatusCode{code: http.StatusTooManyRequests}
	StatusInternalServerError = HTTPErrorStatusCode{code: http.StatusInternalServerError}
)

var (
	ErrNotFound       = NewDomainError("not_found", "resource not found", StatusNotFound)
	ErrJTIBlacklisted = NewDomainError("jti_blacklisted", "jti blacklisted", StatusUnauthorized)
)

type DomainError struct {
	code       string
	msg        string
	statusCode HTTPErrorStatusCode
}

func NewDomainError(code, msg string, statusCode HTTPErrorStatusCode) *DomainError {
	return new(DomainError{code: code, msg: msg, statusCode: statusCode})
}

func (e *DomainError) Code() string {
	return e.code
}

func (e *DomainError) Error() string {
	return e.msg
}

func (e *DomainError) StatusCode() HTTPErrorStatusCode {
	return e.statusCode
}

func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.code == t.code
}
