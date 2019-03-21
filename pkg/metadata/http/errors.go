/*
 * Copyright (c) Vijay Poliboyina 2019.
 */

package http

import "encoding/json"

// httpError sits well with the go-kit ServerErrorDecoder function.
type httpError struct {
	statusCode int
	message    string
	cause      string
}

func (h httpError) Error() string {
	return h.message
}

func (h httpError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Message string `json:"message"`
		Cause   string `json:"cause,omitempty"`
	}{
		Message: h.message,
		Cause:   h.cause,
	})
}

func newError(statusCode int) httpError {
	return httpError{statusCode: statusCode}
}

func (h httpError) WithMessage(message string) httpError {
	h.message = message
	return h
}

func (h httpError) WithCause(cause string) httpError {
	h.cause = cause
	return h
}

func (h httpError) StatusCode() int {
	return h.statusCode
}
