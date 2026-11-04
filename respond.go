package goweb

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Respond to an HTTP request with a json payload.
func Respond(w http.ResponseWriter, r *http.Request, data any) error {
	w.Header().Set("Content-Type", "application/json")

	return writeJSON(requestContext(r), w, data)
}

// RespondError sends a JSON error response to the client.
//
// If the error fulfills [APIError], it will be used to generate the response.
// Otherwise, a generic error response will be sent. If the error is masked
// (see [ErrorMasker]), it will be logged with the request's context.
func RespondError(w http.ResponseWriter, r *http.Request, e error) error {
	ctx := requestContext(r)

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

	if me, ok := apiError.(ErrorMasker); ok && me.MaskError() {
		response.Message = ""
		response.Detail = nil
		slog.ErrorContext(ctx, "error response", slog.Any("error", apiError))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return writeJSON(ctx, w, response)
}

// writeJSON encodes data before writing it so that encoding errors (a bug)
// can be told apart from write errors (usually a client that went away).
func writeJSON(ctx context.Context, w http.ResponseWriter, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to encode JSON response", slog.Any("error", err))
		return err
	}

	if _, err := w.Write(append(body, '\n')); err != nil {
		slog.DebugContext(ctx, "failed to write JSON response", slog.Any("error", err))
		return err
	}

	return nil
}

func requestContext(r *http.Request) context.Context {
	if r == nil {
		return context.Background()
	}

	return r.Context()
}
