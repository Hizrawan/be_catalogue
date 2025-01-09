package responses

import (
	"encoding/json"
	"net/http"
	"net/url"

	httperr "be20250107/internal/errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ErrorData struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func commonErrorHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func InternalServerError(w http.ResponseWriter, d any) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
		Details any `json:"details,omitempty"`
	}{
		ErrorData: ErrorData{
			ErrorCode: "internal_server_error",
			Message:   "Server current cannot process this request",
		},
		Details: d,
	})
}

func ServiceUnavailable(w http.ResponseWriter) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusServiceUnavailable)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
	}{
		ErrorData: ErrorData{
			ErrorCode: "service_unavailable",
			Message: "The service you want to access is currently unavailable. " +
				"Please try again later.",
		},
	})
}

func Unauthenticated(w http.ResponseWriter) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
	}{
		ErrorData: ErrorData{
			ErrorCode: "unauthenticated",
			Message: "Cannot authenticate with provided credentials. " +
				"Please check to make sure the credentials used are correct.",
		},
	})
}

func Forbidden(w http.ResponseWriter) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
	}{
		ErrorData: ErrorData{
			ErrorCode: "forbidden",
			Message:   "you don't have authority to access this resource or perform this action",
		},
	})
}

func NotFound(w http.ResponseWriter, u *url.URL) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusNotFound)

	path := ""
	if u != nil {
		path = u.String()
	}
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
		Path string `json:"path,omitempty"`
	}{
		ErrorData: ErrorData{
			ErrorCode: "not_found",
			Message:   "Requested resource cannot be found",
		},
		Path: path,
	})
}

func MalformedRequest(w http.ResponseWriter) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
	}{
		ErrorData: ErrorData{
			ErrorCode: "malformed_request",
			Message: "Cannot decode malformed request. " +
				"Please make sure the request is correctly formatted with all required fields.",
		},
	})
}

func ValidationError(w http.ResponseWriter, err validation.Errors) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
		Fields validation.Errors `json:"fields"`
	}{
		ErrorData: ErrorData{
			ErrorCode: "validation_error",
			Message:   "Some fields in the request failed validation.",
		},
		Fields: err,
	})
}

func UnprocessableEntity(w http.ResponseWriter, err httperr.ErrUnprocessableEntity) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
		Data any `json:"data,omitempty"`
	}{
		ErrorData: ErrorData{
			ErrorCode: err.ErrorCode,
			Message:   err.Message,
		},
		Data: err.Data,
	})
}

func TooManyRequests(w http.ResponseWriter) {
	commonErrorHeader(w)
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(struct {
		ErrorData
	}{
		ErrorData: ErrorData{
			ErrorCode: "too_many_request",
			Message:   "Cannot request resource due to limit rate",
		},
	})
}
