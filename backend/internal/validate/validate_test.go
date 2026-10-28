package validate

import "testing"

func TestErrorsAccumulate(t *testing.T) {
	var v Errors
	v.Required("name", "")
	v.Email("email", "not-an-email")
	v.MinLen("full_name", "A", 2)
	if !v.HasErrors() {
		t.Fatal("expected errors")
	}
	err := v.Err()
	if err == nil {
		t.Fatal("expected an error")
	}
	if len(v.fields) != 3 {
		t.Fatalf("expected 3 field errors, got %d", len(v.fields))
	}
}

func TestValidInputPasses(t *testing.T) {
	var v Errors
	v.Required("name", "Abudemy")
	v.Email("email", "user@example.com")
	v.MaxLen("name", "Abudemy", 10)
	v.OneOf("role", "admin", "admin", "user")
	v.Range("priority", 5, 1, 10)
	v.Latitude("lat", floatPtr(-33.45))
	v.Longitude("lng", floatPtr(-70.66))
	if err := v.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCoordinateBounds(t *testing.T) {
	var v Errors
	v.Latitude("lat", floatPtr(120))
	v.Longitude("lng", floatPtr(-200))
	if len(v.fields) != 2 {
		t.Fatalf("expected 2 coordinate errors, got %d", len(v.fields))
	}
}

func TestOneOfIgnoresEmpty(t *testing.T) {
	var v Errors
	v.OneOf("role", "", "admin", "user")
	if v.HasErrors() {
		t.Fatal("empty value should be allowed to fall back to a default")
	}
}

func floatPtr(f float64) *float64 { return &f }
