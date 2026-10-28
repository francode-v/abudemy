package api

import "net/http"

func (s *Server) paymentsConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.Payments.Config())
}

func (s *Server) createPaymentIntent(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Payments.CreateIntent(r.Context()); err != nil {
		writeError(w, r, s.log, err)
		return
	}
	writeNoContent(w)
}
