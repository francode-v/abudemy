package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

func (s *Server) listOrganizations(w http.ResponseWriter, r *http.Request) {
	filter := store.OrganizationFilter{
		Type:   queryString(r, "type"),
		Search: queryString(r, "search"),
		Active: queryBoolPtr(r, "active"),
		Page:   pageParams(r),
	}
	result, err := s.svc.Organizations.List(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) listMyOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := s.svc.Organizations.ListMine(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": orgs})
}

type organizationRequest struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	LegalID     string   `json:"legal_id"`
	Description string   `json:"description"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	Address     string   `json:"address"`
	Comuna      string   `json:"comuna"`
	Region      string   `json:"region"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Active      *bool    `json:"active"`
}

func (req organizationRequest) toInput() service.OrganizationInput {
	return service.OrganizationInput{
		Name: req.Name, Type: models.OrganizationType(req.Type), LegalID: req.LegalID,
		Description: req.Description, Email: req.Email, Phone: req.Phone, Address: req.Address,
		Comuna: req.Comuna, Region: req.Region, Latitude: req.Latitude, Longitude: req.Longitude,
		Active: req.Active,
	}
}

func (s *Server) createOrganization(w http.ResponseWriter, r *http.Request) {
	var req organizationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	created, err := s.svc.Organizations.Create(r.Context(), mustPrincipal(r.Context()), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getOrganization(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.Organizations.Get(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var req organizationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	o, err := s.svc.Organizations.Update(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) listOrganizationMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.svc.Organizations.ListMembers(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": members})
}

type memberRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func (s *Server) addOrganizationMember(w http.ResponseWriter, r *http.Request) {
	var req memberRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	member, err := s.svc.Organizations.AddMember(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"),
		service.AddMemberInput{UserID: req.UserID, Email: req.Email, Role: models.OrgMemberRole(req.Role)})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (s *Server) updateOrganizationMember(w http.ResponseWriter, r *http.Request) {
	var req memberRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Organizations.UpdateMemberRole(r.Context(), mustPrincipal(r.Context()),
		r.PathValue("id"), r.PathValue("userId"), models.OrgMemberRole(req.Role)); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) removeOrganizationMember(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Organizations.RemoveMember(r.Context(), mustPrincipal(r.Context()),
		r.PathValue("id"), r.PathValue("userId")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}
