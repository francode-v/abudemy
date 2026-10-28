package api

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// createOrgAndOpportunity provisions an organization and an open opportunity
// owned by the given institutional user, returning their identifiers.
func (e *integrationEnv) createOrgAndOpportunity(t *testing.T, instToken, name string) (string, string) {
	t.Helper()
	resp, body := e.request(t, http.MethodPost, "/api/v1/organizations", instToken, map[string]any{
		"name": name, "type": "community", "comuna": "Concepción", "region": "Biobío",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org: %d %s", resp.StatusCode, body)
	}
	orgID := decodeMap(t, body)["id"].(string)

	resp, body = e.request(t, http.MethodPost, "/api/v1/opportunities", instToken, map[string]any{
		"organization_id": orgID, "title": "Cuidador de jardín", "type": "gig",
		"status": "open", "comuna": "Concepción", "region": "Biobío",
		"contact_email": "empleo@test.dev",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create opportunity: %d %s", resp.StatusCode, body)
	}
	return orgID, decodeMap(t, body)["id"].(string)
}

func TestIntegrationPreferences(t *testing.T) {
	env := newIntegrationEnv(t)
	token, _ := env.registerAndLogin(t, "prefs@test.dev", "mobile_user")

	// Defaults are returned without persisting anything.
	resp, body := env.request(t, http.MethodGet, "/api/v1/users/me/preferences", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get preferences: %d %s", resp.StatusCode, body)
	}
	if got := decodeMap(t, body)["interests"].([]any); len(got) != 0 {
		t.Fatalf("expected empty interests, got %s", body)
	}

	// Over-long interests are rejected.
	if resp, _ := env.request(t, http.MethodPut, "/api/v1/users/me/preferences", token, map[string]any{
		"interests": []string{strings.Repeat("a", 61)},
	}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("over-long interest should be 400, got %d", resp.StatusCode)
	}

	// Interests are trimmed and de-duplicated case-insensitively.
	resp, body = env.request(t, http.MethodPut, "/api/v1/users/me/preferences", token, map[string]any{
		"interests":     []string{" Salud ", "salud", "Cocina"},
		"accessibility": map[string]any{"large_text": true, "high_contrast": false},
		"notifications": map[string]any{"activity_reminders": true},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("put preferences: %d %s", resp.StatusCode, body)
	}
	prefs := decodeMap(t, body)
	interests := prefs["interests"].([]any)
	if len(interests) != 2 || interests[0] != "Salud" || interests[1] != "Cocina" {
		t.Fatalf("interests not normalized: %s", body)
	}
	if prefs["accessibility"].(map[string]any)["large_text"] != true {
		t.Fatalf("accessibility not stored: %s", body)
	}
	if prefs["notifications"].(map[string]any)["activity_reminders"] != true {
		t.Fatalf("notification prefs not stored: %s", body)
	}

	// The values persist across reads.
	resp, body = env.request(t, http.MethodGet, "/api/v1/users/me/preferences", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("re-get preferences: %d %s", resp.StatusCode, body)
	}
	if len(decodeMap(t, body)["interests"].([]any)) != 2 {
		t.Fatalf("preferences did not persist: %s", body)
	}
}

func TestIntegrationEmailChange(t *testing.T) {
	env := newIntegrationEnv(t)
	token, _ := env.registerAndLogin(t, "email-one@test.dev", "mobile_user")
	env.registerAndLogin(t, "email-two@test.dev", "mobile_user")

	// Wrong password is rejected.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/users/me/email", token, map[string]any{
		"new_email": "brand-new@test.dev", "password": "wrong",
	}); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password should be 401, got %d", resp.StatusCode)
	}

	// An email already in use is a conflict.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/users/me/email", token, map[string]any{
		"new_email": "email-two@test.dev", "password": "Passw0rd!",
	}); resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate email should be 409, got %d", resp.StatusCode)
	}

	// A valid change succeeds and the new email can sign in.
	resp, body := env.request(t, http.MethodPost, "/api/v1/users/me/email", token, map[string]any{
		"new_email": "Brand-New@test.dev", "password": "Passw0rd!",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("change email: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["email"] != "brand-new@test.dev" {
		t.Fatalf("email not normalized/stored: %s", body)
	}
	resp, _ = env.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": "brand-new@test.dev", "password": "Passw0rd!",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with new email failed: %d", resp.StatusCode)
	}
}

