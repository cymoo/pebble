package errors

import (
	"encoding/json"
	"net/http"

	m "github.com/cymoo/mint"
)

func NotFound(message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return m.HTTPError{Code: 404, Err: "not_found", Message: msg}
}

func BadRequest(message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return m.HTTPError{Code: 400, Err: "bad_request", Message: msg}
}

func Unauthorized(message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return m.HTTPError{Code: 401, Err: "unauthorized", Message: msg}
}

func InternalError(message ...string) error {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return m.HTTPError{Code: 500, Err: "internal_error", Message: msg}
}

func SendJSONError(w http.ResponseWriter, code int, err string, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	errResp := m.HTTPError{
		Code:    code,
		Err:     err,
		Message: msg,
	}

	json.NewEncoder(w).Encode(errResp)
}
