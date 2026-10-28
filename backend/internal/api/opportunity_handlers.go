package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
	"github.com/abudemy/backend/internal/store"
)

// ---- skills ----

func (s *Server) listMySkills(w http.ResponseWriter, r *http.Request) {
	skills, err := s.svc.Opportunities.ListSkills(r.Context(), mustPrincipal(r.Context()), queryString(r, "user_id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": skills})
}

type skillRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

func (s *Server) createSkill(w http.ResponseWriter, r *http.Request) {
	var req skillRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	skill, err := s.svc.Opportunities.CreateSkill(r.Context(), mustPrincipal(r.Context()), service.SkillInput{
		Name: req.Name, Description: req.Description, Level: models.SkillLevel(req.Level),
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, skill)
}

func (s *Server) updateSkill(w http.ResponseWriter, r *http.Request) {
	var req skillRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	skill, err := s.svc.Opportunities.UpdateSkill(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), service.SkillInput{
		Name: req.Name, Description: req.Description, Level: models.SkillLevel(req.Level),
	})
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

func (s *Server) deleteSkill(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Opportunities.DeleteSkill(r.Context(), mustPrincipal(r.Context()), r.PathValue("id")); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

// ---- offerings ----

func (s *Server) listOfferings(w http.ResponseWriter, r *http.Request) {
	filter := store.OfferingFilter{
		UserID:   queryString(r, "user_id"),
		Type:     queryString(r, "type"),
		Status:   queryString(r, "status"),
		Modality: queryString(r, "modality"),
		Search:   queryString(r, "search"),
		Comuna:   queryString(r, "comuna"),
		Page:     pageParams(r),
	}
	result, err := s.svc.Opportunities.ListOfferings(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type offeringRequest struct {
	SkillID     *string `json:"skill_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Modality    string  `json:"modality"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Comuna      string  `json:"comuna"`
	Region      string  `json:"region"`
}

func (req offeringRequest) toInput() service.OfferingInput {
	return service.OfferingInput{
		SkillID: req.SkillID, Title: req.Title, Description: req.Description,
		Type: models.OfferingType(req.Type), Status: models.OfferingStatus(req.Status),
		Modality: models.Modality(req.Modality), Price: req.Price, Currency: req.Currency,
		Comuna: req.Comuna, Region: req.Region,
	}
}

func (s *Server) createOffering(w http.ResponseWriter, r *http.Request) {
	var req offeringRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	created, err := s.svc.Opportunities.CreateOffering(r.Context(), mustPrincipal(r.Context()), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getOffering(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.Opportunities.GetOffering(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) updateOffering(w http.ResponseWriter, r *http.Request) {
	var req offeringRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Opportunities.UpdateOffering(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), req.toInput())
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) setOfferingStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Opportunities.SetOfferingStatus(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.OfferingStatus(req.Status)); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

// ---- opportunities ----

func (s *Server) listOpportunities(w http.ResponseWriter, r *http.Request) {
	filter := models.OpportunityFilter{
		Type:   queryString(r, "type"),
		Status: queryString(r, "status"),
		Comuna: queryString(r, "comuna"),
		Region: queryString(r, "region"),
		Search: queryString(r, "search"),
		Page:   pageParams(r),
	}
	result, err := s.svc.Opportunities.ListOpportunities(r.Context(), mustPrincipal(r.Context()), filter)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

type opportunityRequest struct {
	OrganizationID *string  `json:"organization_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Type           string   `json:"type"`
	Status         string   `json:"status"`
	Comuna         string   `json:"comuna"`
	Region         string   `json:"region"`
	SalaryMin      *float64 `json:"salary_min"`
	SalaryMax      *float64 `json:"salary_max"`
	Currency       string   `json:"currency"`
	ContactEmail   string   `json:"contact_email"`
	ExpiresAt      string   `json:"expires_at,omitempty"`
}

func (req opportunityRequest) toInput() (service.OpportunityInput, error) {
	expires, err := parseTime(req.ExpiresAt, "expires_at")
	if err != nil {
		return service.OpportunityInput{}, err
	}
	return service.OpportunityInput{
		OrganizationID: req.OrganizationID, Title: req.Title, Description: req.Description,
		Type: models.OpportunityType(req.Type), Status: models.OpportunityStatus(req.Status),
		Comuna: req.Comuna, Region: req.Region, SalaryMin: req.SalaryMin, SalaryMax: req.SalaryMax,
		Currency: req.Currency, ContactEmail: req.ContactEmail, ExpiresAt: expires,
	}, nil
}

func (s *Server) createOpportunity(w http.ResponseWriter, r *http.Request) {
	var req opportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	in, err := req.toInput()
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	created, err := s.svc.Opportunities.CreateOpportunity(r.Context(), mustPrincipal(r.Context()), in)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getOpportunity(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.Opportunities.GetOpportunity(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) updateOpportunity(w http.ResponseWriter, r *http.Request) {
	var req opportunityRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	in, err := req.toInput()
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	updated, err := s.svc.Opportunities.UpdateOpportunity(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), in)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) setOpportunityStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Opportunities.SetOpportunityStatus(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), models.OpportunityStatus(req.Status)); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}

type applyRequest struct {
	Message string `json:"message"`
}

func (s *Server) applyOpportunity(w http.ResponseWriter, r *http.Request) {
	var req applyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	app, err := s.svc.Opportunities.Apply(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), req.Message)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusCreated, app)
}

func (s *Server) listApplications(w http.ResponseWriter, r *http.Request) {
	page := pageParams(r)
	result, err := s.svc.Opportunities.ListApplications(r.Context(), mustPrincipal(r.Context()), r.PathValue("id"), page)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) updateApplication(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if err := s.svc.Opportunities.UpdateApplicationStatus(r.Context(), mustPrincipal(r.Context()),
		r.PathValue("id"), r.PathValue("appId"), models.ApplicationStatus(req.Status)); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}
