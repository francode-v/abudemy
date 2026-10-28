package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/auth"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/notify"
	"github.com/abudemy/backend/internal/store"
	"github.com/abudemy/backend/internal/validate"
)

// AuthService implements registration, authentication and session management.
type AuthService struct{ d Deps }

// RegisterInput is the payload for account registration.
type RegisterInput struct {
	Email           string
	Password        string
	FullName        string
	Phone           string
	BirthDate       *time.Time
	Comuna          string
	HealthSystem    models.HealthSystem
	Sex             models.Sex
	LivingSituation models.LivingSituation
	Role            models.Role
}

// LoginInput is the payload for signing in.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IP        string
}

// TokenPair is the access/refresh token response.
type TokenPair struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	SessionID    string       `json:"session_id"`
	User         *models.User `json:"user"`
}

// Register creates a new account. Public registration only allows mobile users
// and institutional clients. Institutional clients start unapproved: an
// administrator must enable their access to the institutional portal.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*models.User, error) {
	var v validate.Errors
	v.Required("email", in.Email)
	v.Email("email", in.Email)
	v.MaxLen("email", in.Email, 254)
	v.Required("password", in.Password)
	v.Required("full_name", in.FullName)
	v.MinLen("full_name", in.FullName, 2)
	v.MaxLen("full_name", in.FullName, 120)
	v.MaxLen("phone", in.Phone, 40)
	v.MaxLen("comuna", in.Comuna, 80)
	validateBirthDate(&v, "birth_date", in.BirthDate)
	validateHealthSystem(&v, "health_system", in.HealthSystem)
	validateSex(&v, "sex", in.Sex)
	validateLivingSituation(&v, "living_situation", in.LivingSituation)
	if err := v.Err(); err != nil {
		return nil, err
	}
	if err := auth.ValidatePassword(in.Password); err != nil {
		return nil, apperr.InvalidField("password", err.Error())
	}

	role := in.Role
	if role == "" {
		role = models.RoleMobileUser
	}
	if role != models.RoleMobileUser && role != models.RoleInstitutionalClient {
		return nil, apperr.InvalidField("role", "must be mobile_user or institutional_client")
	}

	email := auth.NormalizeEmail(in.Email)
	hash, err := s.d.Hasher.Hash(in.Password)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	// Institutional clients must be approved by an administrator before they
	// can use the institutional portal. Mobile users are unaffected.
	approved := role != models.RoleInstitutionalClient
	user, err := s.d.Store.CreateUser(ctx, store.CreateUserParams{
		Email:                 email,
		PasswordHash:          hash,
		FullName:              strings.TrimSpace(in.FullName),
		Phone:                 strings.TrimSpace(in.Phone),
		BirthDate:             in.BirthDate,
		Comuna:                strings.TrimSpace(in.Comuna),
		HealthSystem:          in.HealthSystem,
		Sex:                   in.Sex,
		LivingSituation:       in.LivingSituation,
		Role:                  role,
		InstitutionalApproved: approved,
	})
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, apperr.Conflict("an account with that email already exists")
		}
		return nil, mapErr(err, "user")
	}
	return user, nil
}

// Login authenticates a user and issues a token pair.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*TokenPair, error) {
	var v validate.Errors
	v.Required("email", in.Email)
	v.Required("password", in.Password)
	if err := v.Err(); err != nil {
		return nil, err
	}

	user, hash, err := s.d.Store.GetUserWithHash(ctx, auth.NormalizeEmail(in.Email))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, apperr.Unauthorized("invalid email or password")
		}
		return nil, mapErr(err, "user")
	}
	if !s.d.Hasher.Verify(hash, in.Password) {
		return nil, apperr.Unauthorized("invalid email or password")
	}
	if user.Status != models.UserStatusActive {
		return nil, apperr.Forbidden("this account is not active")
	}

	pair, err := s.issueTokens(ctx, user, in.UserAgent, in.IP)
	if err != nil {
		return nil, err
	}
	if err := s.d.Store.TouchLastLogin(ctx, user.ID); err != nil {
		s.d.Logger.WarnContext(ctx, "failed to record last login", "error", err)
	}
	return pair, nil
}

// Refresh rotates a refresh token and returns a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (*TokenPair, error) {
	if refreshToken == "" {
		return nil, apperr.Unauthorized("missing refresh token")
	}
	sess, err := s.d.Store.GetSessionByTokenHash(ctx, auth.HashToken(refreshToken))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, apperr.Unauthorized("invalid refresh token")
		}
		return nil, mapErr(err, "session")
	}
	if sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
		return nil, apperr.Unauthorized("session expired or revoked")
	}

	user, err := s.d.Store.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, mapErr(err, "user")
	}
	if user.Status != models.UserStatusActive {
		return nil, apperr.Forbidden("this account is not active")
	}

	if err := s.d.Store.RevokeSession(ctx, sess.ID); err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, mapErr(err, "session")
	}
	return s.issueTokens(ctx, user, userAgent, ip)
}

