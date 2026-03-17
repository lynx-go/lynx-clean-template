package errors

import (
	"fmt"
	"net/http"
)

func New(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func Cause(message string) *APIError {
	return New(http.StatusInternalServerError, message)
}

func Wrap(err error, message string) *APIError {
	return &APIError{
		Code:    http.StatusInternalServerError,
		Message: message,
		Details: []string{
			err.Error(),
		},
		err: err,
	}
}

type APIError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
	err     error
}

func (e *APIError) Wrap(err error) *APIError {
	e.err = err
	if e.Details == nil {
		e.Details = []string{}
	}
	e.Details = append(e.Details, e.err.Error())
	return e
}

func (e *APIError) Cause() error {
	return e.err
}

type ErrorItem struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Message)
}

var _ error = new(APIError)

type StatusError struct {
	*APIError
	Status int `json:"code"`
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("[status=%d] %s", e.Status, e.APIError.Error())
}

func NewStatusError(status int, message string) *StatusError {
	return &StatusError{
		Status: status,
		APIError: &APIError{
			Code:    status,
			Message: message,
		},
	}
}
