package models

import "time"

// SkillLevel describes proficiency in a skill.
type SkillLevel string

const (
	SkillLevelBeginner     SkillLevel = "beginner"
	SkillLevelIntermediate SkillLevel = "intermediate"
	SkillLevelAdvanced     SkillLevel = "advanced"
	SkillLevelExpert       SkillLevel = "expert"
)

// Valid reports whether the skill level is known.
func (l SkillLevel) Valid() bool {
	switch l {
	case SkillLevelBeginner, SkillLevelIntermediate, SkillLevelAdvanced, SkillLevelExpert:
		return true
	default:
		return false
	}
}

// Skill is a capability declared by a user.
type Skill struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Level       SkillLevel `json:"level"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// OfferingType classifies a user offering.
type OfferingType string

const (
	OfferingTypeClass   OfferingType = "class"
	OfferingTypeService OfferingType = "service"
)

// Valid reports whether the offering type is known.
func (t OfferingType) Valid() bool {
	return t == OfferingTypeClass || t == OfferingTypeService
}

// OfferingStatus represents the lifecycle state of an offering.
type OfferingStatus string

const (
	OfferingStatusDraft    OfferingStatus = "draft"
	OfferingStatusActive   OfferingStatus = "active"
	OfferingStatusInactive OfferingStatus = "inactive"
)

// Valid reports whether the offering status is known.
func (s OfferingStatus) Valid() bool {
	return s == OfferingStatusDraft || s == OfferingStatusActive || s == OfferingStatusInactive
}

// Offering is a class or service published by a user to generate income.
type Offering struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	SkillID     *string        `json:"skill_id,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Type        OfferingType   `json:"type"`
	Status      OfferingStatus `json:"status"`
	Modality    Modality       `json:"modality"`
	Price       float64        `json:"price"`
	Currency    string         `json:"currency"`
	Comuna      string         `json:"comuna,omitempty"`
	Region      string         `json:"region,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// OpportunityType classifies an economic or labour opportunity.
type OpportunityType string

const (
	OpportunityTypeJob       OpportunityType = "job"
	OpportunityTypeGig       OpportunityType = "gig"
	OpportunityTypeVolunteer OpportunityType = "volunteer"
	OpportunityTypeEconomic  OpportunityType = "economic"
)

// Valid reports whether the opportunity type is known.
func (t OpportunityType) Valid() bool {
	switch t {
	case OpportunityTypeJob, OpportunityTypeGig, OpportunityTypeVolunteer, OpportunityTypeEconomic:
		return true
	default:
		return false
	}
}

// OpportunityStatus represents the lifecycle state of an opportunity.
type OpportunityStatus string

const (
	OpportunityStatusDraft  OpportunityStatus = "draft"
	OpportunityStatusOpen   OpportunityStatus = "open"
	OpportunityStatusClosed OpportunityStatus = "closed"
)

// Valid reports whether the opportunity status is known.
func (s OpportunityStatus) Valid() bool {
	return s == OpportunityStatusDraft || s == OpportunityStatusOpen || s == OpportunityStatusClosed
}

// Opportunity is a job or economic opportunity published on the platform.
type Opportunity struct {
	ID             string            `json:"id"`
	OrganizationID *string           `json:"organization_id,omitempty"`
	Title          string            `json:"title"`
	Description    string            `json:"description,omitempty"`
	Type           OpportunityType   `json:"type"`
	Status         OpportunityStatus `json:"status"`
	Comuna         string            `json:"comuna,omitempty"`
	Region         string            `json:"region,omitempty"`
	SalaryMin      *float64          `json:"salary_min,omitempty"`
	SalaryMax      *float64          `json:"salary_max,omitempty"`
	Currency       string            `json:"currency"`
	CreatedBy      *string           `json:"created_by,omitempty"`
	ContactEmail   string            `json:"contact_email,omitempty"`
	ExpiresAt      *time.Time        `json:"expires_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ApplicationStatus represents the state of an application.
type ApplicationStatus string

const (
	ApplicationStatusApplied  ApplicationStatus = "applied"
	ApplicationStatusAccepted ApplicationStatus = "accepted"
	ApplicationStatusRejected ApplicationStatus = "rejected"
)

// Valid reports whether the application status is known.
func (s ApplicationStatus) Valid() bool {
	return s == ApplicationStatusApplied || s == ApplicationStatusAccepted || s == ApplicationStatusRejected
}

// Application links a user to an opportunity.
type Application struct {
	ID            string            `json:"id"`
	OpportunityID string            `json:"opportunity_id"`
	UserID        string            `json:"user_id"`
	Message       string            `json:"message,omitempty"`
	Status        ApplicationStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	// Expanded fields, populated when listing a user's own applications.
	OpportunityTitle  string            `json:"opportunity_title,omitempty"`
	OpportunityStatus OpportunityStatus `json:"opportunity_status,omitempty"`
	OpportunityType   OpportunityType   `json:"opportunity_type,omitempty"`
	Comuna            string            `json:"comuna,omitempty"`
	Region            string            `json:"region,omitempty"`
}

// OpportunityFilter describes the supported search parameters for opportunities.
type OpportunityFilter struct {
	Type   string
	Status string
	Comuna string
	Region string
	Search string
	Page   PageParams
}
