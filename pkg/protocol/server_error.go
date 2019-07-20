package protocol

import (
	"fmt"
	"register/pkg/protocol/grpc"
)

var (
	ErrSystem = NewServerError(grpc.ErrSys)
)

// ServerError code/message
type ServerError interface {
	error
	Code() int
	Message() string
}

type serverError struct {
	code       int
	message    string
	underlying error
}

// NewServerError create a new instance of ServerError
// with the specific status and message
func NewServerError(code int, message ...string) ServerError {
	msg := ""
	if len(message) > 0 { //&& len(message[0]) > 0 {
		msg = message[0]
	}
	return &serverError{code: code, message: msg}
}

func NewServerErrorWithUnderlying(underlying error, code int, message ...string) ServerError {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	return &serverError{code: code, message: msg, underlying: underlying}
}

// Code returns the status code of ServerError
func (e *serverError) Code() int {
	return e.code
}

// Message returns the message of ServerError
func (e *serverError) Message() string {
	return e.message
}

func (e *serverError) Error() string {
	return fmt.Sprintf("status: %d, message: %s, underlying: %v", e.code, e.message, e.underlying)
}

func ToServerError(err error) ServerError {
	if err == nil {
		return nil
	}

	if pe, ok := err.(ServerError); ok {
		return pe
	}

	return NewServerErrorWithUnderlying(err, grpc.ErrSys)
}

func FromSvrkitError(err error) ServerError {
	return ToServerError(err)
}
