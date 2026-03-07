package httpd

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type timelineRowJSON struct {
	RecordedAt  string `json:"recordedAt"`
	Listeners   int    `json:"listeners"`
	Relays      int    `json:"relays"`
	Changed     bool   `json:"changed"`
	Name        string `json:"name,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Comment     string `json:"comment,omitempty"`
	TrackTitle  string `json:"trackTitle,omitempty"`
	TrackArtist string `json:"trackArtist,omitempty"`
}

func (s *Server) handleAPITimeline(w http.ResponseWriter, r *http.Request) {
	chanID, err := parseChannelID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if len(dateStr) != 8 {
		http.Error(w, "date must be YYYYMMDD", http.StatusBadRequest)
		return
	}
	date, err := time.ParseInLocation("20060102", dateStr, time.Local)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	dayStart := date
	dayEnd := date.AddDate(0, 0, 1)

	rows, err := s.snapshots.ListByChannelAndDate(r.Context(), chanID, dayStart, dayEnd)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	result := make([]timelineRowJSON, 0, len(rows))
	for _, row := range rows {
		tj := timelineRowJSON{
			RecordedAt: row.RecordedAt.Format(time.RFC3339),
			Listeners:  row.Listeners,
			Relays:     row.Relays,
			Changed:    row.Changed,
		}
		if row.Changed {
			tj.Name = row.Name
			tj.Genre = row.Genre
			tj.Description = row.Description
			tj.URL = row.URL
			tj.Comment = row.Comment
			tj.TrackTitle = row.TrackTitle
			tj.TrackArtist = row.TrackArtist
		}
		result = append(result, tj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
