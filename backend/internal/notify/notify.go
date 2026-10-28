// Package notify abstracts outbound notifications. The platform never claims a
// message was sent unless a real provider accepted it: when no provider is
// configured the returned error is ErrNotConfigured and callers must record the
// attempt as "skipped".
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// ErrNotConfigured is returned when no provider is available.
var ErrNotConfigured = errors.New("notification provider not configured")

// Message is a single outbound notification.
type Message struct {
	Channel string         `json:"channel"`
	To      string         `json:"to"`
	Subject string         `json:"subject,omitempty"`
	Body    string         `json:"body"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// Receipt describes a provider's acknowledgement.
type Receipt struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id,omitempty"`
}

// Provider sends notifications.
type Provider interface {
	// Name identifies the provider in stored notification records.
	Name() string
	// Send delivers a message. It returns ErrNotConfigured when inactive.
	Send(ctx context.Context, m Message) (Receipt, error)
}

// None is the provider used when notifications are disabled. It never pretends
// a message was delivered.
type None struct{}

// Name implements Provider.
func (None) Name() string { return "none" }

// Send implements Provider and always reports ErrNotConfigured.
func (None) Send(context.Context, Message) (Receipt, error) { return Receipt{}, ErrNotConfigured }

// Log writes notifications to the structured log. It is intended for local
// development only and must not be used in production.
type Log struct {
	Logger *slog.Logger
}

// Name implements Provider.
func (Log) Name() string { return "log" }

// Send implements Provider.
func (l Log) Send(ctx context.Context, m Message) (Receipt, error) {
	logger := l.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.InfoContext(ctx, "notification dispatched (log provider)",
		"channel", m.Channel, "to", m.To, "subject", m.Subject)
	return Receipt{Provider: "log"}, nil
}

// Webhook posts notifications as JSON to a configured HTTP endpoint.
type Webhook struct {
	URL    string
	Token  string
	Client *http.Client
}

// Name implements Provider.
func (Webhook) Name() string { return "webhook" }

// Send implements Provider.
func (w Webhook) Send(ctx context.Context, m Message) (Receipt, error) {
	body, err := json.Marshal(m)
	if err != nil {
		return Receipt{}, fmt.Errorf("encode notification: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return Receipt{}, fmt.Errorf("build notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if w.Token != "" {
		req.Header.Set("Authorization", "Bearer "+w.Token)
	}
	client := w.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Receipt{}, fmt.Errorf("send notification: %w", err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Receipt{}, fmt.Errorf("notification provider returned %d: %s", resp.StatusCode, string(payload))
	}
	return Receipt{Provider: "webhook"}, nil
}

// New builds a provider from configuration values.
func New(provider, webhookURL, webhookToken string, logger *slog.Logger) Provider {
	switch provider {
	case "log":
		return Log{Logger: logger}
	case "webhook":
		return Webhook{URL: webhookURL, Token: webhookToken}
	default:
		return None{}
	}
}
