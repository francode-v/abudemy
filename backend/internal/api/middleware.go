package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/models"
	"github.com/abudemy/backend/internal/service"
)

// statusRecorder captures the response status for structured logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		s.log.InfoContext(r.Context(), "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestIDFrom(r.Context()),
		)
	})
}

func (s *Server) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.ErrorContext(r.Context(), "panic recovered", "panic", rec, "path", r.URL.Path)
				writeError(w, r, s.log, apperr.Internal(nil))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// originAllowed reports whether origin matches one of the configured entries.
// Entries are exact origins (https://app.example.com) or may use a single
// leading wildcard label (https://*.example.pages.dev) to cover preview URLs.
func originAllowed(origin string, exact map[string]struct{}, wildcards []string) bool {
	if _, ok := exact[origin]; ok {
		return true
	}
	for _, w := range wildcards { // w looks like "https://" + "*." + "suffix"
		scheme, rest, ok := strings.Cut(w, "://")
		if !ok || !strings.HasPrefix(rest, "*.") {
			continue
		}
		suffix := rest[1:] // ".example.pages.dev"
		oScheme, oHost, ok := strings.Cut(origin, "://")
		if ok && oScheme == scheme && strings.HasSuffix(oHost, suffix) && len(oHost) > len(suffix) {
			return true
		}
	}
	return false
}

func (s *Server) cors(next http.Handler) http.Handler {
	allowAll := false
	allowed := make(map[string]struct{})
	var wildcards []string
	for _, o := range s.corsOrigins {
		switch {
		case o == "*":
			allowAll = true
		case strings.Contains(o, "://*."):
			wildcards = append(wildcards, o)
		default:
			allowed[strings.TrimRight(o, "/")] = struct{}{}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if originAllowed(origin, allowed, wildcards) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
		w.Header().Set("Access-Control-Max-Age", "600")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authenticate validates the bearer token and loads the principal, verifying
// that the backing session is still active (supports session revocation).
func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			writeError(w, r, s.log, apperr.Unauthorized("missing authorization header"))
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, r, s.log, apperr.Unauthorized("authorization header must be 'Bearer <token>'"))
			return
		}
		claims, err := s.tokens.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			writeError(w, r, s.log, apperr.Unauthorized("invalid or expired token"))
			return
		}
		active, err := s.store.SessionIsActive(r.Context(), claims.SessionID)
		if err != nil {
			writeError(w, r, s.log, apperr.Internal(err))
			return
		}
		if !active {
			writeError(w, r, s.log, apperr.Unauthorized("session has been revoked or expired"))
			return
		}
		// Load the account on every request so role changes, suspensions and
		// institutional approvals take effect immediately instead of waiting
		// for the access token to expire.
		user, err := s.store.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			writeError(w, r, s.log, apperr.Unauthorized("the account no longer exists"))
			return
		}
		if user.Status != models.UserStatusActive {
			writeError(w, r, s.log, apperr.Unauthorized("this account is not active"))
			return
		}
		p := service.Principal{
			UserID:                user.ID,
			Role:                  user.Role,
			SessionID:             claims.SessionID,
			InstitutionalApproved: user.InstitutionalApproved,
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
	})
}

// requireRoles enforces that the caller holds one of the allowed roles.
func (s *Server) requireRoles(roles ...models.Role) func(http.Handler) http.Handler {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := principalFrom(r.Context())
			if !ok {
				writeError(w, r, s.log, apperr.Unauthorized("authentication required"))
				return
			}
			if _, ok := allowed[p.Role]; !ok {
				writeError(w, r, s.log, apperr.Forbidden("your role does not permit this action"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// roleInstitutional is the set of roles allowed to use institutional features.
var roleInstitutional = []models.Role{models.RoleAdmin, models.RoleInstitutionalClient, models.RoleOrgCollaborator}

// roleAdminOnly is the set of roles allowed to use administrative features.
var roleAdminOnly = []models.Role{models.RoleAdmin}

// roleRegisteredPeopleReaders may list community accounts. The service also
// requires institutional clients to be approved and restricts them to
// mobile_user records.
var roleRegisteredPeopleReaders = []models.Role{models.RoleAdmin, models.RoleInstitutionalClient}

// auth is the per-route authentication middleware.
func (s *Server) auth(next http.Handler) http.Handler { return s.authenticate(next) }
