package api

import (
	"net/http"

	"github.com/abudemy/backend/internal/apperr"
)

func (s *Server) geoSearch(w http.ResponseWriter, r *http.Request) {
	results, err := s.svc.Geo.Search(r.Context(), queryString(r, "q"), queryInt(r, "limit"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": results})
}

func (s *Server) geoReverse(w http.ResponseWriter, r *http.Request) {
	lat := queryFloatPtr(r, "lat")
	lng := queryFloatPtr(r, "lng")
	if lat == nil {
		writeError(w, r, s.log, apperr.InvalidField("lat", "is required and must be a number"))
		return
	}
	if lng == nil {
		writeError(w, r, s.log, apperr.InvalidField("lng", "is required and must be a number"))
		return
	}
	result, err := s.svc.Geo.Reverse(r.Context(), *lat, *lng)
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
