package models

import "time"

// OrganizationType classifies an organization.
type OrganizationType string

const (
	OrgTypeHealth      OrganizationType = "health"
	OrgTypeCommunity   OrganizationType = "community"
	OrgTypeNGO         OrganizationType = "ngo"
	OrgTypeEmployer    OrganizationType = "employer"
	OrgTypeGovernment  OrganizationType = "government"
	OrgTypeEducational OrganizationType = "educational"
	OrgTypeOther       OrganizationType = "other"
)

// Valid reports whether the organization type is known.
func (t OrganizationType) Valid() bool {
	switch t {
	case OrgTypeHealth, OrgTypeCommunity, OrgTypeNGO, OrgTypeEmployer,
		OrgTypeGovernment, OrgTypeEducational, OrgTypeOther:
		return true
	default:
		return false
	}
}

// OrgMemberRole is the role of a user inside an organization.
type OrgMemberRole string

const (
	OrgRoleOwner  OrgMemberRole = "owner"
	OrgRoleAdmin  OrgMemberRole = "admin"
	OrgRoleEditor OrgMemberRole = "editor"
	OrgRoleViewer OrgMemberRole = "viewer"
)

// Valid reports whether the organization member role is known.
func (r OrgMemberRole) Valid() bool {
	switch r {
	case OrgRoleOwner, OrgRoleAdmin, OrgRoleEditor, OrgRoleViewer:
		return true
	default:
		return false
	}
}

// CanWrite reports whether the role may create or modify organization data.
func (r OrgMemberRole) CanWrite() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin || r == OrgRoleEditor
}

// CanManageMembers reports whether the role may manage organization members.
func (r OrgMemberRole) CanManageMembers() bool {
	return r == OrgRoleOwner || r == OrgRoleAdmin
}

// Organization is an institution providing or coordinating services.
type Organization struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        OrganizationType `json:"type"`
	LegalID     string           `json:"legal_id,omitempty"`
	Description string           `json:"description,omitempty"`
	Email       string           `json:"email,omitempty"`
	Phone       string           `json:"phone,omitempty"`
	Address     string           `json:"address,omitempty"`
	Comuna      string           `json:"comuna,omitempty"`
	Region      string           `json:"region,omitempty"`
	Latitude    *float64         `json:"latitude,omitempty"`
	Longitude   *float64         `json:"longitude,omitempty"`
	Active      bool             `json:"active"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// OrganizationMember links a user to an organization with a role.
type OrganizationMember struct {
	ID             string        `json:"id"`
	OrganizationID string        `json:"organization_id"`
	UserID         string        `json:"user_id"`
	Role           OrgMemberRole `json:"role"`
	CreatedAt      time.Time     `json:"created_at"`
	// Expanded user fields, populated on demand.
	UserEmail    string `json:"user_email,omitempty"`
	UserFullName string `json:"user_full_name,omitempty"`
}
