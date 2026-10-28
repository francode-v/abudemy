package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/store"
	"github.com/abudemy/backend/internal/validate"
)

// ActivityService manages workshops, events, courses and enrollments.
type ActivityService struct{ d Deps }

// ActivityInput is the payload for creating or updating an activity.
type ActivityInput struct {
	OrganizationID *string
	ServiceID      *string
	CategoryID     *string
	Title          string
	Description    string
	Type           models.ActivityType
	Modality       models.Modality
	Status         models.ActivityStatus
	Location       string
	Comuna         string
	Region         string
	Latitude       *float64
	Longitude      *float64
	StartAt        time.Time
	EndAt          *time.Time
	Capacity       int
	Price          float64
	Currency       string
}

// List returns a page of activities. Organization filters require access.
func (s *ActivityService) List(ctx context.Context, p Principal, f models.ActivityFilter) (models.Page[models.Activity], error) {
	if f.Organization != "" && !p.IsAdmin() && !s.d.canReadOrg(ctx, p, f.Organization) {
		return models.Page[models.Activity]{}, apperr.Forbidden("you do not have access to this organization")
	}
	activities, total, err := s.d.Store.ListActivities(ctx, f)
	if err != nil {
		return models.Page[models.Activity]{}, mapErr(err, "activity")
	}
	return models.NewPage(activities, total, f.Page), nil
}

// Get returns an activity by id.
func (s *ActivityService) Get(ctx context.Context, _ Principal, id string) (*models.Activity, error) {
	a, err := s.d.Store.GetActivityByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "activity")
	}
	return a, nil
}

// Create creates an activity.
func (s *ActivityService) Create(ctx context.Context, p Principal, in ActivityInput) (*models.Activity, error) {
	if err := s.d.authorizeOrgWrite(ctx, p, in.OrganizationID); err != nil {
		return nil, err
	}
	if err := validateActivityInput(in); err != nil {
		return nil, err
	}
	a := &models.Activity{
		OrganizationID: in.OrganizationID,
		ServiceID:      in.ServiceID,
		CategoryID:     in.CategoryID,
		Title:          in.Title,
		Description:    in.Description,
		Type:           defaultActivityType(in.Type),
		Modality:       defaultModality(in.Modality),
		Status:         defaultActivityStatus(in.Status),
		Location:       in.Location,
		Comuna:         in.Comuna,
		Region:         in.Region,
		Latitude:       in.Latitude,
		Longitude:      in.Longitude,
		StartAt:        in.StartAt,
		EndAt:          in.EndAt,
		Capacity:       in.Capacity,
		Price:          in.Price,
		Currency:       defaultCurrency(in.Currency),
		CreatedBy:      &p.UserID,
	}
	created, err := s.d.Store.CreateActivity(ctx, a)
	if err != nil {
		return nil, mapErr(err, "activity")
	}
	return created, nil
}

// Update updates an activity.
func (s *ActivityService) Update(ctx context.Context, p Principal, id string, in ActivityInput) (*models.Activity, error) {
	existing, err := s.d.Store.GetActivityByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "activity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, existing.OrganizationID); err != nil {
		return nil, err
	}
	if err := validateActivityInput(in); err != nil {
		return nil, err
	}
	existing.CategoryID = in.CategoryID
	existing.Title = in.Title
	existing.Description = in.Description
	existing.Type = defaultActivityType(in.Type)
	existing.Modality = defaultModality(in.Modality)
	existing.Status = defaultActivityStatus(in.Status)
	existing.Location = in.Location
	existing.Comuna = in.Comuna
	existing.Region = in.Region
	existing.Latitude = in.Latitude
	existing.Longitude = in.Longitude
	existing.StartAt = in.StartAt
	existing.EndAt = in.EndAt
	existing.Capacity = in.Capacity
	existing.Price = in.Price
	existing.Currency = defaultCurrency(in.Currency)

	updated, err := s.d.Store.UpdateActivity(ctx, existing)
	if err != nil {
		return nil, mapErr(err, "activity")
	}
	return updated, nil
}

// SetStatus changes the lifecycle status of an activity.
func (s *ActivityService) SetStatus(ctx context.Context, p Principal, id string, status models.ActivityStatus) (*models.Activity, error) {
	existing, err := s.d.Store.GetActivityByID(ctx, id)
	if err != nil {
		return nil, mapErr(err, "activity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, existing.OrganizationID); err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, apperr.InvalidField("status", "must be draft, published, cancelled or finished")
	}
	if err := s.d.Store.SetActivityStatus(ctx, id, status); err != nil {
		return nil, mapErr(err, "activity")
	}
	return s.d.Store.GetActivityByID(ctx, id)
}

// Enroll registers the caller in an activity, respecting capacity limits.
func (s *ActivityService) Enroll(ctx context.Context, p Principal, activityID string) (*store.EnrollStatus, error) {
	res, err := s.d.Store.Enroll(ctx, activityID, p.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, apperr.NotFound("activity")
		}
		if errors.Is(err, store.ErrConflict) {
			return nil, apperr.Conflict("the activity is not open for enrollment or you are already enrolled")
		}
		return nil, mapErr(err, "enrollment")
	}
	return res, nil
}

// CancelEnrollment cancels the caller's enrollment.
func (s *ActivityService) CancelEnrollment(ctx context.Context, p Principal, activityID string) error {
	err := s.d.Store.UpdateEnrollmentStatus(ctx, activityID, p.UserID, models.EnrollmentStatusCancelled)
	if err != nil {
		return mapErr(err, "enrollment")
	}
	return nil
}