func TestIntegrationApplicationsAndReports(t *testing.T) {
	env := newIntegrationEnv(t)
	mobileToken, _ := env.registerAndLogin(t, "applier@test.dev", "mobile_user")
	instToken, _ := env.registerAndLogin(t, "employer@test.dev", "institutional_client")
	_, oppID := env.createOrgAndOpportunity(t, instToken, "Bolsa Local")

	// Apply.
	resp, body := env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppID+"/applications", mobileToken, map[string]any{
		"message": "Tengo experiencia",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("apply: %d %s", resp.StatusCode, body)
	}
	appID := decodeMap(t, body)["id"].(string)

	// Duplicate application returns the friendly conflict message (regression
	// for the errors.Is fix).
	resp, body = env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppID+"/applications", mobileToken, map[string]any{
		"message": "otra vez",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate apply should be 409, got %d", resp.StatusCode)
	}
	msg := decodeMap(t, body)["error"].(map[string]any)["message"].(string)
	if !strings.Contains(msg, "already applied") {
		t.Fatalf("duplicate apply should explain the conflict, got %q", msg)
	}

	// The user can list their own applications with the opportunity summary.
	resp, body = env.request(t, http.MethodGet, "/api/v1/users/me/applications", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("my applications: %d %s", resp.StatusCode, body)
	}
	items := decodeMap(t, body)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 application, got %s", body)
	}
	first := items[0].(map[string]any)
	if first["opportunity_title"] != "Cuidador de jardín" || first["status"] != "applied" {
		t.Fatalf("application summary missing: %s", body)
	}

	// Accepting the application notifies the applicant.
	if resp, _ := env.request(t, http.MethodPatch, "/api/v1/opportunities/"+oppID+"/applications/"+appID, instToken, map[string]any{
		"status": "accepted",
	}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("accept application: %d", resp.StatusCode)
	}
	resp, body = env.request(t, http.MethodGet, "/api/v1/notifications/unread-count", mobileToken, nil)
	if decodeMap(t, body)["unread"].(float64) != 1 {
		t.Fatalf("expected 1 unread notification after acceptance, got %s", body)
	}

	// Reports are recorded once per user and opportunity.
	resp, body = env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppID+"/reports", mobileToken, map[string]any{
		"reason": "sospechoso", "details": "pide dinero por adelantado",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("report: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["status"] != "open" {
		t.Fatalf("report should be open: %s", body)
	}
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppID+"/reports", mobileToken, map[string]any{
		"reason": "otra vez",
	}); resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate report should be 409, got %d", resp.StatusCode)
	}
}

// An application may only be mutated through the opportunity it belongs to; a
// caller must not be able to reach another organization's application by
// addressing it through an opportunity of their own.
func TestIntegrationApplicationStatusIsOpportunityScoped(t *testing.T) {
	env := newIntegrationEnv(t)
	applicant, _ := env.registerAndLogin(t, "scope-applier@test.dev", "mobile_user")
	employerA, _ := env.registerAndLogin(t, "scope-employer-a@test.dev", "institutional_client")
	employerB, _ := env.registerAndLogin(t, "scope-employer-b@test.dev", "institutional_client")
	_, oppA := env.createOrgAndOpportunity(t, employerA, "Bolsa A")
	_, oppB := env.createOrgAndOpportunity(t, employerB, "Bolsa B")

	// The applicant applies to both opportunities.
	_, body := env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppA+"/applications", applicant, map[string]any{"message": "a"})
	appA := decodeMap(t, body)["id"].(string)
	_, body = env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppB+"/applications", applicant, map[string]any{"message": "b"})
	appB := decodeMap(t, body)["id"].(string)

	// Employer A must not be able to mutate employer B's application through
	// employer A's own opportunity.
	if resp, _ := env.request(t, http.MethodPatch,
		"/api/v1/opportunities/"+oppA+"/applications/"+appB, employerA,
		map[string]any{"status": "accepted"}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-opportunity application update should be 403, got %d", resp.StatusCode)
	}

	// Employer B's application is untouched.
	resp, body := env.request(t, http.MethodGet, "/api/v1/opportunities/"+oppB+"/applications", employerB, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list B applications: %d %s", resp.StatusCode, body)
	}
	items := decodeMap(t, body)["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["status"] != "applied" {
		t.Fatalf("foreign application was mutated: %s", body)
	}

	// Employer A can still manage their own application.
	if resp, _ := env.request(t, http.MethodPatch,
		"/api/v1/opportunities/"+oppA+"/applications/"+appA, employerA,
		map[string]any{"status": "accepted"}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("own application update should be 204, got %d", resp.StatusCode)
	}
}

