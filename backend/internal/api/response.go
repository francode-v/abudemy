package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// errorEnvelope is the consistent error response body.
type errorEnvelope struct {
	Error *apperr.Error `json:"error"`
}

// writeJSON encodes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

// writeNoContent sends an empty 204 response.
func writeNoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// writeError renders an application error and logs internal failures.
func writeError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	appErr := apperr.As(err)
	if appErr == nil {
		return
	}
	if appErr.Code == apperr.CodeInternal {
		logger.ErrorContext(r.Context(), "request failed", "code", appErr.Code, "error", err)
	} else {
		logger.DebugContext(r.Context(), "request rejected", "code", appErr.Code, "message", appErr.Message)
	}
	writeJSON(w, appErr.Status(), errorEnvelope{Error: appErr})
}

// decodeJSON reads a single JSON object from the request body with a size cap
// and strict field checking.
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return apperr.Validation(apperr.FieldError{Field: "body", Message: "a JSON body is required"})
	}
	defer r.Body.Close()

	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}
	// Ensure there is no trailing content.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return apperr.Validation(apperr.FieldError{Field: "body", Message: "must contain a single JSON object"})
	}
	return nil
}

func decodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &syntaxErr):
		return apperr.Validation(apperr.FieldError{Field: "body", Message: "contains malformed JSON"})
	case errors.As(err, &typeErr):
		return apperr.Validation(apperr.FieldError{Field: typeErr.Field, Message: "has an invalid type"})
	case errors.Is(err, io.EOF):
		return apperr.Validation(apperr.FieldError{Field: "body", Message: "is required"})
	case strings.Contains(err.Error(), "unknown field"):
		return apperr.Validation(apperr.FieldError{Field: "body", Message: "contains an unknown field"})
	default:
		return apperr.Validation(apperr.FieldError{Field: "body", Message: "is invalid"})
	}
}

// ---- query parameter helpers ----

func queryString(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

func queryInt(r *http.Request, key string) int {
	v := queryString(r, key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}

func queryFloatPtr(r *http.Request, key string) *float64 {
	v := queryString(r, key)
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func queryBoolPtr(r *http.Request, key string) *bool {
	v := queryString(r, key)
	if v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil
	}
	return &b
}

func queryTimePtr(r *http.Request, key string) *time.Time {
	v := queryString(r, key)
	if v == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return &t
	}
	return nil
}

func pageParams(r *http.Request) models.PageParams {
	return models.NewPageParams(queryInt(r, "page"), queryInt(r, "per_page"))
}
