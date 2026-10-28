package service

import (
	"context"
	"errors"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/store"
	"github.com/abudemy/backend/internal/validate"
)

// OpportunityService manages skills, user offerings and economic opportunities.
type OpportunityService struct{ d Deps }

// SkillInput is the payload for creating or updating a skill.
type SkillInput struct {
	Name        string
	Description string
	Level       models.SkillLevel
}

// OfferingInput is the payload for creating or updating an offering.
type OfferingInput struct {
	SkillID     *string
	Title       string
	Description string
	Type        models.OfferingType
	Status      models.OfferingStatus
	Modality    models.Modality
	Price       float64
	Currency    string
	Comuna      string
	Region      string
}

// OpportunityInput is the payload for creating or updating an opportunity.
type OpportunityInput struct {
	OrganizationID *string
	Title          string
	Description    string
	Type           models.OpportunityType
	Status         models.OpportunityStatus
	Comuna         string
	Region         string
	SalaryMin      *float64
	SalaryMax      *float64
	Currency       string
	ContactEmail   string
	ExpiresAt      *time.Time
}

// ---- Skills ----

// ListSkills returns the skills of a user. Callers may only list their own
// skills unless they are administrators.
func (s *OpportunityService) ListSkills(ctx context.Context, p Principal, userID string) ([]models.Skill, error) {
	if userID == "" {
		userID = p.UserID
	}
	if userID != p.UserID && !p.IsAdmin() {
		return nil, apperr.Forbidden("you may only access your own skills")
	}
	skills, err := s.d.Store.ListSkills(ctx, userID)
	if err != nil {
		return nil, mapErr(err, "skill")
	}
	return skills, nil
}

// CreateSkill adds a skill to the caller's profile.
func (s *OpportunityService) CreateSkill(ctx context.Context, p Principal, in SkillInput) (*models.Skill, error) {
	var v validate.Errors
	v.Required("name", in.Name)
	v.MaxLen("name", in.Name, 120)
	v.MaxLen("description", in.Description, 1000)
	v.OneOf("level", string(in.Level),
		string(models.SkillLevelBeginner), string(models.SkillLevelIntermediate),
		string(models.SkillLevelAdvanced), string(models.SkillLevelExpert))
	if err := v.Err(); err != nil {
		return nil, err
	}
	level := in.Level
	if level == "" {
		level = models.SkillLevelBeginner
	}
	skill, err := s.d.Store.CreateSkill(ctx, &models.Skill{
		UserID: p.UserID, Name: in.Name, Description: in.Description, Level: level,
	})
	if err != nil {
		return nil, mapErr(err, "skill")
	}
	return skill, nil
}

// UpdateSkill updates one of the caller's skills.
func (s *OpportunityService) UpdateSkill(ctx context.Context, p Principal, id string, in SkillInput) (*models.Skill, error) {
	existing, err := s.d.Store.GetSkill(ctx, id)
	if err != nil {
		return nil, mapErr(err, "skill")
	}
	if existing.UserID != p.UserID && !p.IsAdmin() {
		return nil, apperr.Forbidden("you may only modify your own skills")
	}
	if in.Name == "" {
		return nil, apperr.InvalidField("name", "is required")
	}
	if in.Level != "" && !in.Level.Valid() {
		return nil, apperr.InvalidField("level", "unknown level")
	}
	existing.Name = in.Name
	existing.Description = in.Description
	if in.Level != "" {
		existing.Level = in.Level
	}
	updated, err := s.d.Store.UpdateSkill(ctx, existing)
	if err != nil {
		return nil, mapErr(err, "skill")
	}
	return updated, nil
}

// DeleteSkill removes one of the caller's skills.
func (s *OpportunityService) DeleteSkill(ctx context.Context, p Principal, id string) error {
	return mapErr(s.d.Store.DeleteSkill(ctx, id, p.UserID), "skill")
}

// ---- Offerings ----

// ListOfferings returns a page of offerings.
func (s *OpportunityService) ListOfferings(ctx context.Context, _ Principal, f store.OfferingFilter) (models.Page[models.Offering], error) {
	items, total, err := s.d.Store.ListOfferings(ctx, f)
	if err != nil {
		return models.Page[models.Offering]{}, mapErr(err, "offering")
	}
	return models.NewPage(items, total, f.Page), nil
}

