// Package validate provides small, dependency-free request validation helpers
// that accumulate field errors for a consistent 400 response.
package validate

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/abudemy/backend/internal/apperr"
)

// Errors collects field errors during a validation pass.
type Errors struct {
	fields []apperr.FieldError
}

// Add records a field error.
func (e *Errors) Add(field, message string) {
	e.fields = append(e.fields, apperr.FieldError{Field: field, Message: message})
}

// Required records an error when value is empty.
func (e *Errors) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		e.Add(field, "is required")
	}
}

// MaxLen records an error when the rune length of value exceeds n.
func (e *Errors) MaxLen(field, value string, n int) {
	if utf8.RuneCountInString(value) > n {
		e.Add(field, "must be at most "+itoa(n)+" characters")
	}
}

// MinLen records an error when the rune length of value is below n.
func (e *Errors) MinLen(field, value string, n int) {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < n {
		e.Add(field, "must be at least "+itoa(n)+" characters")
	}
}

// Email records an error when value is not a syntactically valid email.
func (e *Errors) Email(field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if _, err := mail.ParseAddress(value); err != nil || !strings.Contains(value, "@") {
		e.Add(field, "must be a valid email address")
	}
}

// OneOf records an error when value is not in allowed (empty values pass).
func (e *Errors) OneOf(field, value string, allowed ...string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	e.Add(field, "must be one of: "+strings.Join(allowed, ", "))
}

// Range records an error when value is outside [min,max].
func (e *Errors) Range(field string, value, min, max int) {
	if value < min || value > max {
		e.Add(field, "must be between "+itoa(min)+" and "+itoa(max))
	}
}

// Latitude records an error when the coordinate is out of range.
func (e *Errors) Latitude(field string, v *float64) {
	if v != nil && (*v < -90 || *v > 90) {
		e.Add(field, "must be between -90 and 90")
	}
}

// Longitude records an error when the coordinate is out of range.
func (e *Errors) Longitude(field string, v *float64) {
	if v != nil && (*v < -180 || *v > 180) {
		e.Add(field, "must be between -180 and 180")
	}
}

// HasErrors reports whether any field errors were collected.
func (e *Errors) HasErrors() bool { return len(e.fields) > 0 }

// Err returns an *apperr.Error when validation failed, otherwise nil.
func (e *Errors) Err() error {
	if !e.HasErrors() {
		return nil
	}
	return apperr.Validation(e.fields...)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
