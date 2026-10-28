package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/store"
	"github.com/abudemy/backend/internal/validate"
)

// UserService manages user profiles and administrative user operations.
type UserService struct{ d Deps }

// UpdateProfileInput is the payload for updating the caller's profile.
// The socio-demographic fields are pointers so that clients which do not send
// them (for example the institutional web portal) do not erase stored values.
type UpdateProfileInput struct {
	FullName        string
	Phone           string
	BirthDate       *time.Time
	Comuna          *string
	HealthSystem    *models.HealthSystem
	Sex             *models.Sex
	LivingSituation *models.LivingSituation
}

// Profile returns the caller's own user record.
func (s *UserService) Profile(ctx context.Context, p Principal) (*models.User, error) {
	user, err := s.d.Store.GetUserByID(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "user")
	}
	return user, nil
}

// UpdateProfile updates the caller's own profile.
func (s *UserService) UpdateProfile(ctx context.Context, p Principal, in UpdateProfileInput) (*models.User, error) {
	var v validate.Errors
	v.Required("full_name", in.FullName)
	v.MinLen("full_name", in.FullName, 2)
	v.MaxLen("full_name", in.FullName, 120)
	v.MaxLen("phone", in.Phone, 40)
	validateBirthDate(&v, "birth_date", in.BirthDate)
	if in.Comuna != nil {
		v.MaxLen("comuna", *in.Comuna, 80)
	}
	if in.HealthSystem != nil {
		validateHealthSystem(&v, "health_system", *in.HealthSystem)
	}
	if in.Sex != nil {
		validateSex(&v, "sex", *in.Sex)
	}
	if in.LivingSituation != nil {
		validateLivingSituation(&v, "living_situation", *in.LivingSituation)
	}
	if err := v.Err(); err != nil {
		return nil, err
	}

	current, err := s.d.Store.GetUserByID(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "user")
	}
	comuna := current.Comuna
	if in.Comuna != nil {
		comuna = strings.TrimSpace(*in.Comuna)
	}
	healthSystem := current.HealthSystem
	if in.HealthSystem != nil {
		healthSystem = *in.HealthSystem
	}
	sex := current.Sex
	if in.Sex != nil {
		sex = *in.Sex
	}
	livingSituation := current.LivingSituation
	if in.LivingSituation != nil {
		livingSituation = *in.LivingSituation
	}

	user, err := s.d.Store.UpdateUserProfile(ctx, p.UserID, store.UpdateUserProfileParams{
		FullName:        strings.TrimSpace(in.FullName),
		Phone:           strings.TrimSpace(in.Phone),
		BirthDate:       in.BirthDate,
		Comuna:          comuna,
		HealthSystem:    healthSystem,
		Sex:             sex,
		LivingSituation: livingSituation,
	})
	if err != nil {
		return nil, mapErr(err, "user")
	}
	return user, nil
}

// Get returns a user. Callers may only read their own record unless they are
// administrators.
func (s *UserService) Get(ctx context.Context, p Principal, id string) (*models.User, error) {
	if id != p.UserID && !p.IsAdmin() {
		return nil, apperr.Forbidden("you may only access your own account")
	}
	user, err := s.d.Store.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "user")
	}
	return user, nil
}

// List returns a page of users. Administrators can query every role; approved
// institutional clients can only consult the community (mobile users).
func (s *UserService) List(ctx context.Context, p Principal, f store.UserFilter) (models.Page[models.User], error) {
	if !p.CanViewRegisteredPeople() {
		return models.Page[models.User]{}, apperr.Forbidden("administrator or approved institutional client role required")
	}
	// Non-administrators must not be able to enumerate staff or administrator
	// accounts through this endpoint: the registered-people view only covers
	// community accounts.
	if !p.IsAdmin() {
		f.Role = string(models.RoleMobileUser)
	}
	var v validate.Errors
	v.OneOf("health_system", f.HealthSystem,
		"fonasa", "isapre", "capredena", "dipreca", "private", "other")
	v.OneOf("sex", f.Sex, "female", "male", "other", "prefer_not_to_say")
	v.OneOf("living_situation", f.LivingSituation,
		"alone", "with_partner", "with_children", "with_partner_and_children",
		"with_family", "with_caregiver", "other")
	if f.AgeMin != nil && (*f.AgeMin < 0 || *f.AgeMin > maxAgeYears) {
		v.Add("age_min", "must be between 0 and 120")
	}
	if f.AgeMax != nil && (*f.AgeMax < 0 || *f.AgeMax > maxAgeYears) {
		v.Add("age_max", "must be between 0 and 120")
	}
	if f.AgeMin != nil && f.AgeMax != nil && *f.AgeMin > *f.AgeMax {
		v.Add("age_max", "must be greater than or equal to age_min")
	}
	if err := v.Err(); err != nil {
		return models.Page[models.User]{}, err
	}
	users, total, err := s.d.Store.ListUsers(ctx, f)
	if err != nil {
		return models.Page[models.User]{}, mapErr(err, "user")
	}
	return models.NewPage(users, total, f.Page), nil
}

