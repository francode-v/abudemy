package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

func (s *Server) listContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := s.svc.Assistance.ListContacts(r.Context(), mustPrincipal(r.Context()))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": contacts})
}

type contactRequest struct {
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Priority     int    `json:"priority"`
}

func (req contactRequest) toInput() service.ContactInput {
	return service.ContactInput{
		Name: req.Name, Relationship: req.Relationship,
		Phone: req.Phone, Email: req.Email, Priority: req.Priority,
	}
}

func (s *Server) createContact(w http.ResponseWriter, r *http.Request) {
	var req contactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	c, err := s.svc.Assistance.CreateContact(r.Context(), mustPrincipal(r.Context()), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) updateContact(w http.ResponseWriter, r *http.Request) {
	var req contactRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	c, err := s.svc.Assistance.UpdateContact(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) deleteContact(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Assistance.DeleteContact(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

func (s *Server) listAssistanceRequests(w http.ResponseWriter, r *http.Request) {
	filter := store.AssistanceFilter{
		Status: queryString(r, "status"),
		From:   queryTimePtr(r, "from"),
		To:     queryTimePtr(r, "to"),
		Page:   pageParams(r),
	}
	result, err := s.svc.Assistance.ListRequests(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type assistanceRequestDTO struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	ContactID   *string  `json:"contact_id"`
}

func (s *Server) createAssistanceRequest(w http.ResponseWriter, r *http.Request) {
	var req assistanceRequestDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	created, err := s.svc.Assistance.CreateRequest(r.Context(), mustPrincipal(r.Context()), service.AssistanceInput{
		Type: req.Type, Description: req.Description,
		Latitude: req.Latitude, Longitude: req.Longitude, ContactID: req.ContactID,
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getAssistanceRequest(w http.ResponseWriter, r *http.Request) {
	req, err := s.svc.Assistance.GetRequest(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (s *Server) updateAssistanceStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Assistance.UpdateStatus(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.AssistanceStatus(req.Status))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
