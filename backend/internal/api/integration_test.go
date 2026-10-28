package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/database"
	"github.com/abudemy/backend/internal/logging"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/notify"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

// integrationEnv boots a full server against a real PostgreSQL database.
// It requires TEST_DATABASE_URL to point at a *dedicated test database* (its
// name must contain "test") because the suite truncates all tables.
type integrationEnv struct {
	base  string
	http  *http.Client
	ts    *httptest.Server
	pool  *pgxpool.Pool
	store *store.Store
}

func newIntegrationEnv(t *testing.T) *integrationEnv {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration tests")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dsn, 10)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	var dbName string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&dbName); err != nil {
		t.Fatalf("current_database: %v", err)
	}
	if !strings.Contains(strings.ToLower(dbName), "test") {
		t.Skipf("refusing to run destructive integration tests against database %q", dbName)
	}

	if err := database.Migrate(ctx, pool, database.Options{SeedDemo: true}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE users, sessions, password_reset_tokens, organizations,
		organization_members, service_categories, services, service_schedules, activities,
		activity_enrollments, skills, offerings, opportunities, opportunity_applications,
		trusted_contacts, assistance_requests, assistance_notifications,
		user_preferences, devices, notifications, activity_reminders, opportunity_reports
		RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	st := store.New(pool)
	logger := logging.New("error", false)
	tokens := auth.NewTokenManager("integration-test-secret-key-0123456789", "abudemy-test", 15*time.Minute)
	hasher := auth.NewPasswordHasher(4) // fast hashing for tests
	notifier := notify.New("none", "", "", logger)

	services := service.New(service.Deps{
		Store: st, Tokens: tokens, Hasher: hasher, Notifier: notifier, Logger: logger,
		RefreshTTL: time.Hour, PasswordResetTTL: time.Hour,
	})
	server := NewServer(Config{
		Services: services, Store: st, Tokens: tokens, Logger: logger,
		CORSOrigins: []string{"*"}, Pinger: pool,
	})
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(ts.Close)

	return &integrationEnv{base: ts.URL, http: ts.Client(), ts: ts, pool: pool, store: st}
}

func (e *integrationEnv) request(t *testing.T, method, path, token string, body any) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.base+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := e.http.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, data
}

func decodeMap(t *testing.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode json (%s): %v", truncate(data), err)
	}
	return m
}

func truncate(b []byte) string {
	if len(b) > 300 {
		return string(b[:300])
	}
	return string(b)
}

// registerAndLogin creates an account and returns its access token.
func (e *integrationEnv) registerAndLogin(t *testing.T, email, role string) (string, string) {
	t.Helper()
	resp, body := e.request(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"email": email, "password": "Passw0rd!", "full_name": "Test User", "role": role,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: status %d body %s", email, resp.StatusCode, body)
	}
	userID := decodeMap(t, body)["id"].(string)

	// Public institutional registrations start pending administrator approval.
	// Tests exercise the institutional portal, so approve them explicitly.
	if role == "institutional_client" {
		if err := e.store.SetUserInstitutionalApproval(context.Background(), userID, true); err != nil {
			t.Fatalf("approve %s: %v", email, err)
		}
	}

	resp, body = e.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": email, "password": "Passw0rd!",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s: status %d body %s", email, resp.StatusCode, body)
	}
	token := decodeMap(t, body)["access_token"].(string)
	return token, userID
}

func TestIntegrationHealth(t *testing.T) {
	env := newIntegrationEnv(t)
	resp, body := env.request(t, http.MethodGet, "/healthz", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", resp.StatusCode)
	}
	if decodeMap(t, body)["status"] != "ok" {
		t.Fatalf("unexpected health body: %s", body)
	}
	resp, _ = env.request(t, http.MethodGet, "/readyz", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ready status = %d", resp.StatusCode)
	}
}

