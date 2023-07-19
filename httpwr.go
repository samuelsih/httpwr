package httpwr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	CreatedMsg             = "Created"
	OKMsg                  = "OK"
	InternalServerErrorMsg = "Internal Server Error"
	BadRequestMsg          = "Bad Request"
)

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
)

// M is a map type with key string and value any.
type M map[string]any

// Error is a HTTP error with an underlying error and a status code.
type Error struct {
	Status int   `json:"status"`
	Err    error `json:"error"`
}

// Error() implements the error interface.
func (e Error) Error() string {
	return e.Err.Error()
}

// Is conforms with errors.Is.
func (e Error) Is(err error) bool {
	switch err.(type) {
	case Error:
		return true
	default:
		return errors.Is(e.Err, err)
	}
}

// Wrap a given error with the given status.
// Returns nil if the given error is nil.
func Wrap(status int, err error) error {
	if err == nil {
		return nil
	}

	return Error{
		Err:    err,
		Status: status,
	}
}

// Errorf creates a new error and wraps it with the given status
func Errorf(status int, format string, args ...any) error {
	return Wrap(status, fmt.Errorf(format, args...))
}

// Handler is like http.Handler, but the ServeHTTP method can also return
// an error.
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc is just like http.HandlerFunc.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// ErrorHandler handles an error.
type ErrorHandler func(w http.ResponseWriter, status int, err error)

// DefaultErrorHandler is the default error handler.
// It converts the error to JSON and prints writes it to the response.
func DefaultErrorHandler(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(errorResponse{
		Status: status,
		Err:    err.Error(),
	})

}

// OK converts the status and message to JSON and sends it to user.
// Also, it will write the header based on the status.
func OK(w http.ResponseWriter, status int, msg string) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	type r struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
	}

	_ = json.NewEncoder(w).Encode(r{
		Status: status,
		Msg:    msg,
	})

	return nil
}

// OK converts the status, message and custom data you want to JSON.
// Also, it will write the header based on the status.
func OKWithData(w http.ResponseWriter, status int, msg string, data M) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	type r struct {
		Status int    `json:"status"`
		Msg    string `json:"msg"`
		Data   M      `json:"data"`
	}

	_ = json.NewEncoder(w).Encode(r{
		Status: status,
		Msg:    msg,
		Data:   data,
	})

	return nil
}

// NewWithHandler() wraps a given http.Handler and returns a http.Handler.
// You can also customize how the error is handled.
func NewWithHandler(next Handler, eh ErrorHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := next.ServeHTTP(w, r)
		if err == nil {
			return
		}

		var herr Error
		if errors.As(err, &herr) {
			eh(w, herr.Status, herr.Err)
		} else {
			eh(w, http.StatusInternalServerError, err)
		}
	})
}

// New wraps a given http.Handler and returns a http.Handler.
func New(next Handler) http.Handler {
	return NewWithHandler(next, DefaultErrorHandler)
}

// NewFWithHandler wraps a given http.HandlerFunc and return a http.Handler.
// You can also customize how the error is handled.
func NewFWithHandler(next HandlerFunc, eh ErrorHandler) http.Handler {
	return NewWithHandler(next, eh)
}

// NewF wraps a given http.HandlerFunc and return a http.Handler.
func NewF(next HandlerFunc) http.Handler {
	return New(next)
}

type errorResponse struct {
	Status int    `json:"status"`
	Err    string `json:"error"`
}
