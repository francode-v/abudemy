package store

import (
	"context"
	"strings"
	"time"

	"github.com/abudemy/backend/internal/models"
)

const userColumns = `id, email, full_name, phone, birth_date, comuna, health_system, sex, living_situation, role, status, institutional_approved, created_at, updated_at, last_login_at`

// CreateUserParams collects every column needed to create an account.
type CreateUserParams struct {
	Email                 string
	PasswordHash          string
	FullName              string
	Phone                 string
	BirthDate             *time.Time
	Comuna                string
	HealthSystem          models.HealthSystem
	Sex                   models.Sex
	LivingSituation       models.LivingSituation
	Role                  models.Role
	InstitutionalApproved bool
}

// CreateUser inserts a new user and returns the persisted row.
func (s *Store) CreateUser(ctx context.Context, p CreateUserParams) (*models.User, error) {
	const q = `
		INSERT INTO users (email, password_hash, full_name, phone, birth_date,
		                   comuna, health_system, sex, living_situation, role, institutional_approved)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING ` + userColumns
	u := &models.User{}
	err := scanRow(s.q.QueryRow(ctx, q, p.Email, p.PasswordHash, p.FullName, p.Phone, p.BirthDate,
		p.Comuna, p.HealthSystem, p.Sex, p.LivingSituation, p.Role, p.InstitutionalApproved),
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna, &u.HealthSystem,
		&u.Sex, &u.LivingSituation, &u.Role, &u.Status, &u.InstitutionalApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByID returns a user by primary key.
func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	u := &models.User{}
	err := scanRow(s.q.QueryRow(ctx, q, id),
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna, &u.HealthSystem,
		&u.Sex, &u.LivingSituation, &u.Role, &u.Status, &u.InstitutionalApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserWithHash returns a user together with the stored password hash.
func (s *Store) GetUserWithHash(ctx context.Context, email string) (*models.User, string, error) {
	const q = `SELECT ` + userColumns + `, password_hash FROM users WHERE email = $1`
	u := &models.User{}
	var hash string
	err := scanRow(s.q.QueryRow(ctx, q, email),
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna, &u.HealthSystem,
		&u.Sex, &u.LivingSituation, &u.Role, &u.Status, &u.InstitutionalApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt, &hash)
	if err != nil {
		return nil, "", err
	}
	return u, hash, nil
}

// GetUserWithHashByID returns a user and password hash by primary key.
func (s *Store) GetUserWithHashByID(ctx context.Context, id string) (*models.User, string, error) {
	const q = `SELECT ` + userColumns + `, password_hash FROM users WHERE id = $1`
	u := &models.User{}
	var hash string
	err := scanRow(s.q.QueryRow(ctx, q, id),
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna, &u.HealthSystem,
		&u.Sex, &u.LivingSituation, &u.Role, &u.Status, &u.InstitutionalApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt, &hash)
	if err != nil {
		return nil, "", err
	}
	return u, hash, nil
}

// UpdateUserProfileParams collects the mutable profile fields.
type UpdateUserProfileParams struct {
	FullName        string
	Phone           string
	BirthDate       *time.Time
	Comuna          string
	HealthSystem    models.HealthSystem
	Sex             models.Sex
	LivingSituation models.LivingSituation
}

// UpdateUserProfile updates mutable profile fields.
func (s *Store) UpdateUserProfile(ctx context.Context, id string, p UpdateUserProfileParams) (*models.User, error) {
	const q = `
		UPDATE users
		SET full_name = $2, phone = $3, birth_date = $4, comuna = $5,
		    health_system = $6, sex = $7, living_situation = $8, updated_at = now()
		WHERE id = $1
		RETURNING ` + userColumns
	u := &models.User{}
	err := scanRow(s.q.QueryRow(ctx, q, id, p.FullName, p.Phone, p.BirthDate, p.Comuna,
		p.HealthSystem, p.Sex, p.LivingSituation),
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna, &u.HealthSystem,
		&u.Sex, &u.LivingSituation, &u.Role, &u.Status, &u.InstitutionalApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserEmail changes the account email. Uniqueness is enforced by the
// database constraint and surfaced as ErrConflict.
func (s *Store) UpdateUserEmail(ctx context.Context, id, email string) (*models.User, error) {
	const q = `
		UPDATE users SET email = $2, updated_at = now()
		WHERE id = $1
		RETURNING ` + userColumns
	u := &models.User{}
	err := scanRow(s.q.QueryRow(ctx, q, id, email),
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna, &u.HealthSystem,
		&u.Sex, &u.LivingSituation, &u.Role, &u.Status, &u.InstitutionalApproved,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserPassword replaces the password hash.
func (s *Store) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	tag, err := s.q.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`, id, passwordHash)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserStatus changes the account status.
func (s *Store) UpdateUserStatus(ctx context.Context, id string, status models.UserStatus) error {
	tag, err := s.q.Exec(ctx, `UPDATE users SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserRole changes the platform role of an account.
func (s *Store) UpdateUserRole(ctx context.Context, id string, role models.Role) error {
	tag, err := s.q.Exec(ctx, `UPDATE users SET role = $2, updated_at = now() WHERE id = $1`, id, role)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SetUserInstitutionalApproval enables or disables institutional portal access
// for an institutional_client account.
func (s *Store) SetUserInstitutionalApproval(ctx context.Context, id string, approved bool) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE users SET institutional_approved = $2, updated_at = now() WHERE id = $1`, id, approved)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// TouchLastLogin records a successful login timestamp.
func (s *Store) TouchLastLogin(ctx context.Context, id string) error {
	_, err := s.q.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, id)
	return TranslateError(err)
}

// UserFilter narrows a user listing.
type UserFilter struct {
	Role            string
	Status          string
	Search          string
	Comuna          string
	HealthSystem    string
	Sex             string
	LivingSituation string
	// AgeMin and AgeMax filter by completed years, computed from birth_date.
	AgeMin *int
	AgeMax *int
	Page   models.PageParams
}

// ListUsers returns a page of users and the total count.
func (s *Store) ListUsers(ctx context.Context, f UserFilter) ([]models.User, int, error) {
	b := &qb{}
	if f.Role != "" {
		b.add("role = %s", f.Role)
	}
	if f.Status != "" {
		b.add("status = %s", f.Status)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		b.add("(lower(full_name) LIKE %s OR lower(email) LIKE %s)", like, like)
	}
	// Comuna uses a partial, case-insensitive match so the admin can type
	// "concep" and still find "Concepción".
	if f.Comuna != "" {
		b.add("lower(comuna) LIKE lower(%s)", "%"+f.Comuna+"%")
	}
	if f.HealthSystem != "" {
		b.add("health_system = %s", f.HealthSystem)
	}
	if f.Sex != "" {
		b.add("sex = %s", f.Sex)
	}
	if f.LivingSituation != "" {
		b.add("living_situation = %s", f.LivingSituation)
	}
	// age() returns the interval between birth_date and today, so date_part
	// yields completed years (not just the calendar-year difference).
	if f.AgeMin != nil {
		b.add("birth_date IS NOT NULL AND date_part('year', age(birth_date)) >= %s", *f.AgeMin)
	}
	if f.AgeMax != nil {
		b.add("birth_date IS NOT NULL AND date_part('year', age(birth_date)) <= %s", *f.AgeMax)
	}

	var total int
	if err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM users`+b.where(), b.args...), &total); err != nil {
		return nil, 0, err
	}

	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, `SELECT `+userColumns+` FROM users`+b.where()+` ORDER BY created_at DESC`+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Phone, &u.BirthDate, &u.Comuna,
			&u.HealthSystem, &u.Sex, &u.LivingSituation, &u.Role, &u.Status,
			&u.InstitutionalApproved, &u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt); err != nil {
			return nil, 0, TranslateError(err)
		}
		users = append(users, u)
	}
	return users, total, TranslateError(rows.Err())
}

// ---- Sessions ----

// CreateSession stores a refresh-token session.
func (s *Store) CreateSession(ctx context.Context, userID, tokenHash, userAgent, ip string, expiresAt time.Time) (*models.Session, error) {
	const q = `
		INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, user_agent, ip_address, created_at, expires_at, revoked_at`
	sess := &models.Session{}
	err := scanRow(s.q.QueryRow(ctx, q, userID, tokenHash, userAgent, ip, expiresAt),
		&sess.ID, &sess.UserID, &sess.UserAgent, &sess.IPAddress, &sess.CreatedAt, &sess.ExpiresAt, &sess.RevokedAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// GetSessionByTokenHash returns the session matching a refresh token hash.
func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	const q = `
		SELECT id, user_id, user_agent, ip_address, created_at, expires_at, revoked_at
		FROM sessions WHERE refresh_token_hash = $1`
	sess := &models.Session{}
	err := scanRow(s.q.QueryRow(ctx, q, tokenHash),
		&sess.ID, &sess.UserID, &sess.UserAgent, &sess.IPAddress, &sess.CreatedAt, &sess.ExpiresAt, &sess.RevokedAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// SessionIsActive reports whether a session is currently valid.
func (s *Store) SessionIsActive(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := scanRow(s.q.QueryRow(ctx,
		`SELECT revoked_at IS NULL AND expires_at > now() FROM sessions WHERE id = $1`, id), &ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// TouchSession updates the last usage timestamp of a session.
func (s *Store) TouchSession(ctx context.Context, id string) error {
	_, err := s.q.Exec(ctx, `UPDATE sessions SET last_used_at = now() WHERE id = $1`, id)
	return TranslateError(err)
}

// RevokeSession marks a single session as revoked.
func (s *Store) RevokeSession(ctx context.Context, id string) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeSessionForUser revokes a single session owned by a user.
func (s *Store) RevokeSessionForUser(ctx context.Context, userID, sessionID string) error {
	tag, err := s.q.Exec(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, sessionID, userID)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeUserSessions revokes every active session of a user.
func (s *Store) RevokeUserSessions(ctx context.Context, userID string) error {
	_, err := s.q.Exec(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return TranslateError(err)
}

// ListUserSessions lists the active and past sessions of a user.
func (s *Store) ListUserSessions(ctx context.Context, userID string) ([]models.Session, error) {
	const q = `
		SELECT id, user_id, user_agent, ip_address, created_at, expires_at, revoked_at
		FROM sessions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 100`
	rows, err := s.q.Query(ctx, q, userID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()

	sessions := make([]models.Session, 0)
	for rows.Next() {
		var sess models.Session
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.UserAgent, &sess.IPAddress,
			&sess.CreatedAt, &sess.ExpiresAt, &sess.RevokedAt); err != nil {
			return nil, TranslateError(err)
		}
		sessions = append(sessions, sess)
	}
	return sessions, TranslateError(rows.Err())
}

// DeleteExpiredSessions removes sessions that are expired or long revoked.
func (s *Store) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	tag, err := s.q.Exec(ctx,
		`DELETE FROM sessions WHERE expires_at < now() OR (revoked_at IS NOT NULL AND revoked_at < now() - interval '30 days')`)
	if err != nil {
		return 0, TranslateError(err)
	}
	return tag.RowsAffected(), nil
}

// ---- Password reset ----

// CreatePasswordReset stores a password reset token hash.
func (s *Store) CreatePasswordReset(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.q.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt)
	return TranslateError(err)
}

// ConsumePasswordReset validates a reset token and marks it as used. It returns
// the associated user id.
func (s *Store) ConsumePasswordReset(ctx context.Context, tokenHash string) (string, error) {
	const q = `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		RETURNING user_id`
	var userID string
	if err := scanRow(s.q.QueryRow(ctx, q, tokenHash), &userID); err != nil {
		return "", err
	}
	return userID, nil
}

// CountUsers returns the number of users, optionally filtered by role.
func (s *Store) CountUsers(ctx context.Context, role string) (int, error) {
	b := &qb{}
	if role != "" {
		b.add("role = %s", role)
	}
	var n int
	err := scanRow(s.q.QueryRow(ctx, `SELECT count(*) FROM users`+b.where(), b.args...), &n)
	return n, TranslateError(err)
}
