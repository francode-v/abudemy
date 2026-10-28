package store

import (
	"context"
	"errors"
	"strings"

	"github.com/abudemy/backend/internal/models"
)

const activitySelect = `
	SELECT a.id, a.organization_id, a.service_id, a.category_id, a.title, a.description,
	       a.type, a.modality, a.status, a.location, a.comuna, a.region, a.latitude, a.longitude,
	       a.start_at, a.end_at, a.capacity, a.price, a.currency, a.created_by, a.created_at, a.updated_at,
	       (SELECT count(*) FROM activity_enrollments e
	         WHERE e.activity_id = a.id AND e.status IN ('enrolled','attended')) AS enrolled_count
	FROM activities a`

func scanActivity(row rowScanner) (*models.Activity, error) {
	a := &models.Activity{}
	err := row.Scan(&a.ID, &a.OrganizationID, &a.ServiceID, &a.CategoryID, &a.Title, &a.Description,
		&a.Type, &a.Modality, &a.Status, &a.Location, &a.Comuna, &a.Region, &a.Latitude, &a.Longitude,
		&a.StartAt, &a.EndAt, &a.Capacity, &a.Price, &a.Currency, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
		&a.EnrolledCount)
	if err != nil {
		return nil, TranslateError(err)
	}
	return a, nil
}

// CreateActivity inserts an activity.
func (s *Store) CreateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	const q = `
		INSERT INTO activities (organization_id, service_id, category_id, title, description, type, modality,
		                        status, location, comuna, region, latitude, longitude, start_at, end_at,
		                        capacity, price, currency, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		RETURNING id`
	var id string
	err := scanRow(s.q.QueryRow(ctx, q, a.OrganizationID, a.ServiceID, a.CategoryID, a.Title, a.Description,
		a.Type, a.Modality, a.Status, a.Location, a.Comuna, a.Region, a.Latitude, a.Longitude,
		a.StartAt, a.EndAt, a.Capacity, a.Price, a.Currency, a.CreatedBy), &id)
	if err != nil {
		return nil, err
	}
	return s.GetActivityByID(ctx, id)
}

// GetActivityByID returns an activity with its live enrollment count.
func (s *Store) GetActivityByID(ctx context.Context, id string) (*models.Activity, error) {
	return scanActivity(s.q.QueryRow(ctx, activitySelect+` WHERE a.id = $1`, id))
}

// UpdateActivity updates mutable fields.
func (s *Store) UpdateActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	const q = `
		UPDATE activities
		SET category_id=$2, title=$3, description=$4, type=$5, modality=$6, status=$7, location=$8,
		    comuna=$9, region=$10, latitude=$11, longitude=$12, start_at=$13, end_at=$14,
		    capacity=$15, price=$16, currency=$17, updated_at=now()
		WHERE id=$1`
	tag, err := s.q.Exec(ctx, q, a.ID, a.CategoryID, a.Title, a.Description, a.Type, a.Modality, a.Status,
		a.Location, a.Comuna, a.Region, a.Latitude, a.Longitude, a.StartAt, a.EndAt, a.Capacity, a.Price, a.Currency)
	if err != nil {
		return nil, TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return s.GetActivityByID(ctx, a.ID)
}

