package api

import (
	"encoding/csv"
	"net/http"
	"strings"

	"github.com/abudemy/backend/internal/apperr"
	"github.com/abudemy/backend/internal/service"
)

func (s *Server) analyticsOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := s.svc.Analytics.Overview(r.Context(), mustPrincipal(r.Context()), queryString(r, "organization_id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) listReports(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"items":   service.SupportedReports(),
		"formats": []string{"csv", "json"},
	})
}

// exportReport streams a whitelisted dataset as CSV, or returns it as JSON when
// `format=json` is requested (used by the on-screen report preview). PDF export
// is documented as a future capability and returns 501 until a renderer exists.
func (s *Server) exportReport(w http.ResponseWriter, r *http.Request) {
	reportType := r.PathValue("type")
	format := strings.ToLower(queryString(r, "format"))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "json" {
		writeError(w, r, s.log, apperr.New(apperr.CodeNotImplemented, "only CSV and JSON export are implemented"))
		return
	}
	report, err := s.svc.Analytics.Export(r.Context(), mustPrincipal(r.Context()), reportType, queryString(r, "organization_id"))
	if err != nil {
		writeError(w, r, s.log, err)
		return
	}
	if format == "json" {
		// Keep the on-screen preview bounded: large datasets stay available as
		// CSV, while the UI only renders the first page worth of rows.
		const previewLimit = 200
		rows := report.Rows
		total := len(rows)
		truncated := false
		if total > previewLimit {
			rows = rows[:previewLimit]
			truncated = true
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"type":      reportType,
			"headers":   report.Headers,
			"rows":      rows,
			"total":     total,
			"truncated": truncated,
		})
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+reportType+`.csv"`)
	w.WriteHeader(http.StatusOK)

	cw := csv.NewWriter(w)
	if err := cw.Write(report.Headers); err != nil {
		s.log.ErrorContext(r.Context(), "failed to write CSV header", "error", err)
		return
	}
	for _, row := range report.Rows {
		if err := cw.Write(row); err != nil {
			s.log.ErrorContext(r.Context(), "failed to write CSV row", "error", err)
			return
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		s.log.ErrorContext(r.Context(), "failed to flush CSV", "error", err)
	}
}