// GetOffering returns an offering by id.
func (s *OpportunityService) GetOffering(ctx context.Context, _ Principal, id string) (*models.Offering, error) {
	o, err := s.d.Store.GetOffering(ctx, id)
	if err != nil {
		return nil, mapErr(err, "offering")
	}
	return o, nil
}

// CreateOffering publishes a class or service for the caller.
func (s *OpportunityService) CreateOffering(ctx context.Context, p Principal, in OfferingInput) (*models.Offering, error) {
	if err := validateOfferingInput(in); err != nil {
		return nil, err
	}
	o := &models.Offering{
		UserID: p.UserID, SkillID: in.SkillID, Title: in.Title, Description: in.Description,
		Type: defaultOfferingType(in.Type), Status: defaultOfferingStatus(in.Status),
		Modality: defaultModality(in.Modality), Price: in.Price, Currency: defaultCurrency(in.Currency),
		Comuna: in.Comuna, Region: in.Region,
	}
	created, err := s.d.Store.CreateOffering(ctx, o)
	if err != nil {
		return nil, mapErr(err, "offering")
	}
	return created, nil
}

// UpdateOffering updates one of the caller's offerings.
func (s *OpportunityService) UpdateOffering(ctx context.Context, p Principal, id string, in OfferingInput) (*models.Offering, error) {
	if err := validateOfferingInput(in); err != nil {
		return nil, err
	}
	o := &models.Offering{
		ID: id, UserID: p.UserID, SkillID: in.SkillID, Title: in.Title, Description: in.Description,
		Type: defaultOfferingType(in.Type), Status: defaultOfferingStatus(in.Status),
		Modality: defaultModality(in.Modality), Price: in.Price, Currency: defaultCurrency(in.Currency),
		Comuna: in.Comuna, Region: in.Region,
	}
	updated, err := s.d.Store.UpdateOffering(ctx, o)
	if err != nil {
		return nil, mapErr(err, "offering")
	}
	return updated, nil
}

// SetOfferingStatus changes the status of one of the caller's offerings.
func (s *OpportunityService) SetOfferingStatus(ctx context.Context, p Principal, id string, status models.OfferingStatus) error {
	if !status.Valid() {
		return apperr.InvalidField("status", "must be draft, active or inactive")
	}
	return mapErr(s.d.Store.SetOfferingStatus(ctx, id, p.UserID, status), "offering")
}

// ---- Opportunities ----

// ListOpportunities returns a page of opportunities.
func (s *OpportunityService) ListOpportunities(ctx context.Context, _ Principal, f models.OpportunityFilter) (models.Page[models.Opportunity], error) {
	items, total, err := s.d.Store.ListOpportunities(ctx, f)
	if err != nil {
		return models.Page[models.Opportunity]{}, mapErr(err, "opportunity")
	}
	return models.NewPage(items, total, f.Page), nil
}

// GetOpportunity returns an opportunity by id.
func (s *OpportunityService) GetOpportunity(ctx context.Context, _ Principal, id string) (*models.Opportunity, error) {
	o, err := s.d.Store.GetOpportunity(ctx, id)
	if err != nil {
		return nil, mapErr(err, "opportunity")
	}
	return o, nil
}

// CreateOpportunity publishes an opportunity for an organization.
func (s *OpportunityService) CreateOpportunity(ctx context.Context, p Principal, in OpportunityInput) (*models.Opportunity, error) {
	if err := s.d.authorizeOrgWrite(ctx, p, in.OrganizationID); err != nil {
		return nil, err
	}
	if err := validateOpportunityInput(in); err != nil {
		return nil, err
	}
	o := &models.Opportunity{
		OrganizationID: in.OrganizationID, Title: in.Title, Description: in.Description,
		Type: defaultOpportunityType(in.Type), Status: defaultOpportunityStatus(in.Status),
		Comuna: in.Comuna, Region: in.Region, SalaryMin: in.SalaryMin, SalaryMax: in.SalaryMax,
		Currency: defaultCurrency(in.Currency), ContactEmail: in.ContactEmail,
		CreatedBy: &p.UserID, ExpiresAt: in.ExpiresAt,
	}
	created, err := s.d.Store.CreateOpportunity(ctx, o)
	if err != nil {
		return nil, mapErr(err, "opportunity")
	}
	return created, nil
}