func TestIntegrationDevicesAndNotifications(t *testing.T) {
	env := newIntegrationEnv(t)
	mobileToken, _ := env.registerAndLogin(t, "notif@test.dev", "mobile_user")

	// Register a device.
	resp, body := env.request(t, http.MethodPost, "/api/v1/devices", mobileToken, map[string]any{
		"platform": "android", "token": "tok-123", "device_name": "Pixel",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register device: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["platform"] != "android" {
		t.Fatalf("unexpected device: %s", body)
	}

	// Unknown platforms are rejected.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/devices", mobileToken, map[string]any{
		"platform": "windows", "token": "tok-x",
	}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid platform should be 400, got %d", resp.StatusCode)
	}

	// Re-registering the same token revives it instead of duplicating.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/devices", mobileToken, map[string]any{
		"platform": "android", "token": "tok-123",
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("re-register device failed")
	}
	resp, body = env.request(t, http.MethodGet, "/api/v1/devices", mobileToken, nil)
	if len(decodeMap(t, body)["items"].([]any)) != 1 {
		t.Fatalf("expected exactly 1 device, got %s", body)
	}

	// No notifications yet.
	resp, body = env.request(t, http.MethodGet, "/api/v1/notifications/unread-count", mobileToken, nil)
	if decodeMap(t, body)["unread"].(float64) != 0 {
		t.Fatalf("expected 0 unread, got %s", body)
	}

	// Trigger a notification through an application acceptance.
	instToken, _ := env.registerAndLogin(t, "notif-employer@test.dev", "institutional_client")
	_, oppID := env.createOrgAndOpportunity(t, instToken, "Notif Org")
	resp, body = env.request(t, http.MethodPost, "/api/v1/opportunities/"+oppID+"/applications", mobileToken, map[string]any{
		"message": "hola",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("apply: %d %s", resp.StatusCode, body)
	}
	appID := decodeMap(t, body)["id"].(string)
	if resp, _ := env.request(t, http.MethodPatch, "/api/v1/opportunities/"+oppID+"/applications/"+appID, instToken, map[string]any{
		"status": "accepted",
	}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("accept: %d", resp.StatusCode)
	}

	// The feed contains the notification, honestly marked as not pushed.
	resp, body = env.request(t, http.MethodGet, "/api/v1/notifications", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list notifications: %d %s", resp.StatusCode, body)
	}
	items := decodeMap(t, body)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 notification, got %s", body)
	}
	n := items[0].(map[string]any)
	if n["kind"] != "application_status" {
		t.Fatalf("unexpected notification kind: %s", body)
	}
	if n["push_status"] != "skipped" || n["push_provider"] != "none" {
		t.Fatalf("push must be honestly reported as skipped, got %s", body)
	}
	notifID := n["id"].(string)

	// Mark one as read.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/notifications/"+notifID+"/read", mobileToken, nil); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mark read failed")
	}
	resp, body = env.request(t, http.MethodGet, "/api/v1/notifications/unread-count", mobileToken, nil)
	if decodeMap(t, body)["unread"].(float64) != 0 {
		t.Fatalf("expected 0 unread after read, got %s", body)
	}

	// Revoking the device works and is idempotently rejected afterwards.
	devices := decodeMap(t, mustBody(t, env, http.MethodGet, "/api/v1/devices", mobileToken))["items"].([]any)
	deviceID := devices[0].(map[string]any)["id"].(string)
	if resp, _ := env.request(t, http.MethodDelete, "/api/v1/devices/"+deviceID, mobileToken, nil); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete device failed")
	}
	if resp, _ := env.request(t, http.MethodDelete, "/api/v1/devices/"+deviceID, mobileToken, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("deleting a revoked device should be 404, got %d", resp.StatusCode)
	}
}

func mustBody(t *testing.T, env *integrationEnv, method, path, token string) []byte {
	t.Helper()
	resp, body := env.request(t, method, path, token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s %s: %d %s", method, path, resp.StatusCode, body)
	}
	return body
}

