package store

import (
	"context"
	"strings"

	"github.com/abudemy/backend/internal/models"
)

// ---- Skills ----

const skillColumns = `id, user_id, name, description, level, created_at, updated_at`

func scanSkill(row rowScanner) (*models.Skill, error) {
	sk := &models.Skill{}
	err := scanRow(row, &sk.ID, &sk.UserID, &sk.Name, &sk.Description, &sk.Level, &sk.CreatedAt, &sk.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return sk, nil
}

// CreateSkill inserts a user skill.
func (s *Store) CreateSkill(ctx context.Context, sk *models.Skill) (*models.Skill, error) {
	const q = `
		INSERT INTO skills (user_id, name, description, level)
		VALUES ($1,$2,$3,$4)
		RETURNING ` + skillColumns
	return scanSkill(s.q.QueryRow(ctx, q, sk.UserID, sk.Name, sk.Description, sk.Level))
}

// GetSkill returns a skill by id.
func (s *Store) GetSkill(ctx context.Context, id string) (*models.Skill, error) {
	return scanSkill(s.q.QueryRow(ctx, `SELECT `+skillColumns+` FROM skills WHERE id = $1`, id))
}

// ListSkills lists the skills of a user.
func (s *Store) ListSkills(ctx context.Context, userID string) ([]models.Skill, error) {
	rows, err := s.q.Query(ctx, `SELECT `+skillColumns+` FROM skills WHERE user_id = $1 ORDER BY name ASC`, userID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Skill, 0)
	for rows.Next() {
		sk, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *sk)
	}
	return out, TranslateError(rows.Err())
}

// UpdateSkill updates a skill.
func (s *Store) UpdateSkill(ctx context.Context, sk *models.Skill) (*models.Skill, error) {
	const q = `
		UPDATE skills SET name=$3, description=$4, level=$5, updated_at=now()
		WHERE id=$1 AND user_id=$2
		RETURNING ` + skillColumns
	return scanSkill(s.q.QueryRow(ctx, q, sk.ID, sk.UserID, sk.Name, sk.Description, sk.Level))
}

// DeleteSkill removes a skill owned by a user.
func (s *Store) DeleteSkill(ctx context.Context, id, userID string) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM skills WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- Offerings ----

const offeringColumns = `id, user_id, skill_id, title, description, type, status, modality, price, currency, comuna, region, created_at, updated_at`

