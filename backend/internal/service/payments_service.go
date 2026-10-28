package service

import (
	"context"
	"strings"

	"github.com/abudemy/backend/internal/apperr"
)

// PaymentsService reports the online payments capability. No payment provider
// is enabled by default, so the platform is honest about the feature being
// unavailable and never fakes a checkout.
type PaymentsService struct{ d Deps }

// PaymentsConfig describes the payments capability.
type PaymentsConfig struct {
	Provider  string `json:"provider"`
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

const paymentsUnavailableMessage = "Los pagos en línea no están disponibles en esta plataforma. " +
	"Coordina el pago directamente con quien ofrece la clase o servicio."

// Config returns the payments capability for the current deployment.
func (s *PaymentsService) Config() PaymentsConfig {
	provider := strings.TrimSpace(s.d.PaymentsProvider)
	if provider == "" {
		provider = "none"
	}
	available := provider != "none"
	message := paymentsUnavailableMessage
	if available {
		message = ""
	}
	return PaymentsConfig{Provider: provider, Available: available, Message: message}
}

// CreateIntent reports that no payment provider is configured.
func (s *PaymentsService) CreateIntent(context.Context) error {
	if !s.Config().Available {
		return apperr.New(apperr.CodeNotImplemented, paymentsUnavailableMessage)
	}
	return apperr.New(apperr.CodeNotImplemented, "payment intents are not implemented")
}