func TestIntegrationActivityReminders(t *testing.T) {
	env := newIntegrationEnv(t)
	mobileToken, _ := env.registerAndLogin(t, "remind@test.dev", "mobile_user")
	instToken, _ := env.registerAndLogin(t, "remind-org@test.dev", "institutional_client")

	resp, body := env.request(t, http.MethodPost, "/api/v1/organizations", instToken, map[string]any{
		"name": "Centro Comunitario", "type": "community",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org: %d %s", resp.StatusCode, body)
	}
	orgID := decodeMap(t, body)["id"].(string)
	resp, body = env.request(t, http.MethodPost, "/api/v1/activities", instToken, map[string]any{
		"organization_id": orgID, "title": "Yoga suave", "type": "workshop", "modality": "in_person",
		"status": "published", "start_at": time.Now().Add(72 * time.Hour).Format(time.RFC3339),
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create activity: %d %s", resp.StatusCode, body)
	}
	activityID := decodeMap(t, body)["id"].(string)

	// Default reminder is accepted.
	resp, body = env.request(t, http.MethodPost, "/api/v1/activities/"+activityID+"/reminders", mobileToken, map[string]any{})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("set reminder: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["status"] != "pending" {
		t.Fatalf("reminder should be pending: %s", body)
	}

	// A reminder after the activity starts is rejected.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/activities/"+activityID+"/reminders", mobileToken, map[string]any{
		"remind_at": time.Now().Add(1000 * time.Hour).Format(time.RFC3339),
	}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("late reminder should be 400, got %d", resp.StatusCode)
	}

	// The reminder is listed with its activity title.
	resp, body = env.request(t, http.MethodGet, "/api/v1/users/me/reminders", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list reminders: %d %s", resp.StatusCode, body)
	}
	items := decodeMap(t, body)["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["activity_title"] != "Yoga suave" {
		t.Fatalf("unexpected reminders: %s", body)
	}

	// Cancelling twice is idempotent for the first call and 404 for the second.
	if resp, _ := env.request(t, http.MethodDelete, "/api/v1/activities/"+activityID+"/reminders/me", mobileToken, nil); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("cancel reminder failed")
	}
	if resp, _ := env.request(t, http.MethodDelete, "/api/v1/activities/"+activityID+"/reminders/me", mobileToken, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second cancel should be 404, got %d", resp.StatusCode)
	}
}

func TestIntegrationAssistant(t *testing.T) {
	env := newIntegrationEnv(t)
	token, _ := env.registerAndLogin(t, "assist@test.dev", "mobile_user")

	resp, body := env.request(t, http.MethodGet, "/api/v1/assistant/topics", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("topics: %d %s", resp.StatusCode, body)
	}
	if len(decodeMap(t, body)["items"].([]any)) == 0 {
		t.Fatalf("knowledge base must not be empty")
	}

	resp, body = env.request(t, http.MethodGet, "/api/v1/assistant/answer?q="+url.QueryEscape("cómo pedir ayuda"), token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("answer: %d %s", resp.StatusCode, body)
	}
	answer := decodeMap(t, body)
	if answer["source"] != "knowledge_base" || answer["ai"] != false {
		t.Fatalf("answers must be honestly labelled: %s", body)
	}
	if !strings.Contains(answer["answer"].(string), "Asistencia") {
		t.Fatalf("unexpected answer: %s", body)
	}

	// An empty question is a validation error.
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/assistant/answer?q=", token, nil); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty question should be 400, got %d", resp.StatusCode)
	}

	// Unknown questions fall back without inventing an answer.
	resp, body = env.request(t, http.MethodGet, "/api/v1/assistant/answer?q="+url.QueryEscape("zzzqqq"), token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fallback: %d %s", resp.StatusCode, body)
	}
	if _, hasTopic := decodeMap(t, body)["topic"]; hasTopic {
		t.Fatalf("fallback should not claim a topic: %s", body)
	}
}

func TestIntegrationGeoUnconfigured(t *testing.T) {
	env := newIntegrationEnv(t)
	token, _ := env.registerAndLogin(t, "geo@test.dev", "mobile_user")

	// With no provider, geocoding is honestly reported as not implemented.
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/geo/search?q="+url.QueryEscape("Concepción"), token, nil); resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("geocoding search should be 501 when unconfigured, got %d", resp.StatusCode)
	}
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/geo/reverse?lat=-33.4&lng=-70.6", token, nil); resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("geocoding reverse should be 501 when unconfigured, got %d", resp.StatusCode)
	}
	// Missing coordinates are a validation error even before the provider check.
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/geo/reverse", token, nil); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing lat/lng should be 400, got %d", resp.StatusCode)
	}
}

func TestIntegrationPaymentsConfig(t *testing.T) {
	env := newIntegrationEnv(t)
	token, _ := env.registerAndLogin(t, "pay@test.dev", "mobile_user")

	resp, body := env.request(t, http.MethodGet, "/api/v1/payments/config", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("payments config: %d %s", resp.StatusCode, body)
	}
	cfg := decodeMap(t, body)
	if cfg["provider"] != "none" || cfg["available"] != false {
		t.Fatalf("payments must be honestly unavailable: %s", body)
	}
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/payments/intents", token, map[string]any{}); resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("payment intents should be 501, got %d", resp.StatusCode)
	}
}
