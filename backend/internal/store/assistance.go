package store

import (
	"context"
	"time"

	"github.com/abudemy/backend/internal/models"
)

// ---- Trusted contacts ----

const contactColumns = `id, user_id, name, relationship, phone, email, priority, created_at, updated_at`

func scanContact(row rowScanner) (*models.TrustedContact, error) {
	c := &models.TrustedContact{}
	err := scanRow(row, &c.ID, &c.UserID, &c.Name, &c.Relationship, &c.Phone, &c.Email, &c.Priority, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// CreateContact inserts a trusted contact.
func (s *Store) CreateContact(ctx context.Context, c *models.TrustedContact) (*models.TrustedContact, error) {
	const q = `
		INSERT INTO trusted_contacts (user_id, name, relationship, phone, email, priority)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING ` + contactColumns
	return scanContact(s.q.QueryRow(ctx, q, c.UserID, c.Name, c.Relationship, c.Phone, c.Email, c.Priority))
}

// GetContact returns a trusted contact by id.
func (s *Store) GetContact(ctx context.Context, id string) (*models.TrustedContact, error) {
	return scanContact(s.q.QueryRow(ctx, `SELECT `+contactColumns+` FROM trusted_contacts WHERE id = $1`, id))
}

// ListContacts returns the trusted contacts of a user ordered by priority.
func (s *Store) ListContacts(ctx context.Context, userID string) ([]models.TrustedContact, error) {
	rows, err := s.q.Query(ctx,
		`SELECT `+contactColumns+` FROM trusted_contacts WHERE user_id = $1 ORDER BY priority ASC, created_at ASC`, userID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.TrustedContact, 0)
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, TranslateError(rows.Err())
}

// UpdateContact updates a trusted contact owned by a user.
func (s *Store) UpdateContact(ctx context.Context, c *models.TrustedContact) (*models.TrustedContact, error) {
	const q = `
		UPDATE trusted_contacts
		SET name=$3, relationship=$4, phone=$5, email=$6, priority=$7, updated_at=now()
		WHERE id=$1 AND user_id=$2
		RETURNING ` + contactColumns
	return scanContact(s.q.QueryRow(ctx, q, c.ID, c.UserID, c.Name, c.Relationship, c.Phone, c.Email, c.Priority))
}

// DeleteContact removes a trusted contact owned by a user.
func (s *Store) DeleteContact(ctx context.Context, id, userID string) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM trusted_contacts WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- Assistance requests ----

const assistanceColumns = `id, user_id, type, description, latitude, longitude, status, resolved_at, created_at, updated_at`

func scanAssistance(row rowScanner) (*models.AssistanceRequest, error) {
	a := &models.AssistanceRequest{}
	err := scanRow(row, &a.ID, &a.UserID, &a.Type, &a.Description, &a.Latitude, &a.Longitude,
		&a.Status, &a.ResolvedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	a.Notifications = []models.AssistanceNotification{}
	return a, nil
}

// CreateAssistanceRequest inserts an assistance request.
func (s *Store) CreateAssistanceRequest(ctx context.Context, a *models.AssistanceRequest) (*models.AssistanceRequest, error) {
	const q = `
		INSERT INTO assistance_requests (user_id, type, description, latitude, longitude, status)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING ` + assistanceColumns
	return scanAssistance(s.q.QueryRow(ctx, q, a.UserID, a.Type, a.Description, a.Latitude, a.Longitude, a.Status))
}

// GetAssistanceRequest returns a request with its notification attempts.
func (s *Store) GetAssistanceRequest(ctx context.Context, id string) (*models.AssistanceRequest, error) {
	a, err := scanAssistance(s.q.QueryRow(ctx, `SELECT `+assistanceColumns+` FROM assistance_requests WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	a.Notifications, err = s.ListNotifications(ctx, id)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// AssistanceFilter narrows an assistance listing.
type AssistanceFilter struct {
	UserID string
	Status string
	From   *time.Time
	To     *time.Time
	Page   models.PageParams
}

// ListAssistanceRequests returns a page of assistance requests.
func (s *Store) ListAssistanceRequests(ctx context.Context, f AssistanceFilter) ([]models.AssistanceRequest, int, error) {
	b := &qb{}
	if f.UserID != "" {
		b.add("user_id = %s", f.UserID)
	}
	if f.Status != "" {
		b.add("status = %s", f.Status)
	}
	if f.From != nil {
		b.add("created_at >= %s", *f.From)
	}
	if f.To != nil {
		b.add("created_at <= %s", *f.To)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM assistance_requests`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}
	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, `SELECT `+assistanceColumns+` FROM assistance_requests`+b.where()+` ORDER BY created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.AssistanceRequest, 0)
	for rows.Next() {
		a, err := scanAssistance(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, TranslateError(rows.Err())
}

// UpdateAssistanceStatus updates the request status and resolution timestamp.
func (s *Store) UpdateAssistanceStatus(ctx context.Context, id string, status models.AssistanceStatus) error {
	const q = `
		UPDATE assistance_requests
		SET status=$2, resolved_at = CASE WHEN $2 IN ('resolved','cancelled') THEN now() ELSE resolved_at END, updated_at=now()
		WHERE id=$1`
	tag, err := s.q.Exec(ctx, q, id, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CountAssistanceRequests counts requests, optionally filtered by status and
// the organization that owns the requesting user's resources.
func (s *Store) CountAssistanceRequests(ctx context.Context, status string) (int, error) {
	b := &qb{}
	if status != "" {
		b.add("status = %s", status)
	}
	var n int
	err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM assistance_requests`+b.where(), b.args...), &n)
	return n, TranslateError(err)
}

// ---- Assistance notifications ----

// CreateNotification records one delivery attempt.
func (s *Store) CreateNotification(ctx context.Context, n *models.AssistanceNotification) (*models.AssistanceNotification, error) {
	const q = `
		INSERT INTO assistance_notifications (assistance_request_id, contact_id, channel, recipient, provider, status, error, sent_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, assistance_request_id, contact_id, channel, recipient, provider, status, error, sent_at, created_at`
	out := &models.AssistanceNotification{}
	err := scanRow(s.q.QueryRow(ctx, q, n.AssistanceRequestID, n.ContactID, n.Channel, n.Recipient,
		n.Provider, n.Status, n.Error, n.SentAt),
		&out.ID, &out.AssistanceRequestID, &out.ContactID, &out.Channel, &out.Recipient,
		&out.Provider, &out.Status, &out.Error, &out.SentAt, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListNotifications lists the notification attempts of a request.
func (s *Store) ListNotifications(ctx context.Context, requestID string) ([]models.AssistanceNotification, error) {
	const q = `
		SELECT id, assistance_request_id, contact_id, channel, recipient, provider, status, error, sent_at, created_at
		FROM assistance_notifications WHERE assistance_request_id = $1 ORDER BY created_at ASC`
	rows, err := s.q.Query(ctx, q, requestID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.AssistanceNotification, 0)
	for rows.Next() {
		var n models.AssistanceNotification
		if err := rows.Scan(&n.ID, &n.AssistanceRequestID, &n.ContactID, &n.Channel, &n.Recipient,
			&n.Provider, &n.Status, &n.Error, &n.SentAt, &n.CreatedAt); err != nil {
			return nil, TranslateError(err)
		}
		out = append(out, n)
	}
	return out, TranslateError(rows.Err())
}
