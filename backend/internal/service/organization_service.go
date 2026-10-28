package service

import (
	"context"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/store"
	"github.com/abudemy/backend/internal/validate"
)

// OrganizationService manages institutions and their membership.
type OrganizationService struct{ d Deps }

// OrganizationInput is the payload for creating or updating an organization.
type OrganizationInput struct {
	Name        string
	Type        models.OrganizationType
	LegalID     string
	Description string
	Email       string
	Phone       string
	Address     string
	Comuna      string
	Region      string
	Latitude    *float64
	Longitude   *float64
	Active      *bool
}

// AddMemberInput identifies a user to add to an organization.
type AddMemberInput struct {
	UserID string
	Email  string
	Role   models.OrgMemberRole
}

// Create creates an organization. Administrators and approved institutional
// clients may create; the creator becomes the owner.
func (s *OrganizationService) Create(ctx context.Context, p Principal, in OrganizationInput) (*models.Organization, error) {
	if !p.CanCreateOrganization() {
		return nil, apperr.Forbidden("an approved institutional account is required to create an organization")
	}
	if err := validateOrgInput(in); err != nil {
		return nil, err
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	o := &models.Organization{
		Name: in.Name, Type: in.Type, LegalID: in.LegalID, Description: in.Description,
		Email: in.Email, Phone: in.Phone, Address: in.Address, Comuna: in.Comuna, Region: in.Region,
		Latitude: in.Latitude, Longitude: in.Longitude, Active: active,
	}
	created, err := s.d.Store.CreateOrganization(ctx, o)
	if err != nil {
		return nil, mapErr(err, "organization")
	}
	if _, err := s.d.Store.AddMember(ctx, created.ID, p.UserID, models.OrgRoleOwner); err != nil {
		return nil, mapErr(err, "organization member")
	}
	return created, nil
}

// Get returns an organization the caller may read.
func (s *OrganizationService) Get(ctx context.Context, p Principal, id string) (*models.Organization, error) {
	if !s.d.canReadOrg(ctx, p, id) {
		return nil, apperr.Forbidden("you do not have access to this organization")
	}
	o, err := s.d.Store.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "organization")
	}
	return o, nil
}

// Update updates an organization the caller may write to.
func (s *OrganizationService) Update(ctx context.Context, p Principal, id string, in OrganizationInput) (*models.Organization, error) {
	if !s.d.canWriteOrg(ctx, p, id) {
		return nil, apperr.Forbidden("you do not have write access to this organization")
	}
	if err := validateOrgInput(in); err != nil {
		return nil, err
	}
	current, err := s.d.Store.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "organization")
	}
	current.Name = in.Name
	current.Type = in.Type
	current.LegalID = in.LegalID
	current.Description = in.Description
	current.Email = in.Email
	current.Phone = in.Phone
	current.Address = in.Address
	current.Comuna = in.Comuna
	current.Region = in.Region
	current.Latitude = in.Latitude
	current.Longitude = in.Longitude
	if in.Active != nil {
		current.Active = *in.Active
	}
	updated, err := s.d.Store.UpdateOrganization(ctx, current)
	if err != nil {
		return nil, mapErr(err, "organization")
	}
	return updated, nil
}

// List returns organizations. Administrators see everything; other callers see
// only the organizations they belong to.
func (s *OrganizationService) List(ctx context.Context, p Principal, f store.OrganizationFilter) (models.Page[models.Organization], error) {
	if p.Role == models.RoleInstitutionalClient && !p.InstitutionalApproved {
		return models.NewPage([]models.Organization{}, 0, f.Page), nil
	}
	if !p.IsAdmin() {
		orgs, err := s.d.Store.ListUserOrganizations(ctx, p.UserID)
		if err != nil {
			return models.Page[models.Organization]{}, mapErr(err, "organization")
		}
		items := make([]models.Organization, 0, len(orgs))
		for _, o := range orgs {
			items = append(items, o.Organization)
		}
		return models.NewPage(items, len(items), f.Page), nil
	}
	orgs, total, err := s.d.Store.ListOrganizations(ctx, f)
	if err != nil {
		return models.Page[models.Organization]{}, mapErr(err, "organization")
	}
	return models.NewPage(orgs, total, f.Page), nil
}

