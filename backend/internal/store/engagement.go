package store

import (
	"context"
	"encoding/json"

	"github.com/abudemy/backend/internal/models"
)

// ---- Preferences ----

// GetPreferences returns the caller's preferences, creating empty defaults when
// the user has never stored any.
func (s *Store) GetPreferences(ctx context.Context, userID string) (*models.UserPreferences, error) {
	const q = `
		SELECT interests, accessibility, notifications, created_at, updated_at
		FROM user_preferences WHERE user_id = $1`
	p := &models.UserPreferences{
		Interests: []string{},
	}
	var interests, accessibility, notifications []byte
	err := scanRow(s.q.QueryRow(ctx, q, userID),
		&interests, &accessibility, &notifications, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if isNotFound(err) {
			// No row yet: return defaults without persisting anything.
			return p, nil
		}
		return nil, err
	}
	if err := decodeJSONColumn(interests, &p.Interests); err != nil {
		return nil, err
	}
	if err := decodeJSONColumn(accessibility, &p.Accessibility); err != nil {
		return nil, err
	}
	if err := decodeJSONColumn(notifications, &p.Notifications); err != nil {
		return nil, err
	}
	if p.Interests == nil {
		p.Interests = []string{}
	}
	return p, nil
}

// UpsertPreferences stores the caller's preferences document.
func (s *Store) UpsertPreferences(ctx context.Context, userID string, p *models.UserPreferences) (*models.UserPreferences, error) {
	interests, err := json.Marshal(p.Interests)
	if err != nil {
		return nil, err
	}
	accessibility, err := json.Marshal(p.Accessibility)
	if err != nil {
		return nil, err
	}
	notifications, err := json.Marshal(p.Notifications)
	if err != nil {
		return nil, err
	}
	const q = `
		INSERT INTO user_preferences (user_id, interests, accessibility, notifications)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET interests = EXCLUDED.interests,
		    accessibility = EXCLUDED.accessibility,
		    notifications = EXCLUDED.notifications,
		    updated_at = now()
		RETURNING interests, accessibility, notifications, created_at, updated_at`
	out := &models.UserPreferences{}
	var ri, ra, rn []byte
	err = scanRow(s.q.QueryRow(ctx, q, userID, interests, accessibility, notifications),
		&ri, &ra, &rn, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := decodeJSONColumn(ri, &out.Interests); err != nil {
		return nil, err
	}
	if err := decodeJSONColumn(ra, &out.Accessibility); err != nil {
		return nil, err
	}
	if err := decodeJSONColumn(rn, &out.Notifications); err != nil {
		return nil, err
	}
	if out.Interests == nil {
		out.Interests = []string{}
	}
	return out, nil
}

// ---- Devices ----

const deviceColumns = `id, user_id, platform, token, device_name, app_version, last_seen_at, revoked_at, created_at, updated_at`

func scanDevice(row rowScanner) (*models.Device, error) {
	d := &models.Device{}
	err := scanRow(row, &d.ID, &d.UserID, &d.Platform, &d.Token, &d.DeviceName, &d.AppVersion,
		&d.LastSeenAt, &d.RevokedAt, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// UpsertDevice registers (or refreshes) a push device. The token is unique, so
// re-registering the same token revives it and moves it to the current user.
func (s *Store) UpsertDevice(ctx context.Context, d *models.Device) (*models.Device, error) {
	const q = `
		INSERT INTO devices (user_id, platform, token, device_name, app_version)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token) DO UPDATE
		SET user_id = EXCLUDED.user_id,
		    platform = EXCLUDED.platform,
		    device_name = EXCLUDED.device_name,
		    app_version = EXCLUDED.app_version,
		    last_seen_at = now(),
		    revoked_at = NULL,
		    updated_at = now()
		RETURNING ` + deviceColumns
	return scanDevice(s.q.QueryRow(ctx, q, d.UserID, d.Platform, d.Token, d.DeviceName, d.AppVersion))
}

// ListDevices returns the caller's active devices.
func (s *Store) ListDevices(ctx context.Context, userID string) ([]models.Device, error) {
	const q = `SELECT ` + deviceColumns + ` FROM devices WHERE user_id = $1 AND revoked_at IS NULL ORDER BY created_at DESC`
	rows, err := s.q.Query(ctx, q, userID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Device, 0)
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	return out, TranslateError(rows.Err())
}

// RevokeDevice soft-deletes a device owned by a user.
func (s *Store) RevokeDevice(ctx context.Context, id, userID string) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE devices SET revoked_at = now(), updated_at = now() WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`,
		id, userID)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListActiveDeviceTokens returns the push tokens of a user's active devices.
func (s *Store) ListActiveDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.q.Query(ctx,
		`SELECT token FROM devices WHERE user_id = $1 AND revoked_at IS NULL ORDER BY last_seen_at DESC`, userID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, TranslateError(err)
		}
		out = append(out, token)
	}
	return out, TranslateError(rows.Err())
}

// ---- Notifications ----

const notificationColumns = `id, user_id, kind, title, body, data, read_at, push_status, push_provider, push_error, created_at`

func scanNotification(row rowScanner) (*models.Notification, error) {
	n := &models.Notification{}
	var data []byte
	err := scanRow(row, &n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &data,
		&n.ReadAt, &n.PushStatus, &n.PushProvider, &n.PushError, &n.CreatedAt)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		_ = json.Unmarshal(data, &n.Data)
	}
	return n, nil
}

// CreateUserNotification inserts an in-app notification into the user feed.
func (s *Store) CreateUserNotification(ctx context.Context, n *models.Notification) (*models.Notification, error) {
	data, err := json.Marshal(nonNilData(n.Data))
	if err != nil {
		return nil, err
	}
	status := n.PushStatus
	if status == "" {
		status = models.NotificationSkipped
	}
	provider := n.PushProvider
	if provider == "" {
		provider = "none"
	}
	const q = `
		INSERT INTO notifications (user_id, kind, title, body, data, push_status, push_provider, push_error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING ` + notificationColumns
	return scanNotification(s.q.QueryRow(ctx, q, n.UserID, n.Kind, n.Title, n.Body, data, status, provider, n.PushError))
}

// SetNotificationPushResult records the delivery attempt against the provider.
func (s *Store) SetNotificationPushResult(ctx context.Context, id string, status models.NotificationStatus, provider, errMsg string) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE notifications SET push_status = $2, push_provider = $3, push_error = $4 WHERE id = $1`,
		id, status, provider, errMsg)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListUserNotifications returns a page of the caller's notifications.
func (s *Store) ListUserNotifications(ctx context.Context, userID string, p models.PageParams) ([]models.Notification, int, error) {
	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE user_id = $1`, userID), &total); err != nil {
		return nil, 0, err
	}
	b := &qb{}
	b.add("user_id = %s", userID)
	clause := b.limitOffset(p.PerPage, p.Offset)
	rows, err := s.q.Query(ctx, `SELECT `+notificationColumns+` FROM notifications`+b.where()+` ORDER BY created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *n)
	}
	return out, total, TranslateError(rows.Err())
}

// CountUnreadNotifications returns the number of unread notifications.
func (s *Store) CountUnreadNotifications(ctx context.Context, userID string) (int, error) {
	var n int
	err := scanRow(s.q.QueryRow(ctx,
		`SELECT count(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID), &n)
	return n, TranslateError(err)
}

// MarkNotificationRead marks one notification as read.
func (s *Store) MarkNotificationRead(ctx context.Context, id, userID string) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE notifications SET read_at = now() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`, id, userID)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// MarkAllNotificationsRead marks every unread notification as read.
func (s *Store) MarkAllNotificationsRead(ctx context.Context, userID string) (int, error) {
	tag, err := s.q.Exec(ctx,
		`UPDATE notifications SET read_at = now() WHERE user_id = $1 AND read_at IS NULL`, userID)
	if err != nil {
		return 0, TranslateError(err)
	}
	return int(tag.RowsAffected()), nil
}

// ---- Activity reminders ----

const reminderColumns = `id, user_id, activity_id, remind_at, status, created_at, updated_at`

func scanReminder(row rowScanner) (*models.ActivityReminder, error) {
	r := &models.ActivityReminder{}
	err := scanRow(row, &r.ID, &r.UserID, &r.ActivityID, &r.RemindAt, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// UpsertReminder creates or revives a reminder for an activity.
func (s *Store) UpsertReminder(ctx context.Context, r *models.ActivityReminder) (*models.ActivityReminder, error) {
	const q = `
		INSERT INTO activity_reminders (user_id, activity_id, remind_at, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (activity_id, user_id) DO UPDATE
		SET remind_at = EXCLUDED.remind_at,
		    status = EXCLUDED.status,
		    updated_at = now()
		RETURNING ` + reminderColumns
	return scanReminder(s.q.QueryRow(ctx, q, r.UserID, r.ActivityID, r.RemindAt, r.Status))
}

// GetReminder returns a user's reminder for an activity.
func (s *Store) GetReminder(ctx context.Context, activityID, userID string) (*models.ActivityReminder, error) {
	const q = `SELECT ` + reminderColumns + ` FROM activity_reminders WHERE activity_id = $1 AND user_id = $2`
	return scanReminder(s.q.QueryRow(ctx, q, activityID, userID))
}

// CancelReminder cancels a user's reminder for an activity.
func (s *Store) CancelReminder(ctx context.Context, activityID, userID string) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE activity_reminders SET status = $3, updated_at = now()
		 WHERE activity_id = $1 AND user_id = $2 AND status <> 'cancelled'`,
		activityID, userID, models.ReminderStatusCancelled)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListReminders returns a page of the caller's reminders with the activity title.
func (s *Store) ListReminders(ctx context.Context, userID string, p models.PageParams) ([]models.ActivityReminder, int, error) {
	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM activity_reminders WHERE user_id = $1`, userID), &total); err != nil {
		return nil, 0, err
	}
	b := &qb{}
	b.add("r.user_id = %s", userID)
	clause := b.limitOffset(p.PerPage, p.Offset)
	const sel = `
		SELECT r.id, r.user_id, r.activity_id, r.remind_at, r.status, r.created_at, r.updated_at, a.title
		FROM activity_reminders r
		JOIN activities a ON a.id = r.activity_id`
	rows, err := s.q.Query(ctx, sel+b.where()+` ORDER BY r.remind_at ASC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.ActivityReminder, 0)
	for rows.Next() {
		var r models.ActivityReminder
		if err := rows.Scan(&r.ID, &r.UserID, &r.ActivityID, &r.RemindAt, &r.Status,
			&r.CreatedAt, &r.UpdatedAt, &r.ActivityTitle); err != nil {
			return nil, 0, TranslateError(err)
		}
		out = append(out, r)
	}
	return out, total, TranslateError(rows.Err())
}

// ---- Opportunity reports ----

const reportColumns = `id, opportunity_id, user_id, reason, details, status, created_at, updated_at`

func scanReport(row rowScanner) (*models.OpportunityReport, error) {
	r := &models.OpportunityReport{}
	err := scanRow(row, &r.ID, &r.OpportunityID, &r.UserID, &r.Reason, &r.Details, &r.Status, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// CreateReport stores a user report about an opportunity.
func (s *Store) CreateReport(ctx context.Context, r *models.OpportunityReport) (*models.OpportunityReport, error) {
	const q = `
		INSERT INTO opportunity_reports (opportunity_id, user_id, reason, details)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + reportColumns
	return scanReport(s.q.QueryRow(ctx, q, r.OpportunityID, r.UserID, r.Reason, r.Details))
}

// decodeJSONColumn unmarshals a jsonb column into dst, ignoring empty values.
func decodeJSONColumn(raw []byte, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}

func nonNilData(data map[string]any) map[string]any {
	if data == nil {
		return map[string]any{}
	}
	return data
}
