package models

import "time"

// ServiceType distinguishes an establishment from a support resource.
type ServiceType string

const (
	ServiceTypeEstablishment ServiceType = "establishment"
	ServiceTypeResource      ServiceType = "resource"
)

// Valid reports whether the service type is known.
func (t ServiceType) Valid() bool {
	return t == ServiceTypeEstablishment || t == ServiceTypeResource
}

// ServiceStatus represents the lifecycle state of a service.
type ServiceStatus string

const (
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusInactive ServiceStatus = "inactive"
	ServiceStatusDraft    ServiceStatus = "draft"
)

// Valid reports whether the service status is known.
func (s ServiceStatus) Valid() bool {
	return s == ServiceStatusActive || s == ServiceStatusInactive || s == ServiceStatusDraft
}

// ServiceCategory groups services into browsable categories.
type ServiceCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
}

// ScheduleEntry is one weekly opening window for a service.
type ScheduleEntry struct {
	ID        string `json:"id,omitempty"`
	DayOfWeek int    `json:"day_of_week"` // 0=Sunday .. 6=Saturday
	OpensAt   string `json:"opens_at"`    // HH:MM
	ClosesAt  string `json:"closes_at"`   // HH:MM
}

// Service is an establishment or support resource offered on the platform.
type Service struct {
	ID               string           `json:"id"`
	OrganizationID   *string          `json:"organization_id,omitempty"`
	OrganizationName string           `json:"organization_name,omitempty"`
	CategoryID       *string          `json:"category_id,omitempty"`
	Category         *ServiceCategory `json:"category,omitempty"`
	Name             string           `json:"name"`
	Description      string           `json:"description,omitempty"`
	Type             ServiceType      `json:"type"`
	Status           ServiceStatus    `json:"status"`
	Address          string           `json:"address,omitempty"`
	Comuna           string           `json:"comuna,omitempty"`
	Region           string           `json:"region,omitempty"`
	Latitude         *float64         `json:"latitude,omitempty"`
	Longitude        *float64         `json:"longitude,omitempty"`
	Phone            string           `json:"phone,omitempty"`
	Email            string           `json:"email,omitempty"`
	Website          string           `json:"website,omitempty"`
	Schedule         []ScheduleEntry  `json:"schedule"`
	CreatedBy        *string          `json:"created_by,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// ServiceFilter describes the supported search parameters for services.
type ServiceFilter struct {
	CategoryID   string
	Type         string
	Status       string
	Comuna       string
	Region       string
	Search       string
	Organization string
	NearLat      *float64
	NearLng      *float64
	RadiusKM     *float64
	Page         PageParams
}
