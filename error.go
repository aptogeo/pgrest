package pgrest

import "fmt"

// Error struct
type Error struct {
	Message string
	Cause   error
	Code    int
}

// NewErrorBadRequest constructs Error with bad request code
func NewErrorBadRequest(message string) *Error {
	return &Error{Message: message, Code: 400}
}

// NewErrorForbbiden constructs Error with forbidden code
func NewErrorForbbiden(message string) *Error {
	return &Error{Message: message, Code: 403}
}

// NewErrorFromCause constructs Error from cause error
func NewErrorFromCause(restQuery *RestQuery, cause error) *Error {
	errStr := cause.Error()
	if errStr == "pg: no rows in result set" || errStr == "pg: multiple rows in result set" {
		return &Error{Message: fmt.Sprintf("resource '%v' with key '%v' not found", restQuery.Resource, restQuery.Key), Code: 404, Cause: cause}
	}
	return &Error{Cause: cause}
}

// Error implements the error interface
func (e Error) Error() string {
	msg := e.Message
	if e.Cause != nil {
		msg += " (" + e.Cause.Error() + ")"
	}
	return msg
}

// StatusCode returns code
func (e Error) StatusCode() int {
	if e.Code != 0 {
		return e.Code
	}
	if e.Cause != nil {
		if causeError, ok := e.Cause.(*Error); ok {
			return causeError.StatusCode()
		}
	}
	return 500
}