func scanOffering(row rowScanner) (*models.Offering, error) {
	o := &models.Offering{}
	err := scanRow(row, &o.ID, &o.UserID, &o.SkillID, &o.Title, &o.Description, &o.Type, &o.Status,
		&o.Modality, &o.Price, &o.Currency, &o.Comuna, &o.Region, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// CreateOffering inserts an offering.
func (s *Store) CreateOffering(ctx context.Context, o *models.Offering) (*models.Offering, error) {
	const q = `
		INSERT INTO offerings (user_id, skill_id, title, description, type, status, modality, price, currency, comuna, region)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING ` + offeringColumns
	return scanOffering(s.q.QueryRow(ctx, q, o.UserID, o.SkillID, o.Title, o.Description, o.Type, o.Status,
		o.Modality, o.Price, o.Currency, o.Comuna, o.Region))
}

// GetOffering returns an offering by id.
func (s *Store) GetOffering(ctx context.Context, id string) (*models.Offering, error) {
	return scanOffering(s.q.QueryRow(ctx, `SELECT `+offeringColumns+` FROM offerings WHERE id = $1`, id))
}

// UpdateOffering updates an offering owned by a user.
func (s *Store) UpdateOffering(ctx context.Context, o *models.Offering) (*models.Offering, error) {
	const q = `
		UPDATE offerings
		SET skill_id=$3, title=$4, description=$5, type=$6, status=$7, modality=$8, price=$9, currency=$10, comuna=$11, region=$12, updated_at=now()
		WHERE id=$1 AND user_id=$2
		RETURNING ` + offeringColumns
	return scanOffering(s.q.QueryRow(ctx, q, o.ID, o.UserID, o.SkillID, o.Title, o.Description, o.Type,
		o.Status, o.Modality, o.Price, o.Currency, o.Comuna, o.Region))
}

// SetOfferingStatus updates an offering status.
func (s *Store) SetOfferingStatus(ctx context.Context, id, userID string, status models.OfferingStatus) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE offerings SET status=$3, updated_at=now() WHERE id=$1 AND user_id=$2`, id, userID, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// OfferingFilter narrows an offering listing.
type OfferingFilter struct {
	UserID   string
	Type     string
	Status   string
	Modality string
	Search   string
	Comuna   string
	Page     models.PageParams
}

// ListOfferings returns a page of offerings.
func (s *Store) ListOfferings(ctx context.Context, f OfferingFilter) ([]models.Offering, int, error) {
	b := &qb{}
	if f.UserID != "" {
		b.add("user_id = %s", f.UserID)
	}
	if f.Type != "" {
		b.add("type = %s", f.Type)
	}
	if f.Status != "" {
		b.add("status = %s", f.Status)
	}
	if f.Modality != "" {
		b.add("modality = %s", f.Modality)
	}
	if f.Comuna != "" {
		b.add("lower(comuna) = lower(%s)", f.Comuna)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		b.add("(lower(title) LIKE %s OR lower(description) LIKE %s)", like, like)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM offerings`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}
	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, `SELECT `+offeringColumns+` FROM offerings`+b.where()+` ORDER BY created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Offering, 0)
	for rows.Next() {
		o, err := scanOffering(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *o)
	}
	return out, total, TranslateError(rows.Err())
}

// ---- Opportunities ----

const opportunityColumns = `id, organization_id, title, description, type, status, comuna, region, salary_min, salary_max, currency, contact_email, created_by, expires_at, created_at, updated_at`

func scanOpportunity(row rowScanner) (*models.Opportunity, error) {
	o := &models.Opportunity{}
	err := scanRow(row, &o.ID, &o.OrganizationID, &o.Title, &o.Description, &o.Type, &o.Status,
		&o.Comuna, &o.Region, &o.SalaryMin, &o.SalaryMax, &o.Currency, &o.ContactEmail, &o.CreatedBy,
		&o.ExpiresAt, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// CreateOpportunity inserts an opportunity.
func (s *Store) CreateOpportunity(ctx context.Context, o *models.Opportunity) (*models.Opportunity, error) {
	const q = `
		INSERT INTO opportunities (organization_id, title, description, type, status, comuna, region,
		                           salary_min, salary_max, currency, contact_email, created_by, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING ` + opportunityColumns
	return scanOpportunity(s.q.QueryRow(ctx, q, o.OrganizationID, o.Title, o.Description, o.Type, o.Status,
		o.Comuna, o.Region, o.SalaryMin, o.SalaryMax, o.Currency, o.ContactEmail, o.CreatedBy, o.ExpiresAt))
}

// GetOpportunity returns an opportunity by id.
func (s *Store) GetOpportunity(ctx context.Context, id string) (*models.Opportunity, error) {
	return scanOpportunity(s.q.QueryRow(ctx, `SELECT `+opportunityColumns+` FROM opportunities WHERE id = $1`, id))
}

// UpdateOpportunity updates an opportunity.
func (s *Store) UpdateOpportunity(ctx context.Context, o *models.Opportunity) (*models.Opportunity, error) {
	const q = `
		UPDATE opportunities
		SET title=$2, description=$3, type=$4, status=$5, comuna=$6, region=$7,
		    salary_min=$8, salary_max=$9, currency=$10, contact_email=$11, expires_at=$12, updated_at=now()
		WHERE id=$1
		RETURNING ` + opportunityColumns
	return scanOpportunity(s.q.QueryRow(ctx, q, o.ID, o.Title, o.Description, o.Type, o.Status, o.Comuna,
		o.Region, o.SalaryMin, o.SalaryMax, o.Currency, o.ContactEmail, o.ExpiresAt))
}

// SetOpportunityStatus updates an opportunity status.
func (s *Store) SetOpportunityStatus(ctx context.Context, id string, status models.OpportunityStatus) error {
	tag, err := s.q.Exec(ctx, `UPDATE opportunities SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListOpportunities returns a page of opportunities.
func (s *Store) ListOpportunities(ctx context.Context, f models.OpportunityFilter) ([]models.Opportunity, int, error) {
	b := &qb{}
	if f.Type != "" {
		b.add("type = %s", f.Type)
	}
	if f.Status != "" {
		b.add("status = %s", f.Status)
	}
	if f.Comuna != "" {
		b.add("lower(comuna) = lower(%s)", f.Comuna)
	}
	if f.Region != "" {
		b.add("lower(region) = lower(%s)", f.Region)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		b.add("(lower(title) LIKE %s OR lower(description) LIKE %s)", like, like)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM opportunities`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}
	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, `SELECT `+opportunityColumns+` FROM opportunities`+b.where()+` ORDER BY created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Opportunity, 0)
	for rows.Next() {
		o, err := scanOpportunity(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *o)
	}
	return out, total, TranslateError(rows.Err())
}

// CreateApplication records a user application to an opportunity.
func (s *Store) CreateApplication(ctx context.Context, oppID, userID, message string) (*models.Application, error) {
	const q = `
		INSERT INTO opportunity_applications (opportunity_id, user_id, message)
		VALUES ($1,$2,$3)
		RETURNING id, opportunity_id, user_id, message, status, created_at, updated_at`
	a := &models.Application{}
	err := scanRow(s.q.QueryRow(ctx, q, oppID, userID, message),
		&a.ID, &a.OpportunityID, &a.UserID, &a.Message, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// GetApplication returns a user's application to an opportunity.
func (s *Store) GetApplication(ctx context.Context, oppID, userID string) (*models.Application, error) {
	const q = `
		SELECT id, opportunity_id, user_id, message, status, created_at, updated_at
		FROM opportunity_applications WHERE opportunity_id = $1 AND user_id = $2`
	a := &models.Application{}
	err := scanRow(s.q.QueryRow(ctx, q, oppID, userID),
		&a.ID, &a.OpportunityID, &a.UserID, &a.Message, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// GetApplicationByID returns an application by primary key.
func (s *Store) GetApplicationByID(ctx context.Context, id string) (*models.Application, error) {
	const q = `
		SELECT id, opportunity_id, user_id, message, status, created_at, updated_at
		FROM opportunity_applications WHERE id = $1`
	a := &models.Application{}
	err := scanRow(s.q.QueryRow(ctx, q, id),
		&a.ID, &a.OpportunityID, &a.UserID, &a.Message, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ListApplications lists applications for an opportunity.
func (s *Store) ListApplications(ctx context.Context, oppID string, p models.PageParams) ([]models.Application, int, error) {
	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM opportunity_applications WHERE opportunity_id = $1`, oppID), &total); err != nil {
		return nil, 0, err
	}
	b := &qb{}
	b.add("opportunity_id = %s", oppID)
	clause := b.limitOffset(p.PerPage, p.Offset)
	const sel = `
		SELECT a.id, a.opportunity_id, a.user_id, a.message, a.status, a.created_at, a.updated_at
		FROM opportunity_applications a`
	rows, err := s.q.Query(ctx, sel+b.where()+` ORDER BY a.created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Application, 0)
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.OpportunityID, &a.UserID, &a.Message, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, TranslateError(err)
		}
		out = append(out, a)
	}
	return out, total, TranslateError(rows.Err())
}

// UpdateApplicationStatus updates the status of an application, but only when
// it belongs to the given opportunity. Scoping by opportunity prevents a caller
// from mutating an application of a different opportunity or organization.
func (s *Store) UpdateApplicationStatus(ctx context.Context, oppID, id string, status models.ApplicationStatus) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE opportunity_applications SET status=$3, updated_at=now() WHERE id=$1 AND opportunity_id=$2`,
		id, oppID, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListApplicationsByUser returns a page of a user's applications with the
// opportunity summary expanded so the mobile app can render them directly.
func (s *Store) ListApplicationsByUser(ctx context.Context, userID string, p models.PageParams) ([]models.Application, int, error) {
	var total int
	if err := scanRow(s.q.QueryRow(ctx,
		`SELECT count(*) FROM opportunity_applications WHERE user_id = $1`, userID), &total); err != nil {
		return nil, 0, err
	}
	b := &qb{}
	b.add("a.user_id = %s", userID)
	clause := b.limitOffset(p.PerPage, p.Offset)
	const sel = `
		SELECT a.id, a.opportunity_id, a.user_id, a.message, a.status, a.created_at, a.updated_at,
		       o.title, o.status, o.type, o.comuna, o.region
		FROM opportunity_applications a
		JOIN opportunities o ON o.id = a.opportunity_id`
	rows, err := s.q.Query(ctx, sel+b.where()+` ORDER BY a.created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.Application, 0)
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.OpportunityID, &a.UserID, &a.Message, &a.Status, &a.CreatedAt, &a.UpdatedAt,
			&a.OpportunityTitle, &a.OpportunityStatus, &a.OpportunityType, &a.Comuna, &a.Region); err != nil {
			return nil, 0, TranslateError(err)
		}
		out = append(out, a)
	}
	return out, total, TranslateError(rows.Err())
}