func TestIntegrationFullFlow(t *testing.T) {
	env := newIntegrationEnv(t)

	// --- registration and login ---
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"email": "weak@test.dev", "password": "short", "full_name": "Weak",
	}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("weak password should be rejected, got %d", resp.StatusCode)
	}

	mobileToken, mobileID := env.registerAndLogin(t, "older@test.dev", "mobile_user")
	instToken, _ := env.registerAndLogin(t, "clinic@test.dev", "institutional_client")

	// --- profile ---
	resp, body := env.request(t, http.MethodGet, "/api/v1/users/me", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get me: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["id"] != mobileID {
		t.Fatalf("me returned wrong user: %s", body)
	}

	// Mobile user cannot list all users (admin only).
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/users", mobileToken, nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mobile user should not list users, got %d", resp.StatusCode)
	}

	// An approved institutional client can list registered people, but only
	// community (mobile_user) accounts, even if it asks for another role.
	resp, body = env.request(t, http.MethodGet, "/api/v1/users?role=admin", instToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("institutional client should list registered people, got %d %s", resp.StatusCode, body)
	}
	items, _ := decodeMap(t, body)["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("institutional client should see at least one mobile user: %s", body)
	}
	for _, raw := range items {
		if role, _ := raw.(map[string]any)["role"].(string); role != "mobile_user" {
			t.Fatalf("institutional client received a non-community account: %s", role)
		}
	}

	// --- organization ---
	resp, body = env.request(t, http.MethodPost, "/api/v1/organizations", instToken, map[string]any{
		"name": "Clínica del Barrio", "type": "health", "comuna": "Ñuñoa", "region": "Metropolitana",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create org: %d %s", resp.StatusCode, body)
	}
	orgID := decodeMap(t, body)["id"].(string)

	// Mobile user cannot create an organization.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/organizations", mobileToken, map[string]any{"name": "X", "type": "other"}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mobile user should not create org, got %d", resp.StatusCode)
	}

	// --- service catalogue ---
	resp, body = env.request(t, http.MethodPost, "/api/v1/services", instToken, map[string]any{
		"organization_id": orgID, "name": "Control de presión", "type": "establishment",
		"comuna": "Ñuñoa", "region": "Metropolitana", "latitude": -33.45, "longitude": -70.66,
		"schedule": []map[string]any{{"day_of_week": 1, "opens_at": "09:00", "closes_at": "13:00"}},
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create service: %d %s", resp.StatusCode, body)
	}
	serviceBody := decodeMap(t, body)
	if serviceBody["name"] != "Control de presión" {
		t.Fatalf("unexpected service: %s", body)
	}
	if serviceBody["organization_name"] != "Clínica del Barrio" {
		t.Fatalf("service should expose its organization: %s", body)
	}
	if len(serviceBody["schedule"].([]any)) != 1 {
		t.Fatalf("schedule not persisted: %s", body)
	}

	// Mobile user cannot create a platform service.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/services", mobileToken, map[string]any{"name": "X"}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mobile user should not create a service, got %d", resp.StatusCode)
	}

	// Search by comuna.
	resp, body = env.request(t, http.MethodGet, "/api/v1/services?comuna=%C3%91u%C3%B1oa", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list services: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("expected 1 service in comuna, got %s", body)
	}

	// Search is case-insensitive.
	resp, body = env.request(t, http.MethodGet, "/api/v1/services?search=CONTROL", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("search services: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("uppercase search should match, got %s", body)
	}

	// --- activity with capacity 1 ---
	resp, body = env.request(t, http.MethodPost, "/api/v1/activities", instToken, map[string]any{
		"organization_id": orgID, "title": "Taller de memoria", "type": "workshop",
		"modality": "in_person", "status": "published", "comuna": "Ñuñoa",
		"start_at": time.Now().Add(48 * time.Hour).Format(time.RFC3339), "capacity": 1,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create activity: %d %s", resp.StatusCode, body)
	}
	activityID := decodeMap(t, body)["id"].(string)

	// First enrollment succeeds.
	resp, body = env.request(t, http.MethodPost, "/api/v1/activities/"+activityID+"/enrollments", mobileToken, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("enroll: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["waitlisted"].(bool) {
		t.Fatalf("first enrollee should not be waitlisted")
	}

	// Duplicate enrollment is a conflict.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/activities/"+activityID+"/enrollments", mobileToken, nil); resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate enrollment should conflict, got %d", resp.StatusCode)
	}

	// A user who already attended cannot re-enroll: doing so would demote them
	// to enrolled/waitlist and lose the attendance record.
	if resp, body := env.request(t, http.MethodPatch, "/api/v1/activities/"+activityID+"/enrollments/"+mobileID, instToken, map[string]any{"status": "attended"}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("mark attended: %d %s", resp.StatusCode, body)
	}
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/activities/"+activityID+"/enrollments", mobileToken, nil); resp.StatusCode != http.StatusConflict {
		t.Fatalf("re-enrolling an attended user should conflict, got %d", resp.StatusCode)
	}
	resp, body = env.request(t, http.MethodGet, "/api/v1/users/me/enrollments", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("my enrollments after attended: %d %s", resp.StatusCode, body)
	}
	for _, item := range decodeMap(t, body)["items"].([]any) {
		e := item.(map[string]any)
		if e["activity_id"] == activityID && e["status"] != "attended" {
			t.Fatalf("attended status was overwritten: %s", body)
		}
	}

	// Second user is waitlisted because capacity is 1.
	otherToken, _ := env.registerAndLogin(t, "second@test.dev", "mobile_user")
	resp, body = env.request(t, http.MethodPost, "/api/v1/activities/"+activityID+"/enrollments", otherToken, nil)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("second enroll: %d %s", resp.StatusCode, body)
	}
	if !decodeMap(t, body)["waitlisted"].(bool) {
		t.Fatalf("second enrollee should be waitlisted")
	}

	// The mobile user can see their own enrollment.
	resp, body = env.request(t, http.MethodGet, "/api/v1/users/me/enrollments", mobileToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("my enrollments: %d %s", resp.StatusCode, body)
	}

	// --- assistance (no provider configured) ---
	resp, body = env.request(t, http.MethodPost, "/api/v1/assistance/contacts", mobileToken, map[string]any{
		"name": "Hija", "relationship": "family", "phone": "+56911111111",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create contact: %d %s", resp.StatusCode, body)
	}

	resp, body = env.request(t, http.MethodPost, "/api/v1/assistance/requests", mobileToken, map[string]any{
		"type": "medical", "description": "Necesito ayuda",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create assistance request: %d %s", resp.StatusCode, body)
	}
	reqBody := decodeMap(t, body)
	notifications := reqBody["notifications"].([]any)
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification attempt, got %d", len(notifications))
	}
	if notifications[0].(map[string]any)["status"] != "skipped" {
		t.Fatalf("with no provider the attempt must be 'skipped', got %s", body)
	}
	if reqBody["status"] == "notified" {
		t.Fatalf("request must not be marked notified when nothing was sent")
	}

	// The other user cannot read this request.
	requestID := reqBody["id"].(string)
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/assistance/requests/"+requestID, otherToken, nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-user assistance access should be forbidden, got %d", resp.StatusCode)
	}

	// --- analytics & reports ---
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/analytics/overview", mobileToken, nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("mobile user should not read analytics, got %d", resp.StatusCode)
	}
	resp, body = env.request(t, http.MethodGet, "/api/v1/analytics/overview?organization_id="+orgID, instToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("overview: %d %s", resp.StatusCode, body)
	}
	overview := decodeMap(t, body)
	if overview["services_total"].(float64) != 1 || overview["activities_total"].(float64) != 1 {
		t.Fatalf("unexpected analytics: %s", body)
	}

	resp, body = env.request(t, http.MethodGet, "/api/v1/analytics/reports/services?organization_id="+orgID, instToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export: %d %s", resp.StatusCode, body)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("expected CSV content type, got %q", ct)
	}
	if !strings.Contains(string(body), "Control de presión") {
		t.Fatalf("CSV should contain the service: %s", body)
	}

	// JSON preview powers the on-screen report table and includes the
	// organization that provides each service.
	resp, body = env.request(t, http.MethodGet,
		"/api/v1/analytics/reports/services?organization_id="+orgID+"&format=json", instToken, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("json preview: %d %s", resp.StatusCode, body)
	}
	preview := decodeMap(t, body)
	if preview["type"] != "services" {
		t.Fatalf("unexpected preview type: %s", body)
	}
	if preview["total"].(float64) != 1 || preview["truncated"] != false {
		t.Fatalf("unexpected preview totals: %s", body)
	}
	hasOrganization := false
	for _, header := range preview["headers"].([]any) {
		if header == "organization" {
			hasOrganization = true
		}
	}
	if !hasOrganization {
		t.Fatalf("preview headers should include the organization: %s", body)
	}
	if rows := preview["rows"].([]any); len(rows) != 1 {
		t.Fatalf("preview should include one row: %s", body)
	}

	// PDF is explicitly not implemented yet.
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/analytics/reports/services?format=pdf", instToken, nil); resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("pdf export should be 501, got %d", resp.StatusCode)
	}

	// --- token rotation and logout ---
	resp, body = env.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": "older@test.dev", "password": "Passw0rd!",
	})
	pair := decodeMap(t, body)
	refresh := pair["refresh_token"].(string)

	resp, body = env.request(t, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refresh_token": refresh})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("refresh: %d %s", resp.StatusCode, body)
	}
	refreshedAccess := decodeMap(t, body)["access_token"].(string)
	// The old refresh token is rotated and cannot be reused.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refresh_token": refresh}); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reused refresh token should be rejected, got %d", resp.StatusCode)
	}

	resp, _ = env.request(t, http.MethodPost, "/api/v1/auth/logout", refreshedAccess, nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: %d", resp.StatusCode)
	}
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/users/me", refreshedAccess, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("access token must stop working after logout, got %d", resp.StatusCode)
	}
}

