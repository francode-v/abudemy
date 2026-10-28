package service

import (
	"context"
	"strings"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/geo"
)

// GeoService exposes forward and reverse geocoding. When no provider is
// configured the capability is reported as not implemented instead of
// returning invented coordinates.
type GeoService struct{ d Deps }

// Search resolves a free-text place name into candidate coordinates.
func (s *GeoService) Search(ctx context.Context, query string, limit int) ([]geo.Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, apperr.InvalidField("q", "is required")
	}
	if len([]rune(query)) > 200 {
		return nil, apperr.InvalidField("q", "must be at most 200 characters")
	}
	if s.d.Geo == nil || s.d.Geo.Name() == "none" {
		return nil, apperr.New(apperr.CodeNotImplemented, "geocoding is not configured on this server")
	}
	results, err := s.d.Geo.Search(ctx, query, limit)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return results, nil
}

// Reverse resolves coordinates into a place label.
func (s *GeoService) Reverse(ctx context.Context, lat, lng float64) (*geo.Result, error) {
	if lat < -90 || lat > 90 {
		return nil, apperr.InvalidField("lat", "must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return nil, apperr.InvalidField("lng", "must be between -180 and 180")
	}
	if s.d.Geo == nil || s.d.Geo.Name() == "none" {
		return nil, apperr.New(apperr.CodeNotImplemented, "geocoding is not configured on this server")
	}
	result, err := s.d.Geo.Reverse(ctx, lat, lng)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return result, nil
}
