package service

import (
	"context"
	"errors"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/notify"
	"github.com/abudemy/backend/internal/store"
	"github.com/abudemy/backend/internal/validate"
)

// AssistanceService manages trusted contacts and assistance requests.
type AssistanceService struct{ d Deps }

// ContactInput is the payload for creating or updating a trusted contact.
type ContactInput struct {
	Name         string
	Relationship string
	Phone        string
	Email        string
	Priority     int
}

// AssistanceInput is the payload for creating an assistance request.
type AssistanceInput struct {
	Type        string
	Description string
	Latitude    *float64
	Longitude   *float64
	ContactID   *string
}

// ListContacts returns the caller's trusted contacts.
func (s *AssistanceService) ListContacts(ctx context.Context, p Principal) ([]models.TrustedContact, error) {
	contacts, err := s.d.Store.ListContacts(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "trusted contact")
	}
	return contacts, nil
}

// CreateContact adds a trusted contact for the caller.
func (s *AssistanceService) CreateContact(ctx context.Context, p Principal, in ContactInput) (*models.TrustedContact, error) {
	if err := validateContactInput(in); err != nil {
		return nil, err
	}
	priority := in.Priority
	if priority == 0 {
		priority = 1
	}
	c, err := s.d.Store.CreateContact(ctx, &models.TrustedContact{
		UserID: p.UserID, Name: in.Name, Relationship: in.Relationship,
		Phone: in.Phone, Email: in.Email, Priority: priority,
	})
	if err != nil {
		return nil, mapErr(err, "trusted contact")
	}
	return c, nil
}

// UpdateContact updates one of the caller's trusted contacts.
func (s *AssistanceService) UpdateContact(ctx context.Context, p Principal, id string, in ContactInput) (*models.TrustedContact, error) {
	existing, err := s.d.Store.GetContact(ctx, id)
	if err != nil {
		return nil, mapErr(err, "trusted contact")
	}
	if existing.UserID != p.UserID && !p.IsAdmin() {
		return nil, apperr.Forbidden("you may only modify your own contacts")
	}
	if err := validateContactInput(in); err != nil {
		return nil, err
	}
	existing.Name = in.Name
	existing.Relationship = in.Relationship
	existing.Phone = in.Phone
	existing.Email = in.Email
	if in.Priority != 0 {
		existing.Priority = in.Priority
	}
	updated, err := s.d.Store.UpdateContact(ctx, existing)
	if err != nil {
		return nil, mapErr(err, "trusted contact")
	}
	return updated, nil
}

// DeleteContact removes one of the caller's trusted contacts.
func (s *AssistanceService) DeleteContact(ctx context.Context, p Principal, id string) error {
	return mapErr(s.d.Store.DeleteContact(ctx, id, p.UserID), "trusted contact")
}

// CreateRequest registers an assistance request and attempts to notify the
// caller's trusted contacts. When no notification provider is configured the
// attempts are recorded as "skipped"; the platform never reports a false send.
func (s *AssistanceService) CreateRequest(ctx context.Context, p Principal, in AssistanceInput) (*models.AssistanceRequest, error) {
	var v validate.Errors
	if len(in.Type) > 60 {
		v.Add("type", "must be at most 60 characters")
	}
	v.MaxLen("description", in.Description, 2000)
	v.Latitude("latitude", in.Latitude)
	v.Longitude("longitude", in.Longitude)
	if err := v.Err(); err != nil {
		return nil, err
	}
	reqType := in.Type
	if reqType == "" {
		reqType = "general"
	}

	var contacts []models.TrustedContact
	if in.ContactID != nil && *in.ContactID != "" {
		c, err := s.d.Store.GetContact(ctx, *in.ContactID)
		if err != nil {
			return nil, mapErr(err, "trusted contact")
		}
		if c.UserID != p.UserID {
			return nil, apperr.Forbidden("that trusted contact does not belong to you")
		}
		contacts = []models.TrustedContact{*c}
	} else {
		list, err := s.d.Store.ListContacts(ctx, p.UserID)
		if err != nil {
			return nil, mapErr(err, "trusted contact")
		}
		contacts = list
	}

	req, err := s.d.Store.CreateAssistanceRequest(ctx, &models.AssistanceRequest{
		UserID: p.UserID, Type: reqType, Description: in.Description,
		Latitude: in.Latitude, Longitude: in.Longitude, Status: models.AssistanceStatusPending,
	})
	if err != nil {
		return nil, mapErr(err, "assistance request")
	}

	anySent := false
	for i := range contacts {
		rec := s.attemptNotification(ctx, req, &contacts[i])
		if rec.Status == models.NotificationSent {
			anySent = true
		}
		req.Notifications = append(req.Notifications, *rec)
	}
	if anySent {
		if err := s.d.Store.UpdateAssistanceStatus(ctx, req.ID, models.AssistanceStatusNotified); err != nil {
			s.d.Logger.WarnContext(ctx, "failed to update assistance status", "error", err)
		} else {
			req.Status = models.AssistanceStatusNotified
		}
	}
	return req, nil
}