// UpdateOpportunity updates an opportunity.
func (s *OpportunityService) UpdateOpportunity(ctx context.Context, p Principal, id string, in OpportunityInput) (*models.Opportunity, error) {
	existing, err := s.d.Store.GetOpportunity(ctx, id)
	if err != nil {
		return nil, mapErr(err, "opportunity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, existing.OrganizationID); err != nil {
		return nil, err
	}
	if err := validateOpportunityInput(in); err != nil {
		return nil, err
	}
	existing.Title = in.Title
	existing.Description = in.Description
	existing.Type = defaultOpportunityType(in.Type)
	existing.Status = defaultOpportunityStatus(in.Status)
	existing.Comuna = in.Comuna
	existing.Region = in.Region
	existing.SalaryMin = in.SalaryMin
	existing.SalaryMax = in.SalaryMax
	existing.Currency = defaultCurrency(in.Currency)
	existing.ContactEmail = in.ContactEmail
	existing.ExpiresAt = in.ExpiresAt
	updated, err := s.d.Store.UpdateOpportunity(ctx, existing)
	if err != nil {
		return nil, mapErr(err, "opportunity")
	}
	return updated, nil
}

// SetOpportunityStatus changes the status of an opportunity.
func (s *OpportunityService) SetOpportunityStatus(ctx context.Context, p Principal, id string, status models.OpportunityStatus) error {
	existing, err := s.d.Store.GetOpportunity(ctx, id)
	if err != nil {
		return mapErr(err, "opportunity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, existing.OrganizationID); err != nil {
		return err
	}
	if !status.Valid() {
		return apperr.InvalidField("status", "must be draft, open or closed")
	}
	return mapErr(s.d.Store.SetOpportunityStatus(ctx, id, status), "opportunity")
}

// Apply submits the caller's application to an opportunity.
func (s *OpportunityService) Apply(ctx context.Context, p Principal, oppID, message string) (*models.Application, error) {
	opp, err := s.d.Store.GetOpportunity(ctx, oppID)
	if err != nil {
		return nil, mapErr(err, "opportunity")
	}
	if opp.Status != models.OpportunityStatusOpen {
		return nil, apperr.Conflict("this opportunity is not open")
	}
	app, err := s.d.Store.CreateApplication(ctx, oppID, p.UserID, message)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, apperr.Conflict("you have already applied to this opportunity")
		}
		return nil, mapErr(err, "application")
	}
	return app, nil
}

