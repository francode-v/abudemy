package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
)

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := s.svc.Catalog.Categories(r.Context())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": cats})
}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request) {
	filter := models.ServiceFilter{
		CategoryID:   queryString(r, "category_id"),
		Type:         queryString(r, "type"),
		Status:       queryString(r, "status"),
		Comuna:       queryString(r, "comuna"),
		Region:       queryString(r, "region"),
		Search:       queryString(r, "search"),
		Organization: queryString(r, "organization_id"),
		NearLat:      queryFloatPtr(r, "lat"),
		NearLng:      queryFloatPtr(r, "lng"),
		RadiusKM:     queryFloatPtr(r, "radius_km"),
		Page:         pageParams(r),
	}
	result, err := s.svc.Catalog.List(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getService(w http.ResponseWriter, r *http.Request) {
	svc, err := s.svc.Catalog.Get(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

type serviceRequest struct {
	OrganizationID *string                `json:"organization_id"`
	CategoryID     *string                `json:"category_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"`
	Address        string                 `json:"address"`
	Comuna         string                 `json:"comuna"`
	Region         string                 `json:"region"`
	Phone          string                 `json:"phone"`
	Email          string                 `json:"email"`
	Website        string                 `json:"website"`
	Latitude       *float64               `json:"latitude"`
	Longitude      *float64               `json:"longitude"`
	Schedule       []models.ScheduleEntry `json:"schedule"`
}

func (req serviceRequest) toInput() service.ServiceInput {
	return service.ServiceInput{
		OrganizationID: req.OrganizationID,
		CategoryID:     req.CategoryID,
		Name:           req.Name,
		Description:    req.Description,
		Type:           models.ServiceType(req.Type),
		Status:         models.ServiceStatus(req.Status),
		Address:        req.Address,
		Comuna:         req.Comuna,
		Region:         req.Region,
		Phone:          req.Phone,
		Email:          req.Email,
		Website:        req.Website,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		Schedule:       req.Schedule,
	}
}

func (s *Server) createService(w http.ResponseWriter, r *http.Request) {
	var req serviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	created, err := s.svc.Catalog.Create(r.Context(), mustPrincipal(r.Context()), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateService(w http.ResponseWriter, r *http.Request) {
	var req serviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Catalog.Update(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

type statusRequest struct {
	Status string `json:"status"`
}

func (s *Server) setServiceStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Catalog.SetStatus(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.ServiceStatus(req.Status))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteService(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Catalog.Delete(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}
