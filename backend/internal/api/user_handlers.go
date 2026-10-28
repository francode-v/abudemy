package api

import (
	"net/http"
	"strconv"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.svc.Users.Profile(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type updateProfileRequest struct {
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date,omitempty"`
	// Pointers distinguish "not sent" from "sent empty": omitted fields keep
	// their stored value (the web portal does not manage these fields).
	Comuna          *string `json:"comuna,omitempty"`
	HealthSystem    *string `json:"health_system,omitempty"`
	Sex             *string `json:"sex,omitempty"`
	LivingSituation *string `json:"living_situation,omitempty"`
}

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	var req updateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	birthDate, err := parseDate(req.BirthDate, "birth_date")
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	in := service.UpdateProfileInput{
		FullName: req.FullName, Phone: req.Phone, BirthDate: birthDate,
	}
	if req.Comuna != nil {
		in.Comuna = req.Comuna
	}
	if req.HealthSystem != nil {
		value := models.HealthSystem(*req.HealthSystem)
		in.HealthSystem = &value
	}
	if req.Sex != nil {
		value := models.Sex(*req.Sex)
		in.Sex = &value
	}
	if req.LivingSituation != nil {
		value := models.LivingSituation(*req.LivingSituation)
		in.LivingSituation = &value
	}
	user, err := s.svc.Users.UpdateProfile(r.Context(), mustPrincipal(r.Context()), in)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Auth.ChangePassword(r.Context(), mustPrincipal(r.Context()), req.CurrentPassword, req.NewPassword); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) myEnrollments(w http.ResponseWriter, r *http.Request) {
	page := pageParams(r)
	result, err := s.svc.Activities.MyEnrollments(r.Context(), mustPrincipal(r.Context()), page)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	ageMin, err := queryAge(r, "age_min")
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	ageMax, err := queryAge(r, "age_max")
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	filter := store.UserFilter{
		Role:            queryString(r, "role"),
		Status:          queryString(r, "status"),
		Search:          queryString(r, "search"),
		Comuna:          queryString(r, "comuna"),
		HealthSystem:    queryString(r, "health_system"),
		Sex:             queryString(r, "sex"),
		LivingSituation: queryString(r, "living_situation"),
		AgeMin:          ageMin,
		AgeMax:          ageMax,
		Page:            pageParams(r),
	}
	result, err := s.svc.Users.List(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// queryAge reads an optional age filter. The full range is validated by the
// service so the error mentions the field and the allowed values.
func queryAge(r *http.Request, key string) (*int, error) {
	value := queryString(r, key)
	if value == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return nil, apperr.InvalidField(key, "must be an integer")
	}
	return &n, nil
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	user, err := s.svc.Users.Get(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type userStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	var req userStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	user, err := s.svc.Users.UpdateStatus(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.UserStatus(req.Status))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type userRoleRequest struct {
	Role string `json:"role"`
}

func (s *Server) updateUserRole(w http.ResponseWriter, r *http.Request) {
	var req userRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	user, err := s.svc.Users.UpdateRole(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.Role(req.Role))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

type institutionalApprovalRequest struct {
	Approved *bool `json:"approved"`
}

func (s *Server) updateInstitutionalApproval(w http.ResponseWriter, r *http.Request) {
	var req institutionalApprovalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	// Require the field explicitly so an empty body can never revoke access.
	if req.Approved == nil {
		writeError(w, r, s.log, apperr.InvalidField("approved", "is required"))
		return
	}
	user, err := s.svc.Users.SetInstitutionalApproval(
		r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), *req.Approved)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}