func TestIntegrationPasswordReset(t *testing.T) {
	env := newIntegrationEnv(t)
	env.registerAndLogin(t, "reset@test.dev", "mobile_user")

	// Requesting a reset never discloses whether the account exists.
	resp, _ := env.request(t, http.MethodPost, "/api/v1/auth/password/forgot", "", map[string]any{"email": "nobody@test.dev"})
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("forgot for unknown email should be accepted, got %d", resp.StatusCode)
	}
	resp, _ = env.request(t, http.MethodPost, "/api/v1/auth/password/forgot", "", map[string]any{"email": "reset@test.dev"})
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("forgot: %d", resp.StatusCode)
	}

	// A reset token was persisted even though delivery depends on a provider.
	ctx := context.Background()
	var count int
	if err := env.pool.QueryRow(ctx, `SELECT count(*) FROM password_reset_tokens`).Scan(&count); err != nil {
		t.Fatalf("count tokens: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 stored reset token, got %d", count)
	}

	// An invalid token is rejected.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/auth/password/reset", "", map[string]any{
		"token": "not-a-real-token", "new_password": "NewPass123",
	}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid reset token should be rejected, got %d", resp.StatusCode)
	}
}

func TestIntegrationPasswordChange(t *testing.T) {
	env := newIntegrationEnv(t)
	token, _ := env.registerAndLogin(t, "change@test.dev", "mobile_user")

	// Wrong current password is rejected.
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/users/me/password", token, map[string]any{
		"current_password": "wrong", "new_password": "NewPass123",
	}); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password should be 401, got %d", resp.StatusCode)
	}

	if resp, _ := env.request(t, http.MethodPost, "/api/v1/users/me/password", token, map[string]any{
		"current_password": "Passw0rd!", "new_password": "NewPass123",
	}); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("password change failed")
	}

	// Old sessions are revoked and the new password works.
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/users/me", token, nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("sessions should be revoked after password change, got %d", resp.StatusCode)
	}
	resp, _ := env.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": "change@test.dev", "password": "NewPass123",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with new password failed: %d", resp.StatusCode)
	}
}

