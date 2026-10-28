package models

import "time"

// AccessibilityPreferences holds the caller's device/display preferences. They
// are stored server-side so they can follow the user across devices.
type AccessibilityPreferences struct {
	LargeText    bool `json:"large_text"`
	HighContrast bool `json:"high_contrast"`
	ReduceMotion bool `json:"reduce_motion"`
	ScreenReader bool `json:"screen_reader"`
}

// NotificationPreferences holds the caller's notification opt-ins.
type NotificationPreferences struct {
	ActivityReminders bool `json:"activity_reminders"`
	Assistance        bool `json:"assistance"`
	Opportunities     bool `json:"opportunities"`
	Courses           bool `json:"courses"`
}

// UserPreferences is the caller's preferences document.
type UserPreferences struct {
	Interests     []string                 `json:"interests"`
	Accessibility AccessibilityPreferences `json:"accessibility"`
	Notifications NotificationPreferences  `json:"notifications"`
	CreatedAt     *time.Time               `json:"created_at,omitempty"`
	UpdatedAt     *time.Time               `json:"updated_at,omitempty"`
}

// DevicePlatform identifies the operating system of a registered push device.
type DevicePlatform string

const (
	DevicePlatformIOS     DevicePlatform = "ios"
	DevicePlatformAndroid DevicePlatform = "android"
	DevicePlatformWeb     DevicePlatform = "web"
)

// Valid reports whether the platform is known.
func (p DevicePlatform) Valid() bool {
	switch p {
	case DevicePlatformIOS, DevicePlatformAndroid, DevicePlatformWeb:
		return true
	default:
		return false
	}
}

// Device is a push notification target registered by a user.
type Device struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	Platform   DevicePlatform `json:"platform"`
	Token      string         `json:"token"`
	DeviceName string         `json:"device_name,omitempty"`
	AppVersion string         `json:"app_version,omitempty"`
	LastSeenAt time.Time      `json:"last_seen_at"`
	RevokedAt  *time.Time     `json:"revoked_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Notification is an in-app notification for a user. PushStatus records the
// honest result of the delivery attempt against the configured push provider.
type Notification struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	Kind         string             `json:"kind"`
	Title        string             `json:"title"`
	Body         string             `json:"body,omitempty"`
	Data         map[string]any     `json:"data,omitempty"`
	ReadAt       *time.Time         `json:"read_at,omitempty"`
	PushStatus   NotificationStatus `json:"push_status"`
	PushProvider string             `json:"push_provider"`
	PushError    string             `json:"push_error,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

// ReminderStatus represents the lifecycle of an activity reminder.
type ReminderStatus string

const (
	ReminderStatusPending   ReminderStatus = "pending"
	ReminderStatusCancelled ReminderStatus = "cancelled"
	ReminderStatusSent      ReminderStatus = "sent"
	ReminderStatusSkipped   ReminderStatus = "skipped"
)

// Valid reports whether the reminder status is known.
func (s ReminderStatus) Valid() bool {
	switch s {
	case ReminderStatusPending, ReminderStatusCancelled, ReminderStatusSent, ReminderStatusSkipped:
		return true
	default:
		return false
	}
}

// ActivityReminder is a reminder a user sets for an activity.
type ActivityReminder struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	ActivityID string         `json:"activity_id"`
	RemindAt   time.Time      `json:"remind_at"`
	Status     ReminderStatus `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	// Expanded fields.
	ActivityTitle string `json:"activity_title,omitempty"`
}

// ReportStatus represents the moderation state of a report.
type ReportStatus string

const (
	ReportStatusOpen      ReportStatus = "open"
	ReportStatusReviewing ReportStatus = "reviewing"
	ReportStatusDismissed ReportStatus = "dismissed"
	ReportStatusActioned  ReportStatus = "actioned"
)

// Valid reports whether the report status is known.
func (s ReportStatus) Valid() bool {
	switch s {
	case ReportStatusOpen, ReportStatusReviewing, ReportStatusDismissed, ReportStatusActioned:
		return true
	default:
		return false
	}
}

// OpportunityReport is a user report about a suspicious opportunity post.
type OpportunityReport struct {
	ID            string       `json:"id"`
	OpportunityID string       `json:"opportunity_id"`
	UserID        string       `json:"user_id"`
	Reason        string       `json:"reason"`
	Details       string       `json:"details,omitempty"`
	Status        ReportStatus `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}
