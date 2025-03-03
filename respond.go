package goweb

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Respond to an HTTP request with a json payload.
func Respond(w http.ResponseWriter, r *http.Request, data any) error {
	_ = r // Request can be used in the future to check the Accept header

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)

	if err != nil {
		slog.Error("Failed to send JSON response", "error", err, "data", data)
	}

	return err
}

// RespondError sends a JSON error response to the client.
//
// If the error fulfills [APIError], it will be used to generate the response.
// Otherwise, a generic error response will be sent. If the error code is
// [ErrGeneric], the error will be logged.
func RespondError(w http.ResponseWriter, r *http.Request, e error) error {
	_ = r // Request can be used in the future to check the Accept header

	var apiError APIError
	if ok := errors.As(e, &apiError); !ok {
		apiError = ErrGeneric.Wrap(e)
	}

	var statusCode int
	var response struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  any    `json:"detail"`
	}

	response.Code = apiError.ErrorCode()
	response.Message = apiError.Error()
	response.Detail = apiError.ErrorDetail()
	statusCode = apiError.StatusCode()

	if me, ok := e.(ErrorMasker); ok && me.MaskError() {
		response.Message = ""
		response.Detail = nil
		slog.Error("Error response", "error", apiError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)

	if err != nil {
		slog.Error("Failed to send error response as JSON", "error", err, "data", response)
	}

	return err
}