// UpdateStatus activates, deactivates or suspends an account. Admin only.
func (s *UserService) UpdateStatus(ctx context.Context, p Principal, id string, status models.UserStatus) (*models.User, error) {
	if !p.IsAdmin() {
		return nil, apperr.Forbidden("administrator role required")
	}
	if !status.Valid() {
		return nil, apperr.InvalidField("status", "must be active, inactive or suspended")
	}
	if err := s.d.Store.UpdateUserStatus(ctx, id, status); err != nil {
		return nil, mapErr(err, "user")
	}
	if status != models.UserStatusActive {
		if err := s.d.Store.RevokeUserSessions(ctx, id); err != nil {
			s.d.Logger.WarnContext(ctx, "failed to revoke sessions for deactivated user", "error", err)
		}
	}
	return s.d.Store.GetUserByID(ctx, id)
}

// UpdateRole changes the platform role of an account. Admin only.
func (s *UserService) UpdateRole(ctx context.Context, p Principal, id string, role models.Role) (*models.User, error) {
	if !p.IsAdmin() {
		return nil, apperr.Forbidden("administrator role required")
	}
	if !role.Valid() {
		return nil, apperr.InvalidField("role", "unknown role")
	}
	if err := s.d.Store.UpdateUserRole(ctx, id, role); err != nil {
		return nil, mapErr(err, "user")
	}
	return s.d.Store.GetUserByID(ctx, id)
}

// SetInstitutionalApproval enables or disables institutional portal access for
// an institutional_client account. Admin only.
func (s *UserService) SetInstitutionalApproval(ctx context.Context, p Principal, id string, approved bool) (*models.User, error) {
	if !p.IsAdmin() {
		return nil, apperr.Forbidden("administrator role required")
	}
	user, err := s.d.Store.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "user")
	}
	if user.Role != models.RoleInstitutionalClient {
		return nil, apperr.InvalidField("id", "only institutional client accounts can be approved")
	}
	if err := s.d.Store.SetUserInstitutionalApproval(ctx, id, approved); err != nil {
		return nil, mapErr(err, "user")
	}
	return s.d.Store.GetUserByID(ctx, id)
}

// PreferencesInput is the payload for updating the caller's preferences.
type PreferencesInput struct {
	Interests     []string
	Accessibility models.AccessibilityPreferences
	Notifications models.NotificationPreferences
}

// Preferences returns the caller's preferences (empty defaults when unset).
func (s *UserService) Preferences(ctx context.Context, p Principal) (*models.UserPreferences, error) {
	prefs, err := s.d.Store.GetPreferences(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "preferences")
	}
	return prefs, nil
}

// UpdatePreferences stores the caller's preferences.
func (s *UserService) UpdatePreferences(ctx context.Context, p Principal, in PreferencesInput) (*models.UserPreferences, error) {
	var v validate.Errors
	if len(in.Interests) > 30 {
		v.Add("interests", "must contain at most 30 items")
	}
	seen := make(map[string]struct{}, len(in.Interests))
	interests := make([]string, 0, len(in.Interests))
	for i, it := range in.Interests {
		it = strings.TrimSpace(it)
		if it == "" {
			v.Add("interests["+itoa(i)+"]", "must not be empty")
			continue
		}
		if len([]rune(it)) > 60 {
			v.Add("interests["+itoa(i)+"]", "must be at most 60 characters")
			continue
		}
		key := strings.ToLower(it)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		interests = append(interests, it)
	}
	if err := v.Err(); err != nil {
		return nil, err
	}
	prefs, err := s.d.Store.UpsertPreferences(ctx, p.UserID, &models.UserPreferences{
		Interests:     interests,
		Accessibility: in.Accessibility,
		Notifications: in.Notifications,
	})
	if err != nil {
		return nil, mapErr(err, "preferences")
	}
	return prefs, nil
}

// ChangeEmailInput is the payload for changing the caller's email.
type ChangeEmailInput struct {
	NewEmail string
	Password string
}

// ChangeEmail changes the caller's email after verifying the current password.
func (s *UserService) ChangeEmail(ctx context.Context, p Principal, in ChangeEmailInput) (*models.User, error) {
	var v validate.Errors
	v.Required("new_email", in.NewEmail)
	v.Email("new_email", in.NewEmail)
	v.MaxLen("new_email", in.NewEmail, 254)
	v.Required("password", in.Password)
	if err := v.Err(); err != nil {
		return nil, err
	}
	_, hash, err := s.d.Store.GetUserWithHashByID(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "user")
	}
	if !s.d.Hasher.Verify(hash, in.Password) {
		return nil, apperr.Unauthorized("the password is incorrect")
	}
	user, err := s.d.Store.UpdateUserEmail(ctx, p.UserID, auth.NormalizeEmail(in.NewEmail))
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, apperr.Conflict("an account with that email already exists")
		}
		return nil, mapErr(err, "user")
	}
	return user, nil
}
