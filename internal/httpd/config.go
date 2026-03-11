package httpd

import (
	"encoding/json"
	"net/http"
)

type configJSON struct {
	YPIndexURL string `json:"ypIndexURL"`
	PCPAddress string `json:"pcpAddress"`
}

func (s *Server) handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configJSON{
		YPIndexURL: s.ypIndexURL,
		PCPAddress: s.pcpAddress,
	})
}
