package assistant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchMatchesKeywordsAndAccents(t *testing.T) {
	topic, ok := Search("¿Cómo pido ayuda a mis contactos?")
	if !ok {
		t.Fatal("expected a match for an assistance question")
	}
	if topic.Slug != "asistencia" {
		t.Fatalf("expected the assistance topic, got %q", topic.Slug)
	}

	// Accents and case must not matter.
	lower, ok := Search("como pedir ayuda")
	if !ok || lower.Slug != "asistencia" {
		t.Fatalf("accent-insensitive search failed: %q ok=%v", lower.Slug, ok)
	}
}

func TestSearchUnknownQuestion(t *testing.T) {
	if _, ok := Search("zzzqqq nonsense"); ok {
		t.Fatal("unknown question should not match a topic")
	}
	if _, ok := Search("   "); ok {
		t.Fatal("blank question should not match a topic")
	}
}

func TestKnowledgeBaseIntegrity(t *testing.T) {
	if len(KnowledgeBase) == 0 {
		t.Fatal("knowledge base is empty")
	}
	seen := map[string]bool{}
	for _, topic := range KnowledgeBase {
		if topic.Slug == "" || topic.Title == "" || topic.Answer == "" {
			t.Fatalf("topic %q is incomplete", topic.Slug)
		}
		if seen[topic.Slug] {
			t.Fatalf("duplicate topic slug %q", topic.Slug)
		}
		seen[topic.Slug] = true
	}
	if len(SortTopics()) != len(KnowledgeBase) {
		t.Fatal("SortTopics must return every topic")
	}
}

func TestNoneProviderIsHonest(t *testing.T) {
	p := New("none", "", "")
	if p.Name() != "none" {
		t.Fatalf("expected none provider, got %q", p.Name())
	}
	if _, err := p.Ask(context.Background(), "hi"); err != ErrNotConfigured {
		t.Fatalf("None.Ask must report ErrNotConfigured, got %v", err)
	}
}

func TestWebhookProviderParsesAnswer(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["question"] != "hola" {
			t.Errorf("unexpected question: %q", body["question"])
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"answer": "respuesta externa"})
	}))
	defer srv.Close()

	p := New("webhook", srv.URL, "secret-token")
	if p.Name() != "webhook" {
		t.Fatalf("expected webhook provider, got %q", p.Name())
	}
	answer, err := p.Ask(context.Background(), "hola")
	if err != nil {
		t.Fatalf("ask: %v", err)
	}
	if answer != "respuesta externa" {
		t.Fatalf("unexpected answer %q", answer)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("authorization header not forwarded: %q", gotAuth)
	}
}

func TestWebhookProviderErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := New("webhook", srv.URL, "").Ask(context.Background(), "hola"); err == nil {
		t.Fatal("non-2xx responses must be reported as errors")
	}
}

func TestNormalizeStripsPunctuation(t *testing.T) {
	got := normalize("  ¿Cómo, estás?  ")
	if strings.ContainsAny(got, ",?¿") {
		t.Fatalf("punctuation should be removed: %q", got)
	}
	if got != "como estas" {
		t.Fatalf("unexpected normalization: %q", got)
	}
}