func (s *AssistanceService) attemptNotification(ctx context.Context, req *models.AssistanceRequest, c *models.TrustedContact) *models.AssistanceNotification {
	channel := "email"
	recipient := c.Email
	if c.Phone != "" {
		channel = "sms"
		recipient = c.Phone
	}
	rec := &models.AssistanceNotification{
		AssistanceRequestID: req.ID,
		ContactID:           &c.ID,
		Channel:             channel,
		Recipient:           recipient,
		Provider:            s.d.Notifier.Name(),
		Status:              models.NotificationPending,
	}

	body := "Solicitud de asistencia de un usuario de Abudemy. Tipo: " + req.Type
	if req.Latitude != nil && req.Longitude != nil {
		body += ". Ubicación compartida disponible."
	}
	_, err := s.d.Notifier.Send(ctx, notify.Message{
		Channel: channel,
		To:      recipient,
		Subject: "Solicitud de asistencia",
		Body:    body,
		Meta:    map[string]any{"assistance_request_id": req.ID},
	})
	switch {
	case err == nil:
		now := time.Now()
		rec.Status = models.NotificationSent
		rec.SentAt = &now
	case errors.Is(err, notify.ErrNotConfigured):
		rec.Status = models.NotificationSkipped
		rec.Error = "notification provider not configured"
	default:
		rec.Status = models.NotificationFailed
		rec.Error = err.Error()
		s.d.Logger.ErrorContext(ctx, "assistance notification failed",
			"provider", s.d.Notifier.Name(), "channel", channel, "error", err)
	}

	stored, storeErr := s.d.Store.CreateNotification(ctx, rec)
	if storeErr != nil {
		s.d.Logger.ErrorContext(ctx, "failed to persist notification attempt", "error", storeErr)
		return rec
	}
	return stored
}

// GetRequest returns an assistance request. Callers may read their own
// requests; administrators may read any.
func (s *AssistanceService) GetRequest(ctx context.Context, p Principal, id string) (*models.AssistanceRequest, error) {
	req, err := s.d.Store.GetAssistanceRequest(ctx, id)
	if err != nil {
		return nil, mapErr(err, "assistance request")
	}
	if req.UserID != p.UserID && !p.IsAdmin() {
		return nil, apperr.Forbidden("you may only access your own assistance requests")
	}
	return req, nil
}

// ListRequests lists assistance requests. Non-administrators only see their own.
func (s *AssistanceService) ListRequests(ctx context.Context, p Principal, f store.AssistanceFilter) (models.Page[models.AssistanceRequest], error) {
	if !p.IsAdmin() {
		f.UserID = p.UserID
	}
	items, total, err := s.d.Store.ListAssistanceRequests(ctx, f)
	if err != nil {
		return models.Page[models.AssistanceRequest]{}, mapErr(err, "assistance request")
	}
	return models.NewPage(items, total, f.Page), nil
}

// UpdateStatus updates the status of an assistance request.
func (s *AssistanceService) UpdateStatus(ctx context.Context, p Principal, id string, status models.AssistanceStatus) (*models.AssistanceRequest, error) {
	req, err := s.d.Store.GetAssistanceRequest(ctx, id)
	if err != nil {
		return nil, mapErr(err, "assistance request")
	}
	if req.UserID != p.UserID && !p.IsAdmin() {
		return nil, apperr.Forbidden("you may only update your own assistance requests")
	}
	if !status.Valid() {
		return nil, apperr.InvalidField("status", "must be pending, notified, resolved or cancelled")
	}
	if err := s.d.Store.UpdateAssistanceStatus(ctx, id, status); err != nil {
		return nil, mapErr(err, "assistance request")
	}
	if status == models.AssistanceStatusResolved {
		s.d.notifyUser(ctx, req.UserID, "assistance_status", "Tu solicitud de ayuda fue marcada como resuelta",
			"Si necesitas ayuda nuevamente, puedes crear una nueva solicitud.",
			map[string]any{"assistance_request_id": id, "status": string(status)})
	}
	return s.d.Store.GetAssistanceRequest(ctx, id)
}

func validateContactInput(in ContactInput) error {
	var v validate.Errors
	v.Required("name", in.Name)
	v.MaxLen("name", in.Name, 120)
	v.MaxLen("relationship", in.Relationship, 80)
	v.Email("email", in.Email)
	if in.Phone == "" && in.Email == "" {
		v.Add("phone", "provide at least a phone number or an email")
	}
	if in.Priority != 0 {
		v.Range("priority", in.Priority, 1, 10)
	}
	return v.Err()
}
