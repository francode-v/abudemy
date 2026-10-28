package service

import (
	"context"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/validate"
)

// CatalogService manages service establishments and support resources.
type CatalogService struct{ d Deps }

// ServiceInput is the payload for creating or updating a service.
type ServiceInput struct {
	OrganizationID *string
	CategoryID     *string
	Name           string
	Description    string
	Type           models.ServiceType
	Status         models.ServiceStatus
	Address        string
	Comuna         string
	Region         string
	Phone          string
	Email          string
	Website        string
	Latitude       *float64
	Longitude      *float64
	Schedule       []models.ScheduleEntry
}

// List returns a page of services. Authenticated users in any role may browse.
func (s *CatalogService) List(ctx context.Context, _ Principal, f models.ServiceFilter) (models.Page[models.Service], error) {
	services, total, err := s.d.Store.ListServices(ctx, f)
	if err != nil {
		return models.Page[models.Service]{}, mapErr(err, "service")
	}
	return models.NewPage(services, total, f.Page), nil
}

// Get returns a service by id.
func (s *CatalogService) Get(ctx context.Context, _ Principal, id string) (*models.Service, error) {
	svc, err := s.d.Store.GetServiceByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "service")
	}
	return svc, nil
}

// Categories lists the service categories.
func (s *CatalogService) Categories(ctx context.Context) ([]models.ServiceCategory, error) {
	cats, err := s.d.Store.ListCategories(ctx)
	if err != nil {
		return nil, mapErr(err, "service category")
	}
	return cats, nil
}

// Create creates a service owned by an organization (or the platform when
// created by an administrator without an organization).
func (s *CatalogService) Create(ctx context.Context, p Principal, in ServiceInput) (*models.Service, error) {
	if err := s.authorizeWrite(ctx, p, in.OrganizationID); err != nil {
		return nil, err
	}
	if err := validateServiceInput(in); err != nil {
		return nil, err
	}
	svc := &models.Service{
		OrganizationID: in.OrganizationID,
		CategoryID:     in.CategoryID,
		Name:           in.Name,
		Description:    in.Description,
		Type:           defaultServiceType(in.Type),
		Status:         defaultServiceStatus(in.Status),
		Address:        in.Address,
		Comuna:         in.Comuna,
		Region:         in.Region,
		Phone:          in.Phone,
		Email:          in.Email,
		Website:        in.Website,
		Latitude:       in.Latitude,
		Longitude:      in.Longitude,
		CreatedBy:      &p.UserID,
	}
	created, err := s.d.Store.CreateService(ctx, svc)
	if err != nil {
		return nil, mapErr(err, "service")
	}
	if len(in.Schedule) > 0 {
		if err := s.d.Store.ReplaceSchedule(ctx, created.ID, in.Schedule); err != nil {
			return nil, mapErr(err, "service schedule")
		}
	}
	return s.d.Store.GetServiceByID(ctx, created.ID)
}

// Update updates a service.
func (s *CatalogService) Update(ctx context.Context, p Principal, id string, in ServiceInput) (*models.Service, error) {
	existing, err := s.d.Store.GetServiceByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "service")
	}
	if err := s.authorizeWrite(ctx, p, existing.OrganizationID); err != nil {
		return nil, err
	}
	if err := validateServiceInput(in); err != nil {
		return nil, err
	}
	existing.CategoryID = in.CategoryID
	existing.Name = in.Name
	existing.Description = in.Description
	existing.Type = defaultServiceType(in.Type)
	existing.Status = defaultServiceStatus(in.Status)
	existing.Address = in.Address
	existing.Comuna = in.Comuna
	existing.Region = in.Region
	existing.Phone = in.Phone
	existing.Email = in.Email
	existing.Website = in.Website
	existing.Latitude = in.Latitude
	existing.Longitude = in.Longitude

	updated, err := s.d.Store.UpdateService(ctx, existing)
	if err != nil {
		return nil, mapErr(err, "service")
	}
	if in.Schedule != nil {
		if err := s.d.Store.ReplaceSchedule(ctx, id, in.Schedule); err != nil {
			return nil, mapErr(err, "service schedule")
		}
	}
	return s.d.Store.GetServiceByID(ctx, updated.ID)
}

// SetStatus activates or deactivates a service.
func (s *CatalogService) SetStatus(ctx context.Context, p Principal, id string, status models.ServiceStatus) (*models.Service, error) {
	existing, err := s.d.Store.GetServiceByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "service")
	}
	if err := s.authorizeWrite(ctx, p, existing.OrganizationID); err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, apperr.InvalidField("status", "must be active, inactive or draft")
	}
	if err := s.d.Store.SetServiceStatus(ctx, id, status); err != nil {
		return nil, mapErr(err, "service")
	}
	return s.d.Store.GetServiceByID(ctx, id)
}

// Delete removes a service.
func (s *CatalogService) Delete(ctx context.Context, p Principal, id string) error {
	existing, err := s.d.Store.GetServiceByID(ctx, id)
	if err != nil {
		return mapErr(err, "service")
	}
	if err := s.authorizeWrite(ctx, p, existing.OrganizationID); err != nil {
		return err
	}
	return mapErr(s.d.Store.DeleteService(ctx, id), "service")
}

// authorizeWrite allows administrators to manage any service, and organization
// members to manage their own organization's services.
func (s *CatalogService) authorizeWrite(ctx context.Context, p Principal, orgID *string) error {
	if p.IsAdmin() {
		return nil
	}
	if orgID == nil || *orgID == "" {
		return apperr.Forbidden("only administrators may manage platform-level services")
	}
	if !s.d.canWriteOrg(ctx, p, *orgID) {
		return apperr.Forbidden("you do not have write access to this organization")
	}
	return nil
}

func defaultServiceType(t models.ServiceType) models.ServiceType {
	if t == "" {
		return models.ServiceTypeEstablishment
	}
	return t
}

func defaultServiceStatus(s models.ServiceStatus) models.ServiceStatus {
	if s == "" {
		return models.ServiceStatusActive
	}
	return s
}

func validateServiceInput(in ServiceInput) error {
	var v validate.Errors
	v.Required("name", in.Name)
	v.MaxLen("name", in.Name, 160)
	v.MaxLen("description", in.Description, 4000)
	v.Email("email", in.Email)
	v.Latitude("latitude", in.Latitude)
	v.Longitude("longitude", in.Longitude)
	v.OneOf("type", string(in.Type), string(models.ServiceTypeEstablishment), string(models.ServiceTypeResource))
	v.OneOf("status", string(in.Status), string(models.ServiceStatusActive), string(models.ServiceStatusInactive), string(models.ServiceStatusDraft))
	for i, e := range in.Schedule {
		if e.DayOfWeek < 0 || e.DayOfWeek > 6 {
			v.Add("schedule["+itoa(i)+"].day_of_week", "must be between 0 and 6")
		}
		if _, err := time.Parse("15:04", e.OpensAt); err != nil {
			v.Add("schedule["+itoa(i)+"].opens_at", "must use HH:MM format")
		}
		if _, err := time.Parse("15:04", e.ClosesAt); err != nil {
			v.Add("schedule["+itoa(i)+"].closes_at", "must use HH:MM format")
		}
	}
	return v.Err()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
