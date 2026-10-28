// Package api is the HTTP transport layer. Handlers decode requests, delegate
// to the service layer and render responses; they contain no business rules.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

// Pinger checks datastore connectivity for the readiness probe.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Config configures the HTTP server.
type Config struct {
	Services    *service.Services
	Store       *store.Store
	Tokens      *auth.TokenManager
	Logger      *slog.Logger
	CORSOrigins []string
	Pinger      Pinger
}

// Server holds the HTTP dependencies.
type Server struct {
	svc         *service.Services
	store       *store.Store
	tokens      *auth.TokenManager
	log         *slog.Logger
	corsOrigins []string
	pinger      Pinger
}

// NewServer builds the HTTP server.
func NewServer(cfg Config) *Server {
	return &Server{
		svc:         cfg.Services,
		store:       cfg.Store,
		tokens:      cfg.Tokens,
		log:         cfg.Logger,
		corsOrigins: cfg.CORSOrigins,
		pinger:      cfg.Pinger,
	}
}

type middleware func(http.Handler) http.Handler

func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// handle registers a route with per-route middleware.
func (s *Server) handle(mux *http.ServeMux, pattern string, fn http.HandlerFunc, mws ...middleware) {
	mux.Handle(pattern, chain(fn, mws...))
}

// Handler builds the full route table wrapped in global middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health endpoints (public).
	s.handle(mux, "GET /healthz", s.health)
	s.handle(mux, "GET /readyz", s.ready)

	// ---- Authentication (public) ----
	s.handle(mux, "POST /api/v1/auth/register", s.register)
	s.handle(mux, "POST /api/v1/auth/login", s.login)
	s.handle(mux, "POST /api/v1/auth/refresh", s.refresh)
	s.handle(mux, "POST /api/v1/auth/password/forgot", s.forgotPassword)
	s.handle(mux, "POST /api/v1/auth/password/reset", s.resetPassword)

	// ---- Authentication (authenticated) ----
	s.handle(mux, "POST /api/v1/auth/logout", s.logout, s.auth)
	s.handle(mux, "POST /api/v1/auth/logout-all", s.logoutAll, s.auth)
	s.handle(mux, "GET /api/v1/auth/sessions", s.listSessions, s.auth)
	s.handle(mux, "DELETE /api/v1/auth/sessions/{id}", s.revokeSession, s.auth)

	// ---- Users ----
	s.handle(mux, "GET /api/v1/users/me", s.getMe, s.auth)
	s.handle(mux, "PATCH /api/v1/users/me", s.updateMe, s.auth)
	s.handle(mux, "POST /api/v1/users/me/password", s.changePassword, s.auth)
	s.handle(mux, "POST /api/v1/users/me/email", s.changeEmail, s.auth)
	s.handle(mux, "GET /api/v1/users/me/preferences", s.getPreferences, s.auth)
	s.handle(mux, "PUT /api/v1/users/me/preferences", s.updatePreferences, s.auth)
	s.handle(mux, "GET /api/v1/users/me/enrollments", s.myEnrollments, s.auth)
	s.handle(mux, "GET /api/v1/users/me/applications", s.myApplications, s.auth)
	s.handle(mux, "GET /api/v1/users/me/reminders", s.listMyReminders, s.auth)
	s.handle(mux, "GET /api/v1/users/me/skills", s.listMySkills, s.auth)
	s.handle(mux, "POST /api/v1/users/me/skills", s.createSkill, s.auth)
	s.handle(mux, "GET /api/v1/users", s.listUsers, s.auth, s.requireRoles(roleRegisteredPeopleReaders...))
	s.handle(mux, "GET /api/v1/users/{id}", s.getUser, s.auth)
	s.handle(mux, "PATCH /api/v1/users/{id}/status", s.updateUserStatus, s.auth, s.requireRoles(roleAdminOnly...))
	s.handle(mux, "PATCH /api/v1/users/{id}/role", s.updateUserRole, s.auth, s.requireRoles(roleAdminOnly...))
	s.handle(mux, "PATCH /api/v1/users/{id}/institutional-approval", s.updateInstitutionalApproval, s.auth, s.requireRoles(roleAdminOnly...))

	// ---- Services catalogue ----
	s.handle(mux, "GET /api/v1/service-categories", s.listCategories, s.auth)
	s.handle(mux, "GET /api/v1/services", s.listServices, s.auth)
	s.handle(mux, "POST /api/v1/services", s.createService, s.auth)
	s.handle(mux, "GET /api/v1/services/{id}", s.getService, s.auth)
	s.handle(mux, "PUT /api/v1/services/{id}", s.updateService, s.auth)
	s.handle(mux, "PATCH /api/v1/services/{id}/status", s.setServiceStatus, s.auth)
	s.handle(mux, "DELETE /api/v1/services/{id}", s.deleteService, s.auth)

	// ---- Activities ----
	s.handle(mux, "GET /api/v1/activities", s.listActivities, s.auth)
	s.handle(mux, "POST /api/v1/activities", s.createActivity, s.auth)
	s.handle(mux, "GET /api/v1/activities/{id}", s.getActivity, s.auth)
	s.handle(mux, "PUT /api/v1/activities/{id}", s.updateActivity, s.auth)
	s.handle(mux, "PATCH /api/v1/activities/{id}/status", s.setActivityStatus, s.auth)
	s.handle(mux, "POST /api/v1/activities/{id}/enrollments", s.enrollActivity, s.auth)
	s.handle(mux, "DELETE /api/v1/activities/{id}/enrollments/me", s.cancelEnrollment, s.auth)
	s.handle(mux, "GET /api/v1/activities/{id}/enrollments", s.listEnrollments, s.auth)
	s.handle(mux, "PATCH /api/v1/activities/{id}/enrollments/{userId}", s.updateEnrollmentStatus, s.auth)
	s.handle(mux, "POST /api/v1/activities/{id}/reminders", s.setActivityReminder, s.auth)
	s.handle(mux, "DELETE /api/v1/activities/{id}/reminders/me", s.cancelActivityReminder, s.auth)

	// ---- Skills & offerings ----
	s.handle(mux, "PUT /api/v1/skills/{id}", s.updateSkill, s.auth)
	s.handle(mux, "DELETE /api/v1/skills/{id}", s.deleteSkill, s.auth)
	s.handle(mux, "GET /api/v1/offerings", s.listOfferings, s.auth)
	s.handle(mux, "POST /api/v1/offerings", s.createOffering, s.auth)
	s.handle(mux, "GET /api/v1/offerings/{id}", s.getOffering, s.auth)
	s.handle(mux, "PUT /api/v1/offerings/{id}", s.updateOffering, s.auth)
	s.handle(mux, "PATCH /api/v1/offerings/{id}/status", s.setOfferingStatus, s.auth)

	// ---- Opportunities ----
	s.handle(mux, "GET /api/v1/opportunities", s.listOpportunities, s.auth)
	s.handle(mux, "POST /api/v1/opportunities", s.createOpportunity, s.auth)
	s.handle(mux, "GET /api/v1/opportunities/{id}", s.getOpportunity, s.auth)
	s.handle(mux, "PUT /api/v1/opportunities/{id}", s.updateOpportunity, s.auth)
	s.handle(mux, "PATCH /api/v1/opportunities/{id}/status", s.setOpportunityStatus, s.auth)
	s.handle(mux, "POST /api/v1/opportunities/{id}/applications", s.applyOpportunity, s.auth)
	s.handle(mux, "GET /api/v1/opportunities/{id}/applications", s.listApplications, s.auth)
	s.handle(mux, "PATCH /api/v1/opportunities/{id}/applications/{appId}", s.updateApplication, s.auth)
	s.handle(mux, "POST /api/v1/opportunities/{id}/reports", s.reportOpportunity, s.auth)

	// ---- Push devices ----
	s.handle(mux, "GET /api/v1/devices", s.listDevices, s.auth)
	s.handle(mux, "POST /api/v1/devices", s.registerDevice, s.auth)
	s.handle(mux, "DELETE /api/v1/devices/{id}", s.deleteDevice, s.auth)

	// ---- Notification feed ----
	s.handle(mux, "GET /api/v1/notifications", s.listNotifications, s.auth)
	s.handle(mux, "GET /api/v1/notifications/unread-count", s.unreadNotifications, s.auth)
	s.handle(mux, "POST /api/v1/notifications/read-all", s.markAllNotificationsRead, s.auth)
	s.handle(mux, "POST /api/v1/notifications/{id}/read", s.markNotificationRead, s.auth)

	// ---- Assistant (curated knowledge base; optional provider) ----
	s.handle(mux, "GET /api/v1/assistant/topics", s.listAssistantTopics, s.auth)
	s.handle(mux, "GET /api/v1/assistant/answer", s.askAssistant, s.auth)

	// ---- Geocoding (optional provider; 501 when unconfigured) ----
	s.handle(mux, "GET /api/v1/geo/search", s.geoSearch, s.auth)
	s.handle(mux, "GET /api/v1/geo/reverse", s.geoReverse, s.auth)

	// ---- Payments capability (honest: unavailable unless a provider exists) ----
	s.handle(mux, "GET /api/v1/payments/config", s.paymentsConfig, s.auth)
	s.handle(mux, "POST /api/v1/payments/intents", s.createPaymentIntent, s.auth)

	// ---- Assistance ----
	s.handle(mux, "GET /api/v1/assistance/contacts", s.listContacts, s.auth)
	s.handle(mux, "POST /api/v1/assistance/contacts", s.createContact, s.auth)
	s.handle(mux, "PUT /api/v1/assistance/contacts/{id}", s.updateContact, s.auth)
	s.handle(mux, "DELETE /api/v1/assistance/contacts/{id}", s.deleteContact, s.auth)
	s.handle(mux, "GET /api/v1/assistance/requests", s.listAssistanceRequests, s.auth)
	s.handle(mux, "POST /api/v1/assistance/requests", s.createAssistanceRequest, s.auth)
	s.handle(mux, "GET /api/v1/assistance/requests/{id}", s.getAssistanceRequest, s.auth)
	s.handle(mux, "PATCH /api/v1/assistance/requests/{id}/status", s.updateAssistanceStatus, s.auth)

	// ---- Organizations ----
	s.handle(mux, "GET /api/v1/organizations", s.listOrganizations, s.auth)
	s.handle(mux, "POST /api/v1/organizations", s.createOrganization, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "GET /api/v1/organizations/mine", s.listMyOrganizations, s.auth)
	s.handle(mux, "GET /api/v1/organizations/{id}", s.getOrganization, s.auth)
	s.handle(mux, "PUT /api/v1/organizations/{id}", s.updateOrganization, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "GET /api/v1/organizations/{id}/members", s.listOrganizationMembers, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "POST /api/v1/organizations/{id}/members", s.addOrganizationMember, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "PATCH /api/v1/organizations/{id}/members/{userId}", s.updateOrganizationMember, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "DELETE /api/v1/organizations/{id}/members/{userId}", s.removeOrganizationMember, s.auth, s.requireRoles(roleInstitutional...))

	// ---- Analytics & reports ----
	s.handle(mux, "GET /api/v1/analytics/overview", s.analyticsOverview, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "GET /api/v1/analytics/reports", s.listReports, s.auth, s.requireRoles(roleInstitutional...))
	s.handle(mux, "GET /api/v1/analytics/reports/{type}", s.exportReport, s.auth, s.requireRoles(roleInstitutional...))

	return chain(mux, s.requestID, s.recoverer, s.logging, s.cors)
}