func (s *AuthService) issueTokens(ctx context.Context, user *models.User, userAgent, ip string) (*TokenPair, error) {
	refreshToken, err := auth.RandomToken(32)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	sess, err := s.d.Store.CreateSession(ctx, user.ID, auth.HashToken(refreshToken), userAgent, ip,
		time.Now().Add(s.d.RefreshTTL))
	if err != nil {
		return nil, mapErr(err, "session")
	}
	access, expiresAt, err := s.d.Tokens.GenerateAccess(user, sess.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(time.Until(expiresAt).Seconds()),
		SessionID:    sess.ID,
		User:         user,
	}, nil
}

// Logout revokes the caller's current session.
func (s *AuthService) Logout(ctx context.Context, p Principal) error {
	if p.SessionID == "" {
		return nil
	}
	if err := s.d.Store.RevokeSession(ctx, p.SessionID); err != nil && !errors.Is(err, store.ErrNotFound) {
		return mapErr(err, "session")
	}
	return nil
}

// LogoutAll revokes every session of the caller.
func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	return mapErr(s.d.Store.RevokeUserSessions(ctx, userID), "session")
}

// ChangePassword verifies the current password, stores the new one and revokes
// every active session.
func (s *AuthService) ChangePassword(ctx context.Context, p Principal, currentPassword, newPassword string) error {
	if currentPassword == "" || newPassword == "" {
		var v validate.Errors
		v.Required("current_password", currentPassword)
		v.Required("new_password", newPassword)
		return v.Err()
	}
	_, hash, err := s.d.Store.GetUserWithHashByID(ctx, p.UserID)
	if err != nil {
		return mapErr(err, "user")
	}
	if !s.d.Hasher.Verify(hash, currentPassword) {
		return apperr.Unauthorized("the current password is incorrect")
	}
	if err := auth.ValidatePassword(newPassword); err != nil {
		return apperr.InvalidField("new_password", err.Error())
	}
	newHash, err := s.d.Hasher.Hash(newPassword)
	if err != nil {
		return apperr.Internal(err)
	}
	if err := s.d.Store.UpdateUserPassword(ctx, p.UserID, newHash); err != nil {
		return mapErr(err, "user")
	}
	if err := s.d.Store.RevokeUserSessions(ctx, p.UserID); err != nil {
		s.d.Logger.WarnContext(ctx, "failed to revoke sessions after password change", "error", err)
	}
	return nil
}

// RequestPasswordReset creates a reset token and dispatches it through the
// configured provider. It never reveals whether the email exists.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	if email == "" {
		return apperr.InvalidField("email", "is required")
	}
	user, _, err := s.d.Store.GetUserWithHash(ctx, auth.NormalizeEmail(email))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil // Do not disclose account existence.
		}
		return mapErr(err, "user")
	}

	token, err := auth.RandomToken(32)
	if err != nil {
		return apperr.Internal(err)
	}
	if err := s.d.Store.CreatePasswordReset(ctx, user.ID, auth.HashToken(token), time.Now().Add(s.d.PasswordResetTTL)); err != nil {
		return mapErr(err, "password reset")
	}

	msg := notify.Message{
		Channel: "email",
		To:      user.Email,
		Subject: "Recuperación de contraseña de Abudemy",
		Body:    "Usa el siguiente código para restablecer tu contraseña: " + token,
	}
	if _, err := s.d.Notifier.Send(ctx, msg); err != nil {
		// The token is stored, but delivery depends on an external provider.
		s.d.Logger.WarnContext(ctx, "password reset token could not be delivered",
			"provider", s.d.Notifier.Name(), "error", err)
	}
	return nil
}

// ResetPassword consumes a reset token and sets a new password.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if token == "" || newPassword == "" {
		var v validate.Errors
		v.Required("token", token)
		v.Required("new_password", newPassword)
		return v.Err()
	}
	if err := auth.ValidatePassword(newPassword); err != nil {
		return apperr.InvalidField("new_password", err.Error())
	}
	userID, err := s.d.Store.ConsumePasswordReset(ctx, auth.HashToken(token))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return apperr.Validation(apperr.FieldError{Field: "token", Message: "is invalid or has expired"})
		}
		return mapErr(err, "password reset")
	}
	hash, err := s.d.Hasher.Hash(newPassword)
	if err != nil {
		return apperr.Internal(err)
	}
	if err := s.d.Store.UpdateUserPassword(ctx, userID, hash); err != nil {
		return mapErr(err, "user")
	}
	if err := s.d.Store.RevokeUserSessions(ctx, userID); err != nil {
		s.d.Logger.WarnContext(ctx, "failed to revoke sessions after password reset", "error", err)
	}
	return nil
}

// Sessions lists the caller's sessions, flagging the current one.
func (s *AuthService) Sessions(ctx context.Context, p Principal) ([]models.Session, error) {
	sessions, err := s.d.Store.ListUserSessions(ctx, p.UserID)
	if err != nil {
		return nil, mapErr(err, "session")
	}
	for i := range sessions {
		sessions[i].Current = sessions[i].ID == p.SessionID
	}
	return sessions, nil
}

// RevokeSession revokes one of the caller's sessions.
func (s *AuthService) RevokeSession(ctx context.Context, p Principal, sessionID string) error {
	return mapErr(s.d.Store.RevokeSessionForUser(ctx, p.UserID, sessionID), "session")
}
