package models

import "time"

// ActivityType classifies activities.
type ActivityType string

const (
	ActivityTypeWorkshop ActivityType = "workshop"
	ActivityTypeEvent    ActivityType = "event"
	ActivityTypeCourse   ActivityType = "course"
	ActivityTypeSupport  ActivityType = "support_group"
)

// Valid reports whether the activity type is known.
func (t ActivityType) Valid() bool {
	switch t {
	case ActivityTypeWorkshop, ActivityTypeEvent, ActivityTypeCourse, ActivityTypeSupport:
		return true
	default:
		return false
	}
}

// Modality describes how an activity is delivered.
type Modality string

const (
	ModalityInPerson Modality = "in_person"
	ModalityOnline   Modality = "online"
	ModalityHybrid   Modality = "hybrid"
)

// Valid reports whether the modality is known.
func (m Modality) Valid() bool {
	return m == ModalityInPerson || m == ModalityOnline || m == ModalityHybrid
}

// ActivityStatus represents the lifecycle state of an activity.
type ActivityStatus string

const (
	ActivityStatusDraft     ActivityStatus = "draft"
	ActivityStatusPublished ActivityStatus = "published"
	ActivityStatusCancelled ActivityStatus = "cancelled"
	ActivityStatusFinished  ActivityStatus = "finished"
)

// Valid reports whether the activity status is known.
func (s ActivityStatus) Valid() bool {
	switch s {
	case ActivityStatusDraft, ActivityStatusPublished, ActivityStatusCancelled, ActivityStatusFinished:
		return true
	default:
		return false
	}
}

// Activity is a workshop, event or course.
type Activity struct {
	ID             string         `json:"id"`
	OrganizationID *string        `json:"organization_id,omitempty"`
	ServiceID      *string        `json:"service_id,omitempty"`
	CategoryID     *string        `json:"category_id,omitempty"`
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	Type           ActivityType   `json:"type"`
	Modality       Modality       `json:"modality"`
	Status         ActivityStatus `json:"status"`
	Location       string         `json:"location,omitempty"`
	Comuna         string         `json:"comuna,omitempty"`
	Region         string         `json:"region,omitempty"`
	Latitude       *float64       `json:"latitude,omitempty"`
	Longitude      *float64       `json:"longitude,omitempty"`
	StartAt        time.Time      `json:"start_at"`
	EndAt          *time.Time     `json:"end_at,omitempty"`
	Capacity       int            `json:"capacity"`
	EnrolledCount  int            `json:"enrolled_count"`
	Price          float64        `json:"price"`
	Currency       string         `json:"currency"`
	CreatedBy      *string        `json:"created_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// EnrollmentStatus represents the state of an enrollment.
type EnrollmentStatus string

const (
	EnrollmentStatusEnrolled  EnrollmentStatus = "enrolled"
	EnrollmentStatusWaitlist  EnrollmentStatus = "waitlist"
	EnrollmentStatusCancelled EnrollmentStatus = "cancelled"
	EnrollmentStatusAttended  EnrollmentStatus = "attended"
)

// Valid reports whether the enrollment status is known.
func (s EnrollmentStatus) Valid() bool {
	switch s {
	case EnrollmentStatusEnrolled, EnrollmentStatusWaitlist, EnrollmentStatusCancelled, EnrollmentStatusAttended:
		return true
	default:
		return false
	}
}

// Enrollment links a user to an activity.
type Enrollment struct {
	ID         string           `json:"id"`
	ActivityID string           `json:"activity_id"`
	UserID     string           `json:"user_id"`
	Status     EnrollmentStatus `json:"status"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	// Expanded fields.
	ActivityTitle string `json:"activity_title,omitempty"`
	UserEmail     string `json:"user_email,omitempty"`
	UserFullName  string `json:"user_full_name,omitempty"`
}

// ActivityFilter describes the supported search parameters for activities.
type ActivityFilter struct {
	Type         string
	Status       string
	Modality     string
	Comuna       string
	Region       string
	Search       string
	Organization string
	From         *time.Time
	To           *time.Time
	Page         PageParams
}
