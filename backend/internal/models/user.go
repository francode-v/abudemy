package models

import (
	"encoding/json"
	"time"
)

// Role is the platform-wide role assigned to a user account.
type Role string

const (
	// RoleMobileUser is an older adult using the mobile application.
	RoleMobileUser Role = "mobile_user"
	// RoleInstitutionalClient manages an institutional web workspace.
	RoleInstitutionalClient Role = "institutional_client"
	// RoleOrgCollaborator manages resources on behalf of an organization.
	RoleOrgCollaborator Role = "org_collaborator"
	// RoleAdmin is the platform administrator.
	RoleAdmin Role = "admin"
)

// Valid reports whether the role is one of the known roles.
func (r Role) Valid() bool {
	switch r {
	case RoleMobileUser, RoleInstitutionalClient, RoleOrgCollaborator, RoleAdmin:
		return true
	default:
		return false
	}
}

// UserStatus represents the lifecycle state of a user account.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// Valid reports whether the status is known.
func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

// HealthSystem is the health coverage declared by a user.
type HealthSystem string

const (
	HealthSystemFonasa    HealthSystem = "fonasa"
	HealthSystemIsapre    HealthSystem = "isapre"
	HealthSystemCapredena HealthSystem = "capredena"
	HealthSystemDipreca   HealthSystem = "dipreca"
	HealthSystemPrivate   HealthSystem = "private"
	HealthSystemOther     HealthSystem = "other"
)

// Valid reports whether the health system is known. The empty value is allowed
// for accounts created before this field existed.
func (h HealthSystem) Valid() bool {
	switch h {
	case "", HealthSystemFonasa, HealthSystemIsapre, HealthSystemCapredena,
		HealthSystemDipreca, HealthSystemPrivate, HealthSystemOther:
		return true
	default:
		return false
	}
}

// Sex is the sex a user chooses to declare.
type Sex string

const (
	SexFemale         Sex = "female"
	SexMale           Sex = "male"
	SexOther          Sex = "other"
	SexPreferNotToSay Sex = "prefer_not_to_say"
)

// Valid reports whether the sex value is known.
func (s Sex) Valid() bool {
	switch s {
	case "", SexFemale, SexMale, SexOther, SexPreferNotToSay:
		return true
	default:
		return false
	}
}

// LivingSituation describes who a user lives with.
type LivingSituation string

const (
	LivingSituationAlone                  LivingSituation = "alone"
	LivingSituationWithPartner            LivingSituation = "with_partner"
	LivingSituationWithChildren           LivingSituation = "with_children"
	LivingSituationWithPartnerAndChildren LivingSituation = "with_partner_and_children"
	LivingSituationWithFamily             LivingSituation = "with_family"
	LivingSituationWithCaregiver          LivingSituation = "with_caregiver"
	LivingSituationOther                  LivingSituation = "other"
)

// Valid reports whether the living situation is known.
func (l LivingSituation) Valid() bool {
	switch l {
	case "", LivingSituationAlone, LivingSituationWithPartner, LivingSituationWithChildren,
		LivingSituationWithPartnerAndChildren, LivingSituationWithFamily,
		LivingSituationWithCaregiver, LivingSituationOther:
		return true
	default:
		return false
	}
}

// User is a platform account.
type User struct {
	ID              string          `json:"id"`
	Email           string          `json:"email"`
	FullName        string          `json:"full_name"`
	Phone           string          `json:"phone,omitempty"`
	BirthDate       *time.Time      `json:"birth_date,omitempty"`
	Comuna          string          `json:"comuna,omitempty"`
	HealthSystem    HealthSystem    `json:"health_system,omitempty"`
	Sex             Sex             `json:"sex,omitempty"`
	LivingSituation LivingSituation `json:"living_situation,omitempty"`
	Role            Role            `json:"role"`
	Status          UserStatus      `json:"status"`
	// InstitutionalApproved controls whether an institutional_client account
	// may use the institutional portal. It is always true for other roles.
	InstitutionalApproved bool       `json:"institutional_approved"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	LastLoginAt           *time.Time `json:"last_login_at,omitempty"`
}

// MarshalJSON renders `birth_date` as a calendar date (`YYYY-MM-DD`), matching
// the `format: date` contract declared in docs/openapi.yaml. A bare
// `*time.Time` would serialize with a time and zone, which clients then
// misinterpret (or shift by a day when they convert from UTC).
func (u User) MarshalJSON() ([]byte, error) {
	type userAlias User // alias without this method, to avoid recursing
	var birthDate *string
	if u.BirthDate != nil {
		value := u.BirthDate.Format(time.DateOnly)
		birthDate = &value
	}
	return json.Marshal(struct {
		userAlias
		BirthDate *string `json:"birth_date,omitempty"`
	}{userAlias: userAlias(u), BirthDate: birthDate})
}

// Session is a tracked authentication session (refresh token).
type Session struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	UserAgent string     `json:"user_agent,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	Current   bool       `json:"current,omitempty"`
}
