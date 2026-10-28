package api

import "testing"

func TestOriginAllowed(t *testing.T) {
	exact := map[string]struct{}{"https://abudemy.pages.dev": {}}
	wild := []string{"https://*.abudemy.pages.dev"}

	cases := []struct {
		origin string
		want   bool
	}{
		{"https://abudemy.pages.dev", true},
		{"https://abc123.abudemy.pages.dev", true},
		{"http://abc123.abudemy.pages.dev", false},
		{"https://abudemy.pages.dev.evil.com", false},
		{"https://evilabudemy.pages.dev", false},
		{"https://.abudemy.pages.dev", false},
		{"https://other.pages.dev", false},
	}
	for _, c := range cases {
		if got := originAllowed(c.origin, exact, wild); got != c.want {
			t.Errorf("originAllowed(%q) = %v, want %v", c.origin, got, c.want)
		}
	}
}
