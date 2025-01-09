package errors

import (
	"errors"
	"fmt"
)

// [4xx] Client Errors

var ErrUnauthenticated = errors.New("http: unauthenticated")
var ErrForbidden = errors.New("http: forbidden")
var ErrNotFound = errors.New("http: not found")
var ErrTooManyRequests = errors.New("http: too many requests")
var ErrMalformedRequest = errors.New("http: malformed request")

func NewErrUnprocessableEntity(code string, message string, data any) ErrUnprocessableEntity {
	return ErrUnprocessableEntity{
		ErrorCode: code,
		Message:   message,
		Data:      data,
	}
}

type ErrUnprocessableEntity struct {
	ErrorCode string
	Message   string
	Data      any
}

func (e ErrUnprocessableEntity) Error() string {
	additional := ""
	if e.ErrorCode != "" {
		additional = e.ErrorCode
	} else if e.Message != "" {
		additional = e.Message
	}

	if additional != "" {
		return fmt.Sprintf("http: unprocessable entity (%s)", additional)
	}
	return "http: unprocessable entity"
}

// [5xx] Server Errors

var ErrServiceUnavailable = errors.New("http: service unavailable")
