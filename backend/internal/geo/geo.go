// Package geo abstracts forward and reverse geocoding. No provider is enabled
// by default: without one the API reports the capability as not configured
// instead of returning invented coordinates.
package geo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrNotConfigured is returned when no geocoding provider is available.
var ErrNotConfigured = errors.New("geocoding provider not configured")

// Result is a single geocoding match.
type Result struct {
	Label     string  `json:"label"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Type      string  `json:"type,omitempty"`
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	Comuna    string  `json:"comuna,omitempty"`
}

// Provider performs geocoding lookups.
type Provider interface {
	// Name identifies the provider.
	Name() string
	// Search resolves a free-text place name into candidate coordinates.
	Search(ctx context.Context, query string, limit int) ([]Result, error)
	// Reverse resolves coordinates into a place label.
	Reverse(ctx context.Context, lat, lng float64) (*Result, error)
}

// None is the provider used when geocoding is disabled.
type None struct{}

// Name implements Provider.
func (None) Name() string { return "none" }

// Search implements Provider and always reports ErrNotConfigured.
func (None) Search(context.Context, string, int) ([]Result, error) { return nil, ErrNotConfigured }

// Reverse implements Provider and always reports ErrNotConfigured.
func (None) Reverse(context.Context, float64, float64) (*Result, error) {
	return nil, ErrNotConfigured
}

// Nominatim queries an OpenStreetMap Nominatim compatible endpoint. Operators
// must respect the provider's usage policy (a descriptive User-Agent, low
// request volume and local caching).
type Nominatim struct {
	BaseURL   string
	UserAgent string
	Client    *http.Client
}

// Name implements Provider.
func (Nominatim) Name() string { return "nominatim" }

// Search implements Provider.
func (n Nominatim) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit < 1 || limit > 20 {
		limit = 5
	}
	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "jsonv2")
	q.Set("addressdetails", "1")
	q.Set("limit", strconv.Itoa(limit))
	var raw []nominatimPlace
	if err := n.get(ctx, "/search", q, &raw); err != nil {
		return nil, err
	}
	out := make([]Result, 0, len(raw))
	for _, p := range raw {
		out = append(out, p.toResult())
	}
	return out, nil
}

// Reverse implements Provider.
func (n Nominatim) Reverse(ctx context.Context, lat, lng float64) (*Result, error) {
	q := url.Values{}
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lng, 'f', -1, 64))
	q.Set("format", "jsonv2")
	q.Set("addressdetails", "1")
	var raw nominatimPlace
	if err := n.get(ctx, "/reverse", q, &raw); err != nil {
		return nil, err
	}
	res := raw.toResult()
	return &res, nil
}

func (n Nominatim) get(ctx context.Context, path string, q url.Values, dst any) error {
	base := strings.TrimRight(n.BaseURL, "/")
	if base == "" {
		base = "https://nominatim.openstreetmap.org"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path+"?"+q.Encode(), nil)
	if err != nil {
		return fmt.Errorf("build geocoding request: %w", err)
	}
	ua := n.UserAgent
	if ua == "" {
		ua = "Abudemy/1.0 (contact: soporte@abudemy.cl)"
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json")
	client := n.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("geocoding provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("geocoding provider returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read geocoding response: %w", err)
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode geocoding response: %w", err)
	}
	return nil
}

type nominatimPlace struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Type        string `json:"type"`
	Address     struct {
		Country      string `json:"country"`
		State        string `json:"state"`
		Region       string `json:"region"`
		City         string `json:"city"`
		Town         string `json:"town"`
		Village      string `json:"village"`
		County       string `json:"county"`
		Suburb       string `json:"suburb"`
		Municipality string `json:"municipality"`
	} `json:"address"`
}

func (p nominatimPlace) toResult() Result {
	lat, _ := strconv.ParseFloat(p.Lat, 64)
	lng, _ := strconv.ParseFloat(p.Lon, 64)
	region := p.Address.State
	if region == "" {
		region = p.Address.Region
	}
	comuna := p.Address.City
	if comuna == "" {
		comuna = p.Address.Town
	}
	if comuna == "" {
		comuna = p.Address.Village
	}
	if comuna == "" {
		comuna = p.Address.Municipality
	}
	if comuna == "" {
		comuna = p.Address.County
	}
	return Result{
		Label:     p.DisplayName,
		Latitude:  lat,
		Longitude: lng,
		Type:      p.Type,
		Country:   p.Address.Country,
		Region:    region,
		Comuna:    comuna,
	}
}

// New builds a provider from configuration values.
func New(provider, baseURL, userAgent string) Provider {
	if provider == "nominatim" {
		return Nominatim{BaseURL: baseURL, UserAgent: userAgent}
	}
	return None{}
}
