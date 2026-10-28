package store

import (
	"context"
	"strings"

	"github.com/abudemy/backend/internal/models"
)

const orgColumns = `id, name, type, legal_id, description, email, phone, address, comuna, region, latitude, longitude, active, created_at, updated_at`

func scanOrganization(row interface {
	Scan(dest ...any) error
}) (*models.Organization, error) {
	o := &models.Organization{}
	err := scanRow(row, &o.ID, &o.Name, &o.Type, &o.LegalID, &o.Description, &o.Email, &o.Phone,
		&o.Address, &o.Comuna, &o.Region, &o.Latitude, &o.Longitude, &o.Active, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// CreateOrganization inserts an organization.
func (s *Store) CreateOrganization(ctx context.Context, o *models.Organization) (*models.Organization, error) {
	const q = `
		INSERT INTO organizations (name, type, legal_id, description, email, phone, address, comuna, region, latitude, longitude, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING ` + orgColumns
	return scanOrganization(s.q.QueryRow(ctx, q, o.Name, o.Type, o.LegalID, o.Description, o.Email, o.Phone,
		o.Address, o.Comuna, o.Region, o.Latitude, o.Longitude, o.Active))
}

// GetOrganizationByID returns an organization by id.
func (s *Store) GetOrganizationByID(ctx context.Context, id string) (*models.Organization, error) {
	return scanOrganization(s.q.QueryRow(ctx, `SELECT `+orgColumns+` FROM organizations WHERE id = $1`, id))
}

// UpdateOrganization updates mutable fields.
func (s *Store) UpdateOrganization(ctx context.Context, o *models.Organization) (*models.Organization, error) {
	const q = `
		UPDATE organizations
		SET name=$2, type=$3, legal_id=$4, description=$5, email=$6, phone=$7,
		    address=$8, comuna=$9, region=$10, latitude=$11, longitude=$12, active=$13, updated_at=now()
		WHERE id=$1
		RETURNING ` + orgColumns
	return scanOrganization(s.q.QueryRow(ctx, q, o.ID, o.Name, o.Type, o.LegalID, o.Description, o.Email, o.Phone,
		o.Address, o.Comuna, o.Region, o.Latitude, o.Longitude, o.Active))
}

// OrganizationFilter narrows an organization listing.
type OrganizationFilter struct {
	Type   string
	Search string
	Active *bool
	Page   models.PageParams
}

// ListOrganizations returns a page of organizations and the total count.
func (s *Store) ListOrganizations(ctx context.Context, f OrganizationFilter) ([]models.Organization, int, error) {
	b := &qb{}
	if f.Type != "" {
		b.add("type = %s", f.Type)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		b.add("(lower(name) LIKE %s OR lower(comuna) LIKE %s)", like, like)
	}
	if f.Active != nil {
		b.add("active = %s", *f.Active)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM organizations`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}

	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, `SELECT `+orgColumns+` FROM organizations`+b.where()+` ORDER BY name ASC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()

	orgs := make([]models.Organization, 0)
	for rows.Next() {
		o, err := scanOrganization(rows)
		if err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, *o)
	}
	return orgs, total, TranslateError(rows.Err())
}

// AddMember links a user to an organization.
func (s *Store) AddMember(ctx context.Context, orgID, userID string, role models.OrgMemberRole) (*models.OrganizationMember, error) {
	const q = `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, organization_id, user_id, role, created_at`
	m := &models.OrganizationMember{}
	err := scanRow(s.q.QueryRow(ctx, q, orgID, userID, role), &m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetMember returns the membership of a user in an organization.
func (s *Store) GetMember(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error) {
	const q = `
		SELECT id, organization_id, user_id, role, created_at
		FROM organization_members WHERE organization_id = $1 AND user_id = $2`
	m := &models.OrganizationMember{}
	err := scanRow(s.q.QueryRow(ctx, q, orgID, userID), &m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ListMembers lists the members of an organization with user details.
func (s *Store) ListMembers(ctx context.Context, orgID string) ([]models.OrganizationMember, error) {
	const q = `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.created_at, u.email, u.full_name
		FROM organization_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.organization_id = $1
		ORDER BY m.created_at ASC`
	rows, err := s.q.Query(ctx, q, orgID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()

	members := make([]models.OrganizationMember, 0)
	for rows.Next() {
		var m models.OrganizationMember
		if err := rows.Scan(&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt, &m.UserEmail, &m.UserFullName); err != nil {
			return nil, TranslateError(err)
		}
		members = append(members, m)
	}
	return members, TranslateError(rows.Err())
}

// UpdateMemberRole changes the role of a member.
func (s *Store) UpdateMemberRole(ctx context.Context, orgID, userID string, role models.OrgMemberRole) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE organization_members SET role = $3, updated_at = now() WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID, role)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RemoveMember removes a user from an organization.
func (s *Store) RemoveMember(ctx context.Context, orgID, userID string) error {
	tag, err := s.q.Exec(ctx,
		`DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`, orgID, userID)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UserOrganization pairs an organization with the caller's role in it.
type UserOrganization struct {
	Organization models.Organization  `json:"organization"`
	Role         models.OrgMemberRole `json:"role"`
}

// ListUserOrganizations returns the organizations a user belongs to.
func (s *Store) ListUserOrganizations(ctx context.Context, userID string) ([]UserOrganization, error) {
	const q = `
		SELECT ` + `o.id, o.name, o.type, o.legal_id, o.description, o.email, o.phone, o.address, o.comuna, o.region, o.latitude, o.longitude, o.active, o.created_at, o.updated_at` + `, m.role
		FROM organization_members m
		JOIN organizations o ON o.id = m.organization_id
		WHERE m.user_id = $1
		ORDER BY o.name ASC`
	rows, err := s.q.Query(ctx, q, userID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()

	out := make([]UserOrganization, 0)
	for rows.Next() {
		var uo UserOrganization
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Type, &o.LegalID, &o.Description, &o.Email, &o.Phone,
			&o.Address, &o.Comuna, &o.Region, &o.Latitude, &o.Longitude, &o.Active, &o.CreatedAt, &o.UpdatedAt, &uo.Role); err != nil {
			return nil, TranslateError(err)
		}
		uo.Organization = o
		out = append(out, uo)
	}
	return out, TranslateError(rows.Err())
}

// CountOrganizations returns the total number of organizations.
func (s *Store) CountOrganizations(ctx context.Context) (int, error) {
	var n int
	err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM organizations`), &n)
	return n, TranslateError(err)
}
