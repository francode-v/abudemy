package api

import (
	"net/http"
)

func (s *Server) listAssistantTopics(w http.ResponseWriter, r *http.Request) {
	topics := s.svc.Assistant.Topics()
	writeJSON(w, http.StatusOK, map[string]any{"items": topics})
}

func (s *Server) askAssistant(w http.ResponseWriter, r *http.Request) {
	answer, err := s.svc.Assistant.Ask(r.Context(), queryString(r, "q"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, answer)
}