func TestIntegrationInstitutionalApproval(t *testing.T) {
	env := newIntegrationEnv(t)

	// Register an institutional account directly: registerAndLogin approves it,
	// and this test needs the pending state.
	resp, body := env.request(t, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"email": "pending@test.dev", "password": "Passw0rd!", "full_name": "Pending Clinic",
		"role": "institutional_client",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: %d %s", resp.StatusCode, body)
	}
	pendingID := decodeMap(t, body)["id"].(string)
	if decodeMap(t, body)["institutional_approved"] != false {
		t.Fatalf("new institutional account should start pending: %s", body)
	}

	resp, body = env.request(t, http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"email": "pending@test.dev", "password": "Passw0rd!",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: %d %s", resp.StatusCode, body)
	}
	pendingToken := decodeMap(t, body)["access_token"].(string)

	// Pending accounts may sign in but cannot use institutional features.
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/analytics/overview", pendingToken, nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("pending account should not read analytics, got %d", resp.StatusCode)
	}
	if resp, _ := env.request(t, http.MethodPost, "/api/v1/organizations", pendingToken, map[string]any{
		"name": "Pending Org", "type": "other",
	}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("pending account should not create organizations, got %d", resp.StatusCode)
	}

	// A regular account cannot approve institutional access.
	adminToken, adminID := env.registerAndLogin(t, "approver@test.dev", "mobile_user")
	if resp, _ := env.request(t, http.MethodPatch, "/api/v1/users/"+pendingID+"/institutional-approval",
		adminToken, map[string]any{"approved": true}); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("non-admin should not approve accounts, got %d", resp.StatusCode)
	}

	// Promote the account to admin; the middleware reloads the role per request.
	if err := env.store.UpdateUserRole(context.Background(), adminID, models.RoleAdmin); err != nil {
		t.Fatalf("promote admin: %v", err)
	}

	// The flag must be explicit: an empty payload can never revoke access.
	if resp, _ := env.request(t, http.MethodPatch, "/api/v1/users/"+pendingID+"/institutional-approval",
		adminToken, map[string]any{}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty approval payload should be rejected, got %d", resp.StatusCode)
	}
	if resp, _ := env.request(t, http.MethodPatch, "/api/v1/users/"+pendingID+"/institutional-approval",
		adminToken, nil); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing approval body should be rejected, got %d", resp.StatusCode)
	}

	// Only institutional client accounts can be approved.
	if resp, _ := env.request(t, http.MethodPatch, "/api/v1/users/"+adminID+"/institutional-approval",
		adminToken, map[string]any{"approved": true}); resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("non-institutional account should not be approved, got %d", resp.StatusCode)
	}

	resp, body = env.request(t, http.MethodPatch, "/api/v1/users/"+pendingID+"/institutional-approval",
		adminToken, map[string]any{"approved": true})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("approve: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["institutional_approved"] != true {
		t.Fatalf("approval not persisted: %s", body)
	}

	// The pending token now reaches institutional endpoints without re-login.
	if resp, body := env.request(t, http.MethodGet, "/api/v1/analytics/overview", pendingToken, nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("approved account should read analytics: %d %s", resp.StatusCode, body)
	}

	// Revoking the approval takes effect immediately on the same session.
	resp, body = env.request(t, http.MethodPatch, "/api/v1/users/"+pendingID+"/institutional-approval",
		adminToken, map[string]any{"approved": false})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("revoke: %d %s", resp.StatusCode, body)
	}
	if decodeMap(t, body)["institutional_approved"] != false {
		t.Fatalf("revocation not persisted: %s", body)
	}
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/analytics/overview", pendingToken, nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("revoked account should lose institutional access, got %d", resp.StatusCode)
	}
}

