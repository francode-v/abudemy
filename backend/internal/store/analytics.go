package store

import (
	"context"
	"fmt"
	"time"

	"github.com/abudemy/backend/internal/models"
)

func (s *Store) scalar(ctx context.Context, query string, args ...any) (int, error) {
	var n int
	err := scanRow(s.q.QueryRow(ctx, query, args...), &n)
	return n, TranslateError(err)
}

// AnalyticsOverview computes the aggregate indicators for the institutional
// dashboard. When orgID is non-empty the numbers are restricted to that
// organization's resources.
func (s *Store) AnalyticsOverview(ctx context.Context, orgID string) (*models.AnalyticsOverview, error) {
	ov := &models.AnalyticsOverview{}
	var err error

	orgFilter := func(alias string) (string, []any) {
		if orgID == "" {
			return "", nil
		}
		col := "organization_id"
		if alias != "" {
			col = alias + ".organization_id"
		}
		return " WHERE " + col + " = $1", []any{orgID}
	}

	// Services
	if w, a := orgFilter(""); true {
		if ov.ServicesTotal, err = s.scalar(ctx, `SELECT count(*) FROM services`+w, a...); err != nil {
			return nil, err
		}
		if ov.ServicesActive, err = s.scalar(ctx, `SELECT count(*) FROM services`+w+andOrWhere(w)+` status='active'`, a...); err != nil {
			return nil, err
		}
	}

	// Activities
	if w, a := orgFilter(""); true {
		if ov.ActivitiesTotal, err = s.scalar(ctx, `SELECT count(*) FROM activities`+w, a...); err != nil {
			return nil, err
		}
		if ov.ActivitiesPublished, err = s.scalar(ctx, `SELECT count(*) FROM activities`+w+andOrWhere(w)+` status='published'`, a...); err != nil {
			return nil, err
		}
		if ov.ActivitiesUpcoming, err = s.scalar(ctx, `SELECT count(*) FROM activities`+w+andOrWhere(w)+` status='published' AND start_at >= now()`, a...); err != nil {
			return nil, err
		}
	}

	// Enrollments (scoped through activities)
	{
		w := ""
		var a []any
		if orgID != "" {
			w = " WHERE a.organization_id = $1"
			a = []any{orgID}
		}
		if ov.EnrollmentsTotal, err = s.scalar(ctx, `SELECT count(*) FROM activity_enrollments e JOIN activities a ON a.id=e.activity_id`+w, a...); err != nil {
			return nil, err
		}
		if ov.EnrollmentsActive, err = s.scalar(ctx, `SELECT count(*) FROM activity_enrollments e JOIN activities a ON a.id=e.activity_id`+w+andOrWhere(w)+` e.status IN ('enrolled','attended')`, a...); err != nil {
			return nil, err
		}
	}

	// Users and organizations
	if ov.UsersTotal, err = s.scalar(ctx, `SELECT count(*) FROM users`); err != nil {
		return nil, err
	}
	if ov.UsersMobile, err = s.scalar(ctx, `SELECT count(*) FROM users WHERE role = 'mobile_user'`); err != nil {
		return nil, err
	}
	if ov.OrganizationsTotal, err = s.scalar(ctx, `SELECT count(*) FROM organizations`); err != nil {
		return nil, err
	}
	if ov.OpportunitiesOpen, err = s.scalar(ctx, `SELECT count(*) FROM opportunities WHERE status='open'`); err != nil {
		return nil, err
	}

	// Assistance: scoped by organization membership of the requesting user.
	if orgID != "" {
		if ov.AssistanceTotal, err = s.scalar(ctx, `
			SELECT count(*) FROM assistance_requests r
			WHERE r.user_id IN (SELECT user_id FROM organization_members WHERE organization_id = $1)`, orgID); err != nil {
			return nil, err
		}
		if ov.AssistancePending, err = s.scalar(ctx, `
			SELECT count(*) FROM assistance_requests r
			WHERE r.status='pending'
			  AND r.user_id IN (SELECT user_id FROM organization_members WHERE organization_id = $1)`, orgID); err != nil {
			return nil, err
		}
	} else {
		if ov.AssistanceTotal, err = s.scalar(ctx, `SELECT count(*) FROM assistance_requests`); err != nil {
			return nil, err
		}
		if ov.AssistancePending, err = s.scalar(ctx, `SELECT count(*) FROM assistance_requests WHERE status='pending'`); err != nil {
			return nil, err
		}
	}

	// Average occupancy across activities with a defined capacity.
	occQuery := `
		SELECT COALESCE(avg(occ), 0) FROM (
			SELECT LEAST(1.0, (COALESCE(e.cnt,0))::float / a.capacity) AS occ
			FROM activities a
			LEFT JOIN (
				SELECT activity_id, count(*) AS cnt FROM activity_enrollments
				WHERE status IN ('enrolled','attended') GROUP BY activity_id
			) e ON e.activity_id = a.id
			WHERE a.capacity > 0` + andOrWhere2(orgID) + `
		) t`
	if err := scanRow(s.q.QueryRow(ctx, occQuery, argsFor(orgID)...), &ov.AverageOccupancy); err != nil {
		return nil, err
	}

	if ov.ServicesByCategory, err = s.ServicesByCategory(ctx, orgID); err != nil {
		return nil, err
	}
	if ov.ActivitiesByType, err = s.ActivitiesByType(ctx, orgID); err != nil {
		return nil, err
	}
	if ov.Territorial, err = s.TerritorialDistribution(ctx, orgID, 20); err != nil {
		return nil, err
	}
	return ov, nil
}

