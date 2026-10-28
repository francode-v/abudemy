package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
)

// ---- preferences ----

func (s *Server) getPreferences(w http.ResponseWriter, r *http.Request) {
	prefs, err := s.svc.Users.Preferences(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

type preferencesRequest struct {
	Interests     []string                        `json:"interests"`
	Accessibility models.AccessibilityPreferences `json:"accessibility"`
	Notifications models.NotificationPreferences  `json:"notifications"`
}

func (s *Server) updatePreferences(w http.ResponseWriter, r *http.Request) {
	var req preferencesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	prefs, err := s.svc.Users.UpdatePreferences(r.Context(), mustPrincipal(r.Context()), service.PreferencesInput{
		Interests:     req.Interests,
		Accessibility: req.Accessibility,
		Notifications: req.Notifications,
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

// ---- email change ----

type changeEmailRequest struct {
	NewEmail string `json:"new_email"`
	Password string `json:"password"`
}

func (s *Server) changeEmail(w http.ResponseWriter, r *http.Request) {
	var req changeEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	user, err := s.svc.Users.ChangeEmail(r.Context(), mustPrincipal(r.Context()), service.ChangeEmailInput{
		NewEmail: req.NewEmail, Password: req.Password,
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// ---- my applications ----

func (s *Server) myApplications(w http.ResponseWriter, r *http.Request) {
	page := pageParams(r)
	result, err := s.svc.Opportunities.MyApplications(r.Context(), mustPrincipal(r.Context()), page)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ---- reminders ----

func (s *Server) listMyReminders(w http.ResponseWriter, r *http.Request) {
	page := pageParams(r)
	result, err := s.svc.Activities.MyReminders(r.Context(), mustPrincipal(r.Context()), page)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type reminderRequest struct {
	RemindAt string `json:"remind_at,omitempty"`
}

func (s *Server) setActivityReminder(w http.ResponseWriter, r *http.Request) {
	var req reminderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	remindAt, err := parseTime(req.RemindAt, "remind_at")
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	reminder, err := s.svc.Activities.SetReminder(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), remindAt)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, reminder)
}

func (s *Server) cancelActivityReminder(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Activities.CancelReminder(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

// ---- opportunity reports ----

type reportRequest struct {
	Reason  string `json:"reason"`
	Details string `json:"details"`
}

func (s *Server) reportOpportunity(w http.ResponseWriter, r *http.Request) {
	var req reportRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	report, err := s.svc.Opportunities.Report(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), service.ReportInput{
		Reason: req.Reason, Details: req.Details,
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, report)
}

// ---- devices ----

func (s *Server) listDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := s.svc.Notifications.ListDevices(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": devices})
}

type deviceRequest struct {
	Platform   string `json:"platform"`
	Token      string `json:"token"`
	DeviceName string `json:"device_name"`
	AppVersion string `json:"app_version"`
}

func (s *Server) registerDevice(w http.ResponseWriter, r *http.Request) {
	var req deviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	device, err := s.svc.Notifications.RegisterDevice(r.Context(), mustPrincipal(r.Context()), service.DeviceInput{
		Platform: models.DevicePlatform(req.Platform), Token: req.Token,
		DeviceName: req.DeviceName, AppVersion: req.AppVersion,
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, device)
}

func (s *Server) deleteDevice(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Notifications.DeleteDevice(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

// ---- notifications ----

func (s *Server) listNotifications(w http.ResponseWriter, r *http.Request) {
	page := pageParams(r)
	result, err := s.svc.Notifications.List(r.Context(), mustPrincipal(r.Context()), page)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) unreadNotifications(w http.ResponseWriter, r *http.Request) {
	n, err := s.svc.Notifications.UnreadCount(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"unread": n})
}

func (s *Server) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Notifications.MarkRead(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) markAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	n, err := s.svc.Notifications.MarkAllRead(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"updated": n})
}
