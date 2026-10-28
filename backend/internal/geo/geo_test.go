package geo

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSelectsProvider(t *testing.T) {
	if _, ok := New("none", "", "").(None); !ok {
		t.Fatal("expected None provider")
	}
	if _, ok := New("nominatim", "", "").(Nominatim); !ok {
		t.Fatal("expected Nominatim provider")
	}
	// An unknown provider falls back to the honest default.
	if _, ok := New("something-else", "", "").(None); !ok {
		t.Fatal("unknown providers must fall back to None")
	}
}

func TestNoneProviderIsHonest(t *testing.T) {
	var p None
	if p.Name() != "none" {
		t.Fatalf("unexpected name %q", p.Name())
	}
	if _, err := p.Search(context.Background(), "x", 5); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("None.Search must report ErrNotConfigured, got %v", err)
	}
	if _, err := p.Reverse(context.Background(), 0, 0); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("None.Reverse must report ErrNotConfigured, got %v", err)
	}
}

func TestNominatimSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "Concepción" {
			t.Errorf("query not forwarded: %q", r.URL.Query().Get("q"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent must be set for Nominatim")
		}
		_, _ = w.Write([]byte(`[{"display_name":"Concepción, Biobío, Chile","lat":"-36.827","lon":"-73.0503","type":"city","address":{"country":"Chile","state":"Biobío","city":"Concepción"}}]`))
	}))
	defer srv.Close()

	results, err := Nominatim{BaseURL: srv.URL}.Search(context.Background(), "Concepción", 3)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Latitude != -36.827 || r.Longitude != -73.0503 {
		t.Fatalf("coordinates not parsed: %+v", r)
	}
	if r.Region != "Biobío" || r.Comuna != "Concepción" || r.Country != "Chile" {
		t.Fatalf("address not parsed: %+v", r)
	}
}

func TestNominatimReverseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	if _, err := (Nominatim{BaseURL: srv.URL}).Reverse(context.Background(), -33, -70); err == nil {
		t.Fatal("non-2xx responses must be reported as errors")
	}
}