// MyEnrollments lists the caller's enrollments.
func (s *ActivityService) MyEnrollments(ctx context.Context, p Principal, page models.PageParams) (models.Page[models.Enrollment], error) {
	items, total, err := s.d.Store.ListEnrollments(ctx, store.EnrollmentFilter{UserID: p.UserID, Page: page})
	if err != nil {
		return models.Page[models.Enrollment]{}, mapErr(err, "enrollment")
	}
	return models.NewPage(items, total, page), nil
}

// ListEnrollments lists enrollments for an activity. Only administrators or
// members of the activity's organization with read access may view them.
func (s *ActivityService) ListEnrollments(ctx context.Context, p Principal, activityID string, page models.PageParams) (models.Page[models.Enrollment], error) {
	activity, err := s.d.Store.GetActivityByID(ctx, activityID)
	if err != nil {
		return models.Page[models.Enrollment]{}, mapErr(err, "activity")
	}
	if !p.IsAdmin() {
		if activity.OrganizationID == nil || !s.d.canReadOrg(ctx, p, *activity.OrganizationID) {
			return models.Page[models.Enrollment]{}, apperr.Forbidden("you do not have access to this activity's enrollments")
		}
	}
	items, total, err := s.d.Store.ListEnrollments(ctx, store.EnrollmentFilter{ActivityID: activityID, Page: page})
	if err != nil {
		return models.Page[models.Enrollment]{}, mapErr(err, "enrollment")
	}
	return models.NewPage(items, total, page), nil
}

// UpdateEnrollmentStatus marks an enrollment as attended, enrolled, waitlist or
// cancelled on behalf of the activity manager.
func (s *ActivityService) UpdateEnrollmentStatus(ctx context.Context, p Principal, activityID, userID string, status models.EnrollmentStatus) error {
	activity, err := s.d.Store.GetActivityByID(ctx, activityID)
	if err != nil {
		return mapErr(err, "activity")
	}
	if err := s.d.authorizeOrgWrite(ctx, p, activity.OrganizationID); err != nil {
		return err
	}
	if !status.Valid() {
		return apperr.InvalidField("status", "must be enrolled, waitlist, cancelled or attended")
	}
	return mapErr(s.d.Store.UpdateEnrollmentStatus(ctx, activityID, userID, status), "enrollment")
}

func defaultActivityType(t models.ActivityType) models.ActivityType {
	if t == "" {
		return models.ActivityTypeWorkshop
	}
	return t
}

func defaultActivityStatus(s models.ActivityStatus) models.ActivityStatus {
	if s == "" {
		return models.ActivityStatusDraft
	}
	return s
}

func defaultModality(m models.Modality) models.Modality {
	if m == "" {
		return models.ModalityInPerson
	}
	return m
}

func defaultCurrency(c string) string {
	if strings.TrimSpace(c) == "" {
		return "CLP"
	}
	return strings.ToUpper(c)
}

func validateActivityInput(in ActivityInput) error {
	var v validate.Errors
	v.Required("title", in.Title)
	v.MaxLen("title", in.Title, 200)
	v.MaxLen("description", in.Description, 4000)
	if in.StartAt.IsZero() {
		v.Add("start_at", "is required")
	}
	if in.EndAt != nil && !in.StartAt.IsZero() && in.EndAt.Before(in.StartAt) {
		v.Add("end_at", "must be after start_at")
	}
	if in.Capacity < 0 {
		v.Add("capacity", "must not be negative")
	}
	if in.Price < 0 {
		v.Add("price", "must not be negative")
	}
	v.OneOf("type", string(in.Type),
		string(models.ActivityTypeWorkshop), string(models.ActivityTypeEvent),
		string(models.ActivityTypeCourse), string(models.ActivityTypeSupport))
	v.OneOf("modality", string(in.Modality),
		string(models.ModalityInPerson), string(models.ModalityOnline), string(models.ModalityHybrid))
	v.OneOf("status", string(in.Status),
		string(models.ActivityStatusDraft), string(models.ActivityStatusPublished),
		string(models.ActivityStatusCancelled), string(models.ActivityStatusFinished))
	v.Latitude("latitude", in.Latitude)
	v.Longitude("longitude", in.Longitude)
	return v.Err()
}

// SetReminder creates or revives a reminder for an activity. When no explicit
// time is given the reminder defaults to 24 hours before the activity starts.
// Delivery depends on a scheduler and a notification provider.
func (s *ActivityService) SetReminder(ctx context.Context, p Principal, activityID string, remindAt *time.Time) (*models.ActivityReminder, error) {
	activity, err := s.d.Store.GetActivityByID(ctx, activityID)
	if err != nil {
		return nil, mapErr(err, "activity")
	}
	when := activity.StartAt.Add(-24 * time.Hour)
	if remindAt != nil {
		when = *remindAt
	}
	if when.After(activity.StartAt) {
		return nil, apperr.InvalidField("remind_at", "must be before the activity starts")
	}
	r, err := s.d.Store.UpsertReminder(ctx, &models.ActivityReminder{
		UserID: p.UserID, ActivityID: activityID, RemindAt: when, Status: models.ReminderStatusPending,
	})
	if err != nil {
		return nil, mapErr(err, "reminder")
	}
	return r, nil
}

// CancelReminder cancels the caller's reminder for an activity.
func (s *ActivityService) CancelReminder(ctx context.Context, p Principal, activityID string) error {
	return mapErr(s.d.Store.CancelReminder(ctx, activityID, p.UserID), "reminder")
}

// MyReminders lists the caller's activity reminders.
func (s *ActivityService) MyReminders(ctx context.Context, p Principal, page models.PageParams) (models.Page[models.ActivityReminder], error) {
	items, total, err := s.d.Store.ListReminders(ctx, p.UserID, page)
	if err != nil {
		return models.Page[models.ActivityReminder]{}, mapErr(err, "reminder")
	}
	return models.NewPage(items, total, page), nil
}