// SetActivityStatus changes the lifecycle status of an activity.
func (s *Store) SetActivityStatus(ctx context.Context, id string, status models.ActivityStatus) error {
	tag, err := s.q.Exec(ctx, `UPDATE activities SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListActivities returns a page of activities plus the total count.
func (s *Store) ListActivities(ctx context.Context, f models.ActivityFilter) ([]models.Activity, int, error) {
	b := &qb{}
	if f.Type != "" {
		b.add("a.type = %s", f.Type)
	}
	if f.Status != "" {
		b.add("a.status = %s", f.Status)
	}
	if f.Modality != "" {
		b.add("a.modality = %s", f.Modality)
	}
	if f.Comuna != "" {
		b.add("lower(a.comuna) = lower(%s)", f.Comuna)
	}
	if f.Region != "" {
		b.add("lower(a.region) = lower(%s)", f.Region)
	}
	if f.Organization != "" {
		b.add("a.organization_id = %s", f.Organization)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		b.add("(lower(a.title) LIKE %s OR lower(a.description) LIKE %s)", like, like)
	}
	if f.From != nil {
		b.add("a.start_at >= %s", *f.From)
	}
	if f.To != nil {
		b.add("a.start_at <= %s", *f.To)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM activities a`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}

	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, activitySelect+b.where()+` ORDER BY a.start_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()

	out := make([]models.Activity, 0)
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *a)
	}
	return out, total, TranslateError(rows.Err())
}

// EnrollStatus describes the result of an enrollment attempt.
type EnrollStatus struct {
	Enrollment *models.Enrollment
	Waitlisted bool
}

// Enroll registers a user in an activity within a transaction that locks the
// activity row so the capacity limit is never exceeded by concurrent requests.
func (s *Store) Enroll(ctx context.Context, activityID, userID string) (*EnrollStatus, error) {
	var result *EnrollStatus
	err := s.InTx(ctx, func(tx *Store) error {
		var capacity int
		var status models.ActivityStatus
		if err := scanRow(tx.q.QueryRow(ctx,
			`SELECT capacity, status FROM activities WHERE id = $1 FOR UPDATE`, activityID),
			&capacity, &status); err != nil {
			return err
		}
		if status != models.ActivityStatusPublished {
			return ErrConflict
		}

		var existing string
		found := true
		err := scanRow(tx.q.QueryRow(ctx,
			`SELECT status FROM activity_enrollments WHERE activity_id = $1 AND user_id = $2`,
			activityID, userID), &existing)
		if err != nil {
			if !isNotFound(err) {
				return err
			}
			found = false
		}
		// Only a cancelled enrollment can be reactivated. Any other status means
		// the user is already enrolled: reactivating an `attended` enrollment
		// would demote them to `enrolled`/`waitlist` and lose the attendance
		// record, so it is reported as a conflict too.
		if found && existing != string(models.EnrollmentStatusCancelled) {
			return ErrConflict
		}

		var count int
		if err := scanRow(tx.q.QueryRow(ctx,
			`SELECT count(*) FROM activity_enrollments WHERE activity_id = $1 AND status IN ('enrolled','attended')`,
			activityID), &count); err != nil {
			return err
		}

		newStatus := models.EnrollmentStatusEnrolled
		waitlisted := false
		if capacity > 0 && count >= capacity {
			newStatus = models.EnrollmentStatusWaitlist
			waitlisted = true
		}

		e := &models.Enrollment{}
		if !found {
			err = scanRow(tx.q.QueryRow(ctx, `
				INSERT INTO activity_enrollments (activity_id, user_id, status)
				VALUES ($1,$2,$3)
				RETURNING id, activity_id, user_id, status, created_at, updated_at`,
				activityID, userID, newStatus),
				&e.ID, &e.ActivityID, &e.UserID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
		} else {
			err = scanRow(tx.q.QueryRow(ctx, `
				UPDATE activity_enrollments SET status = $3, updated_at = now()
				WHERE activity_id = $1 AND user_id = $2
				RETURNING id, activity_id, user_id, status, created_at, updated_at`,
				activityID, userID, newStatus),
				&e.ID, &e.ActivityID, &e.UserID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
		}
		if err != nil {
			return err
		}
		result = &EnrollStatus{Enrollment: e, Waitlisted: waitlisted}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetEnrollment returns a user's enrollment in an activity.
func (s *Store) GetEnrollment(ctx context.Context, activityID, userID string) (*models.Enrollment, error) {
	const q = `
		SELECT id, activity_id, user_id, status, created_at, updated_at
		FROM activity_enrollments WHERE activity_id = $1 AND user_id = $2`
	e := &models.Enrollment{}
	err := scanRow(s.q.QueryRow(ctx, q, activityID, userID),
		&e.ID, &e.ActivityID, &e.UserID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// UpdateEnrollmentStatus changes the status of an enrollment.
func (s *Store) UpdateEnrollmentStatus(ctx context.Context, activityID, userID string, status models.EnrollmentStatus) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE activity_enrollments SET status = $3, updated_at = now() WHERE activity_id = $1 AND user_id = $2`,
		activityID, userID, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// EnrollmentFilter narrows an enrollment listing.
type EnrollmentFilter struct {
	ActivityID string
	UserID     string
	Status     string
	Page       models.PageParams
}

// ListEnrollments returns a page of enrollments with expanded activity/user data.
func (s *Store) ListEnrollments(ctx context.Context, f EnrollmentFilter) ([]models.Enrollment, int, error) {
	b := &qb{}
	if f.ActivityID != "" {
		b.add("e.activity_id = %s", f.ActivityID)
	}
	if f.UserID != "" {
		b.add("e.user_id = %s", f.UserID)
	}
	if f.Status != "" {
		b.add("e.status = %s", f.Status)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM activity_enrollments e`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}

	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	const sel = `
		SELECT e.id, e.activity_id, e.user_id, e.status, e.created_at, e.updated_at,
		       a.title, u.email, u.full_name
		FROM activity_enrollments e
		JOIN activities a ON a.id = e.activity_id
		JOIN users u ON u.id = e.user_id`
	rows, err := s.q.Query(ctx, sel+b.where()+` ORDER BY e.created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()

	out := make([]models.Enrollment, 0)
	for rows.Next() {
		var e models.Enrollment
		if err := rows.Scan(&e.ID, &e.ActivityID, &e.UserID, &e.Status, &e.CreatedAt, &e.UpdatedAt,
			&e.ActivityTitle, &e.UserEmail, &e.UserFullName); err != nil {
			return nil, 0, TranslateError(err)
		}
		out = append(out, e)
	}
	return out, total, TranslateError(rows.Err())
}

// CountEnrollments returns the number of non-cancelled enrollments, optionally
// scoped to an organization.
func (s *Store) CountEnrollments(ctx context.Context, orgID string) (int, error) {
	b := &qb{}
	b.conds = append(b.conds, "e.status IN ('enrolled','attended')")
	if orgID != "" {
		b.add("a.organization_id = %s", orgID)
	}
	var n int
	err := scanRow(s.q.QueryRow(ctx,
		`SELECT count(*) FROM activity_enrollments e JOIN activities a ON a.id = e.activity_id`+b.where(), b.args...), &n)
	return n, TranslateError(err)
}

func isNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
