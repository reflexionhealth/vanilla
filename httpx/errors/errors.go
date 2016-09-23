package errors

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func Unauthorized(reason string, userMessage string) *Error {
	return &Error{
		HTTPStatus:  http.StatusUnauthorized,
		Meta:        Metadata{Reason: reason},
		UserMessage: userMessage,
	}
}

func Forbidden(reason string, userMessage string) *Error {
	return &Error{
		HTTPStatus:  http.StatusForbidden,
		Meta:        Metadata{Reason: reason},
		UserMessage: userMessage,
	}
}

func InternalError(err error) *Error {
	return &Error{
		HTTPStatus: http.StatusInternalServerError,
		Meta:       Metadata{Error: err},
	}
}

func Unavailable(debugMessage string) *Error {
	return &Error{
		HTTPStatus:   http.StatusServiceUnavailable,
		DebugMessage: debugMessage,
	}
}

func BadRequest(debugMessage string) *Error {
	return &Error{
		HTTPStatus:   http.StatusBadRequest,
		DebugMessage: debugMessage,
	}
}

func InvalidRequest(debugMessage string) *Error {
	return &Error{
		HTTPStatus:   http.StatusUnprocessableEntity,
		DebugMessage: debugMessage,
	}
}

func NotFound(debugMessage string) *Error {
	return &Error{
		HTTPStatus:   http.StatusNotFound,
		DebugMessage: debugMessage,
	}
}

type Error struct {
	HTTPStatus   int
	UserMessage  string
	DebugMessage string
	RequestID    string
	MoreInfo     url.URL

	// Meta stores additional data for internal use by the application
	Meta Metadata `json:"-"`
}

type Metadata struct {
	Reason string
	Error  error
	Trace  []string
}

func (err *Error) Error() string {
	return http.StatusText(err.HTTPStatus) + " - " + err.DebugMessage
}

type jsonError struct {
	UserMessage  string `json:"user_message"`
	DebugMessage string `json:"debug_message,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	MoreInfo     string `json:"more_info,omitempty"`
}

func (err *Error) MarshalJSON() ([]byte, error) {
	userMsg := err.UserMessage
	if len(userMsg) == 0 {
		userMsg = http.StatusText(err.HTTPStatus)
	}

	return json.Marshal(jsonError{
		UserMessage:  err.UserMessage,
		DebugMessage: err.DebugMessage,
		RequestID:    err.RequestID,
		MoreInfo:     err.MoreInfo.String(),
	})
}
