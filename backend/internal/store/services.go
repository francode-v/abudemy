package store

import (
	"context"
	"strconv"
	"strings"

	"github.com/abudemy/backend/internal/models"
)

const serviceSelect = `
	SELECT s.id, s.organization_id, COALESCE(o.name, ''), s.category_id, s.name, s.description, s.type, s.status,
	       s.address, s.comuna, s.region, s.latitude, s.longitude, s.phone, s.email, s.website,
	       s.created_by, s.created_at, s.updated_at,
	       c.id, c.name, c.slug, c.description
	FROM services s
	LEFT JOIN service_categories c ON c.id = s.category_id
	LEFT JOIN organizations o ON o.id = s.organization_id`

type rowScanner interface{ Scan(dest ...any) error }

func scanService(row rowScanner) (*models.Service, error) {
	svc := &models.Service{}
	var catID, catName, catSlug, catDesc *string
	err := row.Scan(&svc.ID, &svc.OrganizationID, &svc.OrganizationName, &svc.CategoryID, &svc.Name, &svc.Description,
		&svc.Type, &svc.Status, &svc.Address, &svc.Comuna, &svc.Region, &svc.Latitude, &svc.Longitude,
		&svc.Phone, &svc.Email, &svc.Website, &svc.CreatedBy, &svc.CreatedAt, &svc.UpdatedAt,
		&catID, &catName, &catSlug, &catDesc)
	if err != nil {
		return nil, TranslateError(err)
	}
	if catID != nil {
		cat := &models.ServiceCategory{ID: *catID}
		if catName != nil {
			cat.Name = *catName
		}
		if catSlug != nil {
			cat.Slug = *catSlug
		}
		if catDesc != nil {
			cat.Description = *catDesc
		}
		svc.Category = cat
	}
	svc.Schedule = []models.ScheduleEntry{}
	return svc, nil
}

// CreateService inserts a service and returns it with category and schedule.
func (s *Store) CreateService(ctx context.Context, svc *models.Service) (*models.Service, error) {
	const q = `
		INSERT INTO services (organization_id, category_id, name, description, type, status,
		                      address, comuna, region, latitude, longitude, phone, email, website, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id`
	var id string
	err := scanRow(s.q.QueryRow(ctx, q, svc.OrganizationID, svc.CategoryID, svc.Name, svc.Description,
		svc.Type, svc.Status, svc.Address, svc.Comuna, svc.Region, svc.Latitude, svc.Longitude,
		svc.Phone, svc.Email, svc.Website, svc.CreatedBy), &id)
	if err != nil {
		return nil, err
	}
	return s.GetServiceByID(ctx, id)
}

