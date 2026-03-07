package httpd

import "net/http"

func (s *Server) handleSPA(w http.ResponseWriter, r *http.Request) {
	// Phase 1: placeholder until React SPA is built and embedded.
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<!DOCTYPE html><html><body><h1>PeerCast YP</h1></body></html>"))
}
