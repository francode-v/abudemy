// Package service holds the business logic. Handlers translate HTTP to service
// calls; services enforce authorization and orchestrate the repository layer.
package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/assistant"
	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/geo"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/notify"
	"github.com/abudemy/backend/internal/store"
)

// Principal is the authenticated caller extracted from the access token.
type Principal struct {
	UserID    string
	Role      models.Role
	SessionID string
	// InstitutionalApproved reflects the persisted approval flag. It only
	// applies to institutional_client accounts.
	InstitutionalApproved bool
}

// IsAdmin reports whether the caller is a platform administrator.
func (p Principal) IsAdmin() bool { return p.Role == models.RoleAdmin }

// IsInstitutional reports whether the caller may use the institutional web
// platform features.
func (p Principal) IsInstitutional() bool {
	switch p.Role {
	case models.RoleAdmin, models.RoleOrgCollaborator:
		return true
	case models.RoleInstitutionalClient:
		return p.InstitutionalApproved
	default:
		return false
	}
}

// CanCreateOrganization reports whether the caller may create organizations.
// Only administrators and approved institutional clients may do so.
func (p Principal) CanCreateOrganization() bool {
	return p.IsAdmin() || (p.Role == models.RoleInstitutionalClient && p.InstitutionalApproved)
}

// CanViewRegisteredPeople reports whether the caller may consult the community's
// registered people (mobile users). Administrators always can; approved
// institutional clients can too, in read-only mode.
func (p Principal) CanViewRegisteredPeople() bool {
	return p.IsAdmin() || (p.Role == models.RoleInstitutionalClient && p.InstitutionalApproved)
}

// Deps holds the shared dependencies of every service.
type Deps struct {
	Store    *store.Store
	Tokens   *auth.TokenManager
	Hasher   *auth.PasswordHasher
	Notifier notify.Provider
	Logger   *slog.Logger

	Assistant        assistant.Provider
	Geo              geo.Provider
	PaymentsProvider string

	RefreshTTL       time.Duration
	PasswordResetTTL time.Duration
}

// Services aggregates every business service.
type Services struct {
	Auth          *AuthService
	Users         *UserService
	Organizations *OrganizationService
	Catalog       *CatalogService
	Activities    *ActivityService
	Opportunities *OpportunityService
	Assistance    *AssistanceService
	Analytics     *AnalyticsService
	Notifications *NotificationService
	Assistant     *AssistantService
	Geo           *GeoService
	Payments      *PaymentsService
}

// New wires the services together.
func New(d Deps) *Services {
	return &Services{
		Auth:          &AuthService{d: d},
		Users:         &UserService{d: d},
		Organizations: &OrganizationService{d: d},
		Catalog:       &CatalogService{d: d},
		Activities:    &ActivityService{d: d},
		Opportunities: &OpportunityService{d: d},
		Assistance:    &AssistanceService{d: d},
		Analytics:     &AnalyticsService{d: d},
		Notifications: &NotificationService{d: d},
		Assistant:     &AssistantService{d: d},
		Geo:           &GeoService{d: d},
		Payments:      &PaymentsService{d: d},
	}
}

// mapErr converts repository errors into application errors.
func mapErr(err error, resource string) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, store.ErrNotFound):
		return apperr.NotFound(resource)
	case errors.Is(err, store.ErrConflict):
		return apperr.Conflict("the operation conflicts with an existing record")
	case errors.Is(err, store.ErrForeignKey):
		return apperr.Validation(apperr.FieldError{Field: "reference", Message: "referenced resource does not exist"})
	}
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperr.Internal(err)
}

// orgRole returns the caller's role in an organization, if any. Administrators
// are treated as owners of every organization.
func (d Deps) orgRole(ctx context.Context, p Principal, orgID string) (models.OrgMemberRole, bool) {
	if p.IsAdmin() {
		return models.OrgRoleOwner, true
	}
	// An institutional client whose access was revoked must lose access to
	// organization data immediately, even if the membership row still exists.
	if p.Role == models.RoleInstitutionalClient && !p.InstitutionalApproved {
		return "", false
	}
	if orgID == "" {
		return "", false
	}
	member, err := d.Store.GetMember(ctx, orgID, p.UserID)
	if err != nil {
		return "", false
	}
	return member.Role, true
}

// canWriteOrg reports whether the caller may modify the organization's data.
func (d Deps) canWriteOrg(ctx context.Context, p Principal, orgID string) bool {
	role, ok := d.orgRole(ctx, p, orgID)
	return ok && role.CanWrite()
}

// canReadOrg reports whether the caller may read the organization's data.
func (d Deps) canReadOrg(ctx context.Context, p Principal, orgID string) bool {
	_, ok := d.orgRole(ctx, p, orgID)
	return ok
}

// canManageOrgMembers reports whether the caller may manage organization members.
func (d Deps) canManageOrgMembers(ctx context.Context, p Principal, orgID string) bool {
	role, ok := d.orgRole(ctx, p, orgID)
	return ok && role.CanManageMembers()
}

// authorizeOrgWrite allows administrators to manage any resource, and members
// with a write role to manage their own organization's resources. Resources
// without an organization are platform-level and admin-only.
func (d Deps) authorizeOrgWrite(ctx context.Context, p Principal, orgID *string) error {
	if p.IsAdmin() {
		return nil
	}
	if orgID == nil || *orgID == "" {
		return apperr.Forbidden("only administrators may manage platform-level resources")
	}
	if !d.canWriteOrg(ctx, p, *orgID) {
		return apperr.Forbidden("you do not have write access to this organization")
	}
	return nil
}
