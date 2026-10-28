package models

import "time"

// NotificationStatus represents the delivery state of a notification.
type NotificationStatus string

const (
	// NotificationPending means the notification was queued but not yet attempted.
	NotificationPending NotificationStatus = "pending"
	// NotificationSent means a configured provider accepted the notification.
	NotificationSent NotificationStatus = "sent"
	// NotificationFailed means a configured provider rejected the notification.
	NotificationFailed NotificationStatus = "failed"
	// NotificationSkipped means no provider was configured, so nothing was sent.
	NotificationSkipped NotificationStatus = "skipped"
)

// TrustedContact is an emergency contact declared by a user.
type TrustedContact struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship,omitempty"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email,omitempty"`
	Priority     int       `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AssistanceStatus represents the lifecycle of an assistance request.
type AssistanceStatus string

const (
	AssistanceStatusPending   AssistanceStatus = "pending"
	AssistanceStatusNotified  AssistanceStatus = "notified"
	AssistanceStatusResolved  AssistanceStatus = "resolved"
	AssistanceStatusCancelled AssistanceStatus = "cancelled"
)

// Valid reports whether the assistance status is known.
func (s AssistanceStatus) Valid() bool {
	switch s {
	case AssistanceStatusPending, AssistanceStatusNotified, AssistanceStatusResolved, AssistanceStatusCancelled:
		return true
	default:
		return false
	}
}

// AssistanceRequest is an emergency or support request from a user.
type AssistanceRequest struct {
	ID          string           `json:"id"`
	UserID      string           `json:"user_id"`
	Type        string           `json:"type"`
	Description string           `json:"description,omitempty"`
	Latitude    *float64         `json:"latitude,omitempty"`
	Longitude   *float64         `json:"longitude,omitempty"`
	Status      AssistanceStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	ResolvedAt  *time.Time       `json:"resolved_at,omitempty"`
	// Expanded data.
	Notifications []AssistanceNotification `json:"notifications,omitempty"`
}

// AssistanceNotification records the delivery attempt for one recipient.
type AssistanceNotification struct {
	ID                  string             `json:"id"`
	AssistanceRequestID string             `json:"assistance_request_id"`
	ContactID           *string            `json:"contact_id,omitempty"`
	Channel             string             `json:"channel"`
	Recipient           string             `json:"recipient"`
	Provider            string             `json:"provider"`
	Status              NotificationStatus `json:"status"`
	Error               string             `json:"error,omitempty"`
	SentAt              *time.Time         `json:"sent_at,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
}
