package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNoneProviderReportsNotConfigured(t *testing.T) {
	p := New("none", "", "", nil)
	if p.Name() != "none" {
		t.Fatalf("name = %s", p.Name())
	}
	if _, err := p.Send(context.Background(), Message{To: "x"}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestWebhookProviderSendsJSON(t *testing.T) {
	var got Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing auth header: %q", r.Header.Get("Authorization"))
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := New("webhook", srv.URL, "secret", nil)
	rec, err := p.Send(context.Background(), Message{Channel: "sms", To: "+56900000000", Body: "hola"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if rec.Provider != "webhook" {
		t.Fatalf("provider = %s", rec.Provider)
	}
	if got.To != "+56900000000" || got.Body != "hola" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestWebhookProviderPropagatesFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := New("webhook", srv.URL, "", nil)
	if _, err := p.Send(context.Background(), Message{To: "x"}); err == nil {
		t.Fatal("expected an error for a 500 response")
	}
}
