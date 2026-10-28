package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
)

// ---- health ----

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if s.pinger != nil {
		if err := s.pinger.Ping(r.Context()); err != nil {
			writeError(w, r, s.log, apperr.Internal(err))
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// ---- auth ----

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	FullName        string `json:"full_name"`
	Phone           string `json:"phone"`
	BirthDate       string `json:"birth_date,omitempty"`
	Comuna          string `json:"comuna,omitempty"`
	HealthSystem    string `json:"health_system,omitempty"`
	Sex             string `json:"sex,omitempty"`
	LivingSituation string `json:"living_situation,omitempty"`
	Role            string `json:"role,omitempty"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	birthDate, err := parseDate(req.BirthDate, "birth_date")
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	user, err := s.svc.Auth.Register(r.Context(), service.RegisterInput{
		Email: req.Email, Password: req.Password, FullName: req.FullName,
		Phone: req.Phone, BirthDate: birthDate,
		Comuna:          req.Comuna,
		HealthSystem:    models.HealthSystem(req.HealthSystem),
		Sex:             models.Sex(req.Sex),
		LivingSituation: models.LivingSituation(req.LivingSituation),
		Role:            models.Role(req.Role),
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	pair, err := s.svc.Auth.Login(r.Context(), service.LoginInput{
		Email: req.Email, Password: req.Password,
		UserAgent: r.UserAgent(), IP: clientIP(r),
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	pair, err := s.svc.Auth.Refresh(r.Context(), req.RefreshToken, r.UserAgent(), clientIP(r))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Auth.Logout(r.Context(), mustPrincipal(r.Context())); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) logoutAll(w http.ResponseWriter, r *http.Request) {
	p := mustPrincipal(r.Context())
	if err := s.svc.Auth.LogoutAll(r.Context(), p.UserID); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (s *Server) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Auth.RequestPasswordReset(r.Context(), req.Email); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "if the email is registered, password reset instructions have been dispatched",
	})
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Auth.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.svc.Auth.Sessions(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": sessions})
}

func (s *Server) revokeSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, r, s.log, apperr.InvalidField("id", "is required"))
		return
	}
	if err := s.svc.Auth.RevokeSession(r.Context(), mustPrincipal(r.Context()), id); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

// ---- helpers ----

func parseDate(value, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, apperr.InvalidField(field, "must use YYYY-MM-DD format")
	}
	return &t, nil
}

func parseTime(value, field string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return &t, nil
	}
	return nil, apperr.InvalidField(field, "must use RFC3339 format")
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}