func argsFor(orgID string) []any {
	if orgID == "" {
		return nil
	}
	return []any{orgID}
}

func andOrWhere(existing string) string {
	if existing == "" {
		return " WHERE"
	}
	return " AND"
}

func andOrWhere2(orgID string) string {
	if orgID == "" {
		return ""
	}
	return " AND a.organization_id = $1"
}

// ServicesByCategory aggregates services per category.
func (s *Store) ServicesByCategory(ctx context.Context, orgID string) ([]models.CountByLabel, error) {
	w := ""
	var a []any
	if orgID != "" {
		w = " WHERE s.organization_id = $1"
		a = []any{orgID}
	}
	q := `
		SELECT COALESCE(c.name, 'Sin categoría') AS label, count(*) AS count
		FROM services s
		LEFT JOIN service_categories c ON c.id = s.category_id` + w + `
		GROUP BY c.name ORDER BY count DESC LIMIT 50`
	return s.countByLabel(ctx, q, a...)
}

// ActivitiesByType aggregates activities per type.
func (s *Store) ActivitiesByType(ctx context.Context, orgID string) ([]models.CountByLabel, error) {
	w := ""
	var a []any
	if orgID != "" {
		w = " WHERE organization_id = $1"
		a = []any{orgID}
	}
	q := `SELECT type AS label, count(*) AS count FROM activities` + w + ` GROUP BY type ORDER BY count DESC LIMIT 50`
	return s.countByLabel(ctx, q, a...)
}

func (s *Store) countByLabel(ctx context.Context, query string, args ...any) ([]models.CountByLabel, error) {
	rows, err := s.q.Query(ctx, query, args...)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.CountByLabel, 0)
	for rows.Next() {
		var c models.CountByLabel
		if err := rows.Scan(&c.Label, &c.Count); err != nil {
			return nil, TranslateError(err)
		}
		out = append(out, c)
	}
	return out, TranslateError(rows.Err())
}

// TerritorialDistribution aggregates services per comuna/region.
func (s *Store) TerritorialDistribution(ctx context.Context, orgID string, limit int) ([]models.TerritorialDistribution, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	w := " WHERE s.comuna <> ''"
	var a []any
	if orgID != "" {
		w += " AND s.organization_id = $1"
		a = []any{orgID}
	}
	q := `
		SELECT s.comuna, s.region, count(*) AS count
		FROM services s` + w + `
		GROUP BY s.comuna, s.region ORDER BY count DESC`
	rows, err := s.q.Query(ctx, fmt.Sprintf(`%s LIMIT %d`, q, limit), a...)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.TerritorialDistribution, 0)
	for rows.Next() {
		var d models.TerritorialDistribution
		if err := rows.Scan(&d.Comuna, &d.Region, &d.Count); err != nil {
			return nil, TranslateError(err)
		}
		out = append(out, d)
	}
	return out, TranslateError(rows.Err())
}

// ReportType identifies an exportable dataset.
type ReportType string

const (
	ReportServices      ReportType = "services"
	ReportActivities    ReportType = "activities"
	ReportEnrollments   ReportType = "enrollments"
	ReportAssistance    ReportType = "assistance"
	ReportOpportunities ReportType = "opportunities"
	ReportTerritorial   ReportType = "territorial"
)