// ListMine lists the organizations the caller belongs to with their role.
func (s *OrganizationService) ListMine(ctx context.Context, p Principal) ([]store.UserOrganization, error) {
	if p.Role == models.RoleInstitutionalClient && !p.InstitutionalApproved {
		return []store.UserOrganization{}, nil
	}
	orgs, err := s.d.Store.ListUserOrganizations(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "organization")
	}
	return orgs, nil
}

// ListMembers lists the members of an organization.
func (s *OrganizationService) ListMembers(ctx context.Context, p Principal, orgID string) ([]models.OrganizationMember, error) {
	if !s.d.canReadOrg(ctx, p, orgID) {
		return nil, apperr.Forbidden("you do not have access to this organization")
	}
	members, err := s.d.Store.ListMembers(ctx, orgID)
	if err != nil {
		return nil, mapErr(err, "organization member")
	}
	return members, nil
}

// AddMember adds a user to an organization.
func (s *OrganizationService) AddMember(ctx context.Context, p Principal, orgID string, in AddMemberInput) (*models.OrganizationMember, error) {
	if !s.d.canManageOrgMembers(ctx, p, orgID) {
		return nil, apperr.Forbidden("you may not manage members of this organization")
	}
	role := in.Role
	if role == "" {
		role = models.OrgRoleViewer
	}
	if !role.Valid() {
		return nil, apperr.InvalidField("role", "must be owner, admin, editor or viewer")
	}
	userID := in.UserID
	if userID == "" {
		if in.Email == "" {
			return nil, apperr.InvalidField("user_id", "provide a user_id or an email")
		}
		user, _, err := s.d.Store.GetUserWithHash(ctx, auth.NormalizeEmail(in.Email))
		if err != nil {
			return nil, mapErr(err, "user")
		}
		userID = user.ID
	}
	member, err := s.d.Store.AddMember(ctx, orgID, userID, role)
	if err != nil {
		return nil, mapErr(err, "organization member")
	}
	return member, nil
}

// UpdateMemberRole changes a member's role.
func (s *OrganizationService) UpdateMemberRole(ctx context.Context, p Principal, orgID, userID string, role models.OrgMemberRole) error {
	if !s.d.canManageOrgMembers(ctx, p, orgID) {
		return apperr.Forbidden("you may not manage members of this organization")
	}
	if !role.Valid() {
		return apperr.InvalidField("role", "must be owner, admin, editor or viewer")
	}
	return mapErr(s.d.Store.UpdateMemberRole(ctx, orgID, userID, role), "organization member")
}

// RemoveMember removes a member from an organization.
func (s *OrganizationService) RemoveMember(ctx context.Context, p Principal, orgID, userID string) error {
	if !s.d.canManageOrgMembers(ctx, p, orgID) {
		return apperr.Forbidden("you may not manage members of this organization")
	}
	if userID == p.UserID {
		return apperr.Conflict("you cannot remove yourself from the organization")
	}
	return mapErr(s.d.Store.RemoveMember(ctx, orgID, userID), "organization member")
}

func validateOrgInput(in OrganizationInput) error {
	var v validate.Errors
	v.Required("name", in.Name)
	v.MaxLen("name", in.Name, 160)
	v.MaxLen("description", in.Description, 2000)
	v.Email("email", in.Email)
	v.MaxLen("phone", in.Phone, 40)
	v.MaxLen("comuna", in.Comuna, 120)
	v.MaxLen("region", in.Region, 120)
	v.OneOf("type", string(in.Type),
		string(models.OrgTypeHealth), string(models.OrgTypeCommunity), string(models.OrgTypeNGO),
		string(models.OrgTypeEmployer), string(models.OrgTypeGovernment), string(models.OrgTypeEducational),
		string(models.OrgTypeOther))
	v.Latitude("latitude", in.Latitude)
	v.Longitude("longitude", in.Longitude)
	return v.Err()
}
