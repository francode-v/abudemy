package service

import (
	"context"
	"strings"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/assistant"
)

// AssistantService answers user questions. It uses the curated knowledge base
// by default and an optional external provider when one is configured. It never
// claims to be an AI model or a substitute for professionals.
type AssistantService struct{ d Deps }

const assistantFallback = "No encontré una respuesta exacta en la guía de ayuda. " +
	"Puedes revisar los temas disponibles o contactar a soporte. " +
	"Recuerda que este asistente no reemplaza la atención médica ni los servicios de emergencia."

// Topics returns the curated help topics.
func (s *AssistantService) Topics() []assistant.Topic { return assistant.SortTopics() }

// Ask returns the best answer for a question.
func (s *AssistantService) Ask(ctx context.Context, question string) (*assistant.Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, apperr.InvalidField("q", "is required")
	}
	if len([]rune(question)) > 500 {
		return nil, apperr.InvalidField("q", "must be at most 500 characters")
	}

	if s.d.Assistant != nil && s.d.Assistant.Name() != "none" {
		text, err := s.d.Assistant.Ask(ctx, question)
		if err == nil {
			return &assistant.Answer{Question: question, Answer: text, Source: "provider", AI: true}, nil
		}
		s.d.Logger.WarnContext(ctx, "assistant provider failed; using knowledge base",
			"provider", s.d.Assistant.Name(), "error", err)
	}

	topic, ok := assistant.Search(question)
	if !ok {
		return &assistant.Answer{
			Question: question, Answer: assistantFallback,
			Source: "knowledge_base", AI: false,
		}, nil
	}
	return &assistant.Answer{
		Question: question, Answer: topic.Answer, Topic: topic.Slug,
		Source: "knowledge_base", AI: false, Related: assistant.Related(topic, 3),
	}, nil
}
