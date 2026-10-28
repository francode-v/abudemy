package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestUserMarshalJSONRendersBirthDateAsCalendarDate(t *testing.T) {
	birth := time.Date(1948, 5, 12, 0, 0, 0, 0, time.UTC)
	user := User{
		ID:        "u1",
		Email:     "maria@ejemplo.cl",
		FullName:  "María",
		BirthDate: &birth,
		Role:      RoleMobileUser,
		Status:    UserStatusActive,
	}

	raw, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := decoded["birth_date"]; got != "1948-05-12" {
		t.Fatalf("birth_date must be the calendar date 1948-05-12, got %v (%s)", got, raw)
	}
}

func TestUserMarshalJSONOmitsMissingBirthDate(t *testing.T) {
	user := User{ID: "u1", Email: "sin@ejemplo.cl", FullName: "Sin fecha"}

	raw, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "birth_date") {
		t.Fatalf("birth_date must be omitted when empty: %s", raw)
	}
}
