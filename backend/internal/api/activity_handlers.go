package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
)

func (s *Server) listActivities(w http.ResponseWriter, r *http.Request) {
	filter := models.ActivityFilter{
		Type:         queryString(r, "type"),
		Status:       queryString(r, "status"),
		Modality:     queryString(r, "modality"),
		Comuna:       queryString(r, "comuna"),
		Region:       queryString(r, "region"),
		Search:       queryString(r, "search"),
		Organization: queryString(r, "organization_id"),
		From:         queryTimePtr(r, "from"),
		To:           queryTimePtr(r, "to"),
		Page:         pageParams(r),
	}
	result, err := s.svc.Activities.List(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getActivity(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.Activities.Get(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

type activityRequest struct {
	OrganizationID *string  `json:"organization_id"`
	ServiceID      *string  `json:"service_id"`
	CategoryID     *string  `json:"category_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Type           string   `json:"type"`
	Modality       string   `json:"modality"`
	Status         string   `json:"status"`
	Location       string   `json:"location"`
	Comuna         string   `json:"comuna"`
	Region         string   `json:"region"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	StartAt        string   `json:"start_at"`
	EndAt          string   `json:"end_at,omitempty"`
	Capacity       int      `json:"capacity"`
	Price          float64  `json:"price"`
	Currency       string   `json:"currency"`
}

func (req activityRequest) toInput() (service.ActivityInput, error) {
	start, err := parseTime(req.StartAt, "start_at")
	if err != nil {
		return service.ActivityInput{}, err
	}
	end, err := parseTime(req.EndAt, "end_at")
	if err != nil {
		return service.ActivityInput{}, err
	}
	in := service.ActivityInput{
		OrganizationID: req.OrganizationID,
		ServiceID:      req.ServiceID,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           models.ActivityType(req.Type),
		Modality:       models.Modality(req.Modality),
		Status:         models.ActivityStatus(req.Status),
		Location:       req.Location,
		Comuna:         req.Comuna,
		Region:         req.Region,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		EndAt:          end,
		Capacity:       req.Capacity,
		Price:          req.Price,
		Currency:       req.Currency,
	}
	if start != nil {
		in.StartAt = *start
	}
	return in, nil
}

func (s *Server) createActivity(w http.ResponseWriter, r *http.Request) {
	var req activityRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	in, err := req.toInput()
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	created, err := s.svc.Activities.Create(r.Context(), mustPrincipal(r.Context()), in)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateActivity(w http.ResponseWriter, r *http.Request) {
	var req activityRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	in, err := req.toInput()
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Activities.Update(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), in)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) setActivityStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Activities.SetStatus(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.ActivityStatus(req.Status))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) enrollActivity(w http.ResponseWriter, r *http.Request) {
	res, err := s.svc.Activities.Enroll(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"enrollment": res.Enrollment,
		"waitlisted": res.Waitlisted,
	})
}

func (s *Server) cancelEnrollment(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Activities.CancelEnrollment(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) listEnrollments(w http.ResponseWriter, r *http.Request) {
	page := pageParams(r)
	result, err := s.svc.Activities.ListEnrollments(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), page)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type enrollmentStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) updateEnrollmentStatus(w http.ResponseWriter, r *http.Request) {
	var req enrollmentStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Activities.UpdateEnrollmentStatus(r.Context(), mustPrincipal(r.Context()),
		r.PathValue("id"), r.PathValue("userId"), models.EnrollmentStatus(req.Status)); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}