// Report defines the columns, headers and safe query of an export.
type Report struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// reportQueries maps a report type to a fixed, parameterized query. Only
// whitelisted reports can be exported; user input never reaches the SQL text.
func reportDefinition(t ReportType) (string, []string, bool) {
	switch t {
	case ReportServices:
		return `SELECT s.id, s.name, COALESCE(o.name,'') AS organization, COALESCE(c.name,'') AS category,
		               s.type, s.status, s.comuna, s.region, s.phone, s.email
		        FROM services s
		        LEFT JOIN service_categories c ON c.id = s.category_id
		        LEFT JOIN organizations o ON o.id = s.organization_id
		        WHERE ($1 = '' OR s.organization_id = $1::uuid)
		        ORDER BY s.name`,
			[]string{"id", "name", "organization", "category", "type", "status", "comuna", "region", "phone", "email"}, true
	case ReportActivities:
		return `SELECT a.id, a.title, a.type, a.modality, a.status, a.comuna, a.region,
		               to_char(a.start_at, 'YYYY-MM-DD HH24:MI'), a.capacity, a.price,
		               (SELECT count(*) FROM activity_enrollments e WHERE e.activity_id=a.id AND e.status IN ('enrolled','attended'))
		        FROM activities a
		        WHERE ($1 = '' OR a.organization_id = $1::uuid)
		        ORDER BY a.start_at DESC`,
			[]string{"id", "title", "type", "modality", "status", "comuna", "region", "start_at", "capacity", "price", "enrolled"}, true
	case ReportEnrollments:
		return `SELECT e.id, a.title, u.email, u.full_name, e.status, to_char(e.created_at,'YYYY-MM-DD HH24:MI')
		        FROM activity_enrollments e
		        JOIN activities a ON a.id = e.activity_id
		        JOIN users u ON u.id = e.user_id
		        WHERE ($1 = '' OR a.organization_id = $1::uuid)
		        ORDER BY e.created_at DESC`,
			[]string{"id", "activity", "user_email", "user_name", "status", "created_at"}, true
	case ReportAssistance:
		return `SELECT r.id, r.type, r.status, to_char(r.created_at,'YYYY-MM-DD HH24:MI'),
		               COALESCE(to_char(r.resolved_at,'YYYY-MM-DD HH24:MI'),'')
		        FROM assistance_requests r
		        WHERE ($1 = '' OR r.user_id IN (SELECT user_id FROM organization_members WHERE organization_id = $1::uuid))
		        ORDER BY r.created_at DESC`,
			[]string{"id", "type", "status", "created_at", "resolved_at"}, true
	case ReportOpportunities:
		return `SELECT o.id, o.title, o.type, o.status, o.comuna, o.region,
		               COALESCE(o.salary_min::text,''), COALESCE(o.salary_max::text,''),
		               to_char(o.created_at,'YYYY-MM-DD HH24:MI')
		        FROM opportunities o
		        WHERE ($1 = '' OR o.organization_id = $1::uuid)
		        ORDER BY o.created_at DESC`,
			[]string{"id", "title", "type", "status", "comuna", "region", "salary_min", "salary_max", "created_at"}, true
	case ReportTerritorial:
		return `SELECT s.comuna, s.region, count(*)
		        FROM services s
		        WHERE s.comuna <> '' AND ($1 = '' OR s.organization_id = $1::uuid)
		        GROUP BY s.comuna, s.region ORDER BY count(*) DESC`,
			[]string{"comuna", "region", "services"}, true
	default:
		return "", nil, false
	}
}

// ExportReport runs a whitelisted report and returns headers and string rows
// ready for CSV encoding.
func (s *Store) ExportReport(ctx context.Context, t ReportType, orgID string) (*Report, error) {
	query, headers, ok := reportDefinition(t)
	if !ok {
		return nil, ErrNotFound
	}
	rows, err := s.q.Query(ctx, query, orgID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()

	out := &Report{Headers: headers, Rows: make([][]string, 0)}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, TranslateError(err)
		}
		record := make([]string, len(values))
		for i, v := range values {
			record[i] = formatValue(v)
		}
		out.Rows = append(out.Rows, record)
	}
	if err := rows.Err(); err != nil {
		return nil, TranslateError(err)
	}
	return out, nil
}

func formatValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case time.Time:
		return t.Format(time.RFC3339)
	case []byte:
		return string(t)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}