// GetServiceByID loads a service with its category and schedule.
func (s *Store) GetServiceByID(ctx context.Context, id string) (*models.Service, error) {
	svc, err := scanService(s.q.QueryRow(ctx, serviceSelect+` WHERE s.id = $1`, id))
	if err != nil {
		return nil, err
	}
	svc.Schedule, err = s.serviceSchedule(ctx, id)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// UpdateService updates mutable fields of a service.
func (s *Store) UpdateService(ctx context.Context, svc *models.Service) (*models.Service, error) {
	const q = `
		UPDATE services
		SET category_id=$2, name=$3, description=$4, type=$5, status=$6, address=$7, comuna=$8,
		    region=$9, latitude=$10, longitude=$11, phone=$12, email=$13, website=$14, updated_at=now()
		WHERE id=$1`
	tag, err := s.q.Exec(ctx, q, svc.ID, svc.CategoryID, svc.Name, svc.Description, svc.Type, svc.Status,
		svc.Address, svc.Comuna, svc.Region, svc.Latitude, svc.Longitude, svc.Phone, svc.Email, svc.Website)
	if err != nil {
		return nil, TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return s.GetServiceByID(ctx, svc.ID)
}

// SetServiceStatus updates only the status (used to deactivate a service).
func (s *Store) SetServiceStatus(ctx context.Context, id string, status models.ServiceStatus) error {
	tag, err := s.q.Exec(ctx, `UPDATE services SET status = $2, updated_at = now() WHERE id = $1`, id, status)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteService permanently removes a service.
func (s *Store) DeleteService(ctx context.Context, id string) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM services WHERE id = $1`, id)
	if err != nil {
		return TranslateError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ReplaceSchedule replaces the weekly schedule of a service atomically.
func (s *Store) ReplaceSchedule(ctx context.Context, serviceID string, entries []models.ScheduleEntry) error {
	return s.InTx(ctx, func(tx *Store) error {
		if _, err := tx.q.Exec(ctx, `DELETE FROM service_schedules WHERE service_id = $1`, serviceID); err != nil {
			return TranslateError(err)
		}
		for _, e := range entries {
			if _, err := tx.q.Exec(ctx,
				`INSERT INTO service_schedules (service_id, day_of_week, opens_at, closes_at) VALUES ($1,$2,$3,$4)`,
				serviceID, e.DayOfWeek, e.OpensAt, e.ClosesAt); err != nil {
				return TranslateError(err)
			}
		}
		return nil
	})
}

func (s *Store) serviceSchedule(ctx context.Context, serviceID string) ([]models.ScheduleEntry, error) {
	const q = `
		SELECT day_of_week, to_char(opens_at,'HH24:MI'), to_char(closes_at,'HH24:MI')
		FROM service_schedules WHERE service_id = $1 ORDER BY day_of_week, opens_at`
	rows, err := s.q.Query(ctx, q, serviceID)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.ScheduleEntry, 0)
	for rows.Next() {
		var e models.ScheduleEntry
		if err := rows.Scan(&e.DayOfWeek, &e.OpensAt, &e.ClosesAt); err != nil {
			return nil, TranslateError(err)
		}
		out = append(out, e)
	}
	return out, TranslateError(rows.Err())
}

// serviceSchedules bulk loads schedules for many services.
func (s *Store) serviceSchedules(ctx context.Context, ids []string) (map[string][]models.ScheduleEntry, error) {
	if len(ids) == 0 {
		return map[string][]models.ScheduleEntry{}, nil
	}
	const q = `
		SELECT service_id, day_of_week, to_char(opens_at,'HH24:MI'), to_char(closes_at,'HH24:MI')
		FROM service_schedules WHERE service_id = ANY($1::uuid[]) ORDER BY day_of_week, opens_at`
	rows, err := s.q.Query(ctx, q, ids)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make(map[string][]models.ScheduleEntry, len(ids))
	for rows.Next() {
		var id string
		var e models.ScheduleEntry
		if err := rows.Scan(&id, &e.DayOfWeek, &e.OpensAt, &e.ClosesAt); err != nil {
			return nil, TranslateError(err)
		}
		out[id] = append(out[id], e)
	}
	return out, TranslateError(rows.Err())
}

// ListServices returns a page of services, applying category, location and
// proximity filters. When NearLat/NearLng/RadiusKM are set the result is
// ordered by distance ascending.
func (s *Store) ListServices(ctx context.Context, f models.ServiceFilter) ([]models.Service, int, error) {
	b := &qb{}
	if f.CategoryID != "" {
		b.add("s.category_id = %s", f.CategoryID)
	}
	if f.Type != "" {
		b.add("s.type = %s", f.Type)
	}
	if f.Status != "" {
		b.add("s.status = %s", f.Status)
	}
	if f.Comuna != "" {
		b.add("lower(s.comuna) = lower(%s)", f.Comuna)
	}
	if f.Region != "" {
		b.add("lower(s.region) = lower(%s)", f.Region)
	}
	if f.Organization != "" {
		b.add("s.organization_id = %s", f.Organization)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		b.add("(lower(s.name) LIKE %s OR lower(s.description) LIKE %s OR lower(s.address) LIKE %s)", like, like, like)
	}

	distanceExpr := ""
	if f.NearLat != nil && f.NearLng != nil && f.RadiusKM != nil {
		// Haversine distance in kilometres; the guard clamps acos input.
		distanceExpr = `(6371 * acos(least(1, greatest(-1,
			sin(radians(` + placeholder(b, *f.NearLat) + `)) * sin(radians(s.latitude)) +
			cos(radians(` + placeholder(b, *f.NearLat) + `)) * cos(radians(s.latitude)) *
			cos(radians(s.longitude - ` + placeholder(b, *f.NearLng) + `))))))`
		b.conds = append(b.conds, "s.latitude IS NOT NULL AND s.longitude IS NOT NULL AND "+distanceExpr+" <= "+placeholder(b, *f.RadiusKM))
	}

	var total int
	countSQL := `SELECT count(*) FROM services s` + b.where()
	if err := scanRow(s.q.QueryRow(ctx, countSQL, b.args...), &total); err != nil {
		return nil, 0, err
	}

	order := " ORDER BY s.created_at DESC"
	if distanceExpr != "" {
		order = " ORDER BY " + distanceExpr + " ASC"
	}
	clause := b.limitOffset(f.Page.PerPage, f.Page.Offset)
	rows, err := s.q.Query(ctx, serviceSelect+b.where()+order+clause, b.args...)
	if err != nil {
		return nil, 0, TranslateError(err)
	}
	defer rows.Close()

	services := make([]models.Service, 0)
	ids := make([]string, 0)
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, 0, err
		}
		services = append(services, *svc)
		ids = append(ids, svc.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, TranslateError(err)
	}

	schedules, err := s.serviceSchedules(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range services {
		if sc, ok := schedules[services[i].ID]; ok {
			services[i].Schedule = sc
		}
	}
	return services, total, nil
}

// placeholder registers a value and returns its numbered placeholder.
func placeholder(b *qb, v any) string {
	b.args = append(b.args, v)
	return "$" + strconv.Itoa(len(b.args))
}

// ---- Service categories ----

// ListCategories returns every service category.
func (s *Store) ListCategories(ctx context.Context) ([]models.ServiceCategory, error) {
	const q = `SELECT id, name, slug, description FROM service_categories ORDER BY name ASC`
	rows, err := s.q.Query(ctx, q)
	if err != nil {
		return nil, TranslateError(err)
	}
	defer rows.Close()
	out := make([]models.ServiceCategory, 0)
	for rows.Next() {
		var c models.ServiceCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description); err != nil {
			return nil, TranslateError(err)
		}
		out = append(out, c)
	}
	return out, TranslateError(rows.Err())
}

// GetCategoryByID returns a category by id.
func (s *Store) GetCategoryByID(ctx context.Context, id string) (*models.ServiceCategory, error) {
	var c models.ServiceCategory
	err := scanRow(s.q.QueryRow(ctx, `SELECT id, name, slug, description FROM service_categories WHERE id = $1`, id),
		&c.ID, &c.Name, &c.Slug, &c.Description)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