func TestIntegrationUserFilters(t *testing.T) {
	env := newIntegrationEnv(t)

	adminToken, adminID := env.registerAndLogin(t, "filter-admin@test.dev", "mobile_user")
	if err := env.store.UpdateUserRole(context.Background(), adminID, models.RoleAdmin); err != nil {
		t.Fatalf("promote admin: %v", err)
	}

	// Ages are relative to today so the assertions stay valid over time.
	today := time.Now()
	type person struct {
		email, name, comuna, health, sex, living string
		age                                      int
	}
	people := []person{
		{"filter-ana@test.dev", "Ana Filtro", "Concepción", "fonasa", "female", "with_family", 80},
		{"filter-bruno@test.dev", "Bruno Filtro", "Talcahuano", "isapre", "male", "alone", 45},
		{"filter-carla@test.dev", "Carla Filtro", "Concepción", "fonasa", "female", "with_children", 80},
	}
	for _, p := range people {
		token, _ := env.registerAndLogin(t, p.email, "mobile_user")
		resp, body := env.request(t, http.MethodPatch, "/api/v1/users/me", token, map[string]any{
			"full_name":        p.name,
			"birth_date":       today.AddDate(-p.age, 0, 0).Format("2006-01-02"),
			"comuna":           p.comuna,
			"health_system":    p.health,
			"sex":              p.sex,
			"living_situation": p.living,
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("update profile %s: %d %s", p.email, resp.StatusCode, body)
		}
	}

	assertTotal := func(query string, want int) {
		t.Helper()
		resp, body := env.request(t, http.MethodGet, "/api/v1/users?role=mobile_user&"+query, adminToken, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET users?%s: %d %s", query, resp.StatusCode, body)
		}
		page, ok := decodeMap(t, body)["pagination"].(map[string]any)
		if !ok {
			t.Fatalf("missing pagination: %s", body)
		}
		total := int(page["total"].(float64))
		if total != want {
			t.Fatalf("GET users?%s: total = %d, want %d (%s)", query, total, want, body)
		}
	}

	// Partial, case-insensitive and accent-aware comuna matching.
	assertTotal("comuna=concep", 2)
	assertTotal("comuna="+url.QueryEscape("Concepción"), 2)
	assertTotal("comuna=talca", 1)
	// Age ranges use completed years.
	assertTotal("age_min=75&age_max=85", 2)
	assertTotal("age_min=0&age_max=50", 1)
	assertTotal("age_max=79", 1)
	// Individual categorical filters.
	assertTotal("sex=female", 2)
	assertTotal("health_system=fonasa", 2)
	assertTotal("living_situation=with_family", 1)
	// Combined filters.
	assertTotal("comuna=concep&sex=female&health_system=fonasa", 2)

	// Invalid filters are rejected by validation.
	invalid := []string{
		"sex=otro",
		"health_system=desconocido",
		"living_situation=solo",
		"age_min=abc",
		"age_min=90&age_max=10",
		"age_max=200",
	}
	for _, query := range invalid {
		resp, body := env.request(t, http.MethodGet, "/api/v1/users?"+query, adminToken, nil)
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("GET users?%s should be invalid, got %d %s", query, resp.StatusCode, body)
		}
	}

	// Non-administrators cannot list users with demographic data.
	personToken, _ := env.registerAndLogin(t, "filter-person@test.dev", "mobile_user")
	if resp, _ := env.request(t, http.MethodGet, "/api/v1/users?role=mobile_user", personToken, nil); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("non-admin should not list users, got %d", resp.StatusCode)
	}
}

func TestIntegrationSchemaHasConstraints(t *testing.T) {
	env := newIntegrationEnv(t)
	ctx := context.Background()

	// Email uniqueness is enforced at the database level.
	_, err := env.pool.Exec(ctx, `INSERT INTO users (email, password_hash, full_name) VALUES ('dup@test.dev','x','A')`)
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}
	if _, err := env.pool.Exec(ctx, `INSERT INTO users (email, password_hash, full_name) VALUES ('dup@test.dev','x','B')`); err == nil {
		t.Fatal("duplicate email constraint not enforced")
	}

	// Check constraints reject invalid coordinates.
	if _, err := env.pool.Exec(ctx, `INSERT INTO services (name, latitude) VALUES ('bad', 999)`); err == nil {
		t.Fatal("latitude check constraint not enforced")
	}

	// Foreign keys reject orphan enrollments.
	if _, err := env.pool.Exec(ctx, `INSERT INTO activity_enrollments (activity_id, user_id)
		VALUES ('11111111-1111-1111-1111-111111111111','22222222-2222-2222-2222-222222222222')`); err == nil {
		t.Fatal("foreign key constraint not enforced")
	}
}