// ListApplications lists the applications of an opportunity.
func (s *OpportunityService) ListApplications(ctx context.Context, p Principal, oppID string, page models.PageParams) (models.Page[models.Application], error) {
	opp, err := s.d.Store.GetOpportunity(ctx, oppID)
	if err != nil {
		return models.Page[models.Application]{}, mapErr(err, "opportunity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, opp.OrganizationID); err != nil {
		return models.Page[models.Application]{}, err
	}
	items, total, err := s.d.Store.ListApplications(ctx, oppID, page)
	if err != nil {
		return models.Page[models.Application]{}, mapErr(err, "application")
	}
	return models.NewPage(items, total, page), nil
}

// UpdateApplicationStatus accepts or rejects an application.
func (s *OpportunityService) UpdateApplicationStatus(ctx context.Context, p Principal, oppID, appID string, status models.ApplicationStatus) error {
	opp, err := s.d.Store.GetOpportunity(ctx, oppID)
	if err != nil {
		return mapErr(err, "opportunity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, opp.OrganizationID); err != nil {
		return err
	}
	if !status.Valid() {
		return apperr.InvalidField("status", "must be applied, accepted or rejected")
	}
	// The application must belong to the opportunity the caller was authorized
	// against; otherwise a caller could mutate another organization's data.
	app, err := s.d.Store.GetApplicationByID(ctx, appID)
	if err != nil {
		return mapErr(err, "application")
	}
	if app.OpportunityID != oppID {
		return apperr.Forbidden("the application does not belong to this opportunity")
	}
	if err := s.d.Store.UpdateApplicationStatus(ctx, oppID, appID, status); err != nil {
		return mapErr(err, "application")
	}
	if status == models.ApplicationStatusAccepted || status == models.ApplicationStatusRejected {
		title := "Tu postulación fue aceptada"
		if status == models.ApplicationStatusRejected {
			title = "Tu postulación no fue seleccionada"
		}
		s.d.notifyUser(ctx, app.UserID, "application_status", title,
			"Oportunidad: "+opp.Title,
			map[string]any{"opportunity_id": oppID, "application_id": appID, "status": string(status)})
	}
	return nil
}

// MyApplications returns the caller's applications with the opportunity summary.
func (s *OpportunityService) MyApplications(ctx context.Context, p Principal, page models.PageParams) (models.Page[models.Application], error) {
	items, total, err := s.d.Store.ListApplicationsByUser(ctx, p.UserID, page)
	if err != nil {
		return models.Page[models.Application]{}, mapErr(err, "application")
	}
	return models.NewPage(items, total, page), nil
}

// ReportInput is the payload for reporting a suspicious opportunity.
type ReportInput struct {
	Reason  string
	Details string
}

// Report records the caller's report about a suspicious opportunity.
func (s *OpportunityService) Report(ctx context.Context, p Principal, oppID string, in ReportInput) (*models.OpportunityReport, error) {
	if _, err := s.d.Store.GetOpportunity(ctx, oppID); err != nil {
		return nil, mapErr(err, "opportunity")
	}
	var v validate.Errors
	v.Required("reason", in.Reason)
	v.MaxLen("reason", in.Reason, 120)
	v.MaxLen("details", in.Details, 2000)
	if err := v.Err(); err != nil {
		return nil, err
	}
	report, err := s.d.Store.CreateReport(ctx, &models.OpportunityReport{
		OpportunityID: oppID, UserID: p.UserID, Reason: in.Reason, Details: in.Details,
	})
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, apperr.Conflict("you have already reported this opportunity")
		}
		return nil, mapErr(err, "report")
	}
	return report, nil
}

func defaultOfferingType(t models.OfferingType) models.OfferingType {
	if t == "" {
		return models.OfferingTypeClass
	}
	return t
}

func defaultOfferingStatus(s models.OfferingStatus) models.OfferingStatus {
	if s == "" {
		return models.OfferingStatusActive
	}
	return s
}

func defaultOpportunityType(t models.OpportunityType) models.OpportunityType {
	if t == "" {
		return models.OpportunityTypeJob
	}
	return t
}

func defaultOpportunityStatus(s models.OpportunityStatus) models.OpportunityStatus {
	if s == "" {
		return models.OpportunityStatusOpen
	}
	return s
}

func validateOfferingInput(in OfferingInput) error {
	var v validate.Errors
	v.Required("title", in.Title)
	v.MaxLen("title", in.Title, 200)
	v.MaxLen("description", in.Description, 4000)
	v.OneOf("type", string(in.Type), string(models.OfferingTypeClass), string(models.OfferingTypeService))
	v.OneOf("status", string(in.Status), string(models.OfferingStatusDraft), string(models.OfferingStatusActive), string(models.OfferingStatusInactive))
	v.OneOf("modality", string(in.Modality), string(models.ModalityInPerson), string(models.ModalityOnline), string(models.ModalityHybrid))
	if in.Price < 0 {
		v.Add("price", "must not be negative")
	}
	return v.Err()
}

func validateOpportunityInput(in OpportunityInput) error {
	var v validate.Errors
	v.Required("title", in.Title)
	v.MaxLen("title", in.Title, 200)
	v.MaxLen("description", in.Description, 4000)
	v.Email("contact_email", in.ContactEmail)
	v.OneOf("type", string(in.Type),
		string(models.OpportunityTypeJob), string(models.OpportunityTypeGig),
		string(models.OpportunityTypeVolunteer), string(models.OpportunityTypeEconomic))
	v.OneOf("status", string(in.Status),
		string(models.OpportunityStatusDraft), string(models.OpportunityStatusOpen), string(models.OpportunityStatusClosed))
	if in.SalaryMin != nil && in.SalaryMax != nil && *in.SalaryMax < *in.SalaryMin {
		v.Add("salary_max", "must be greater than or equal to salary_min")
	}
	return v.Err()
}
