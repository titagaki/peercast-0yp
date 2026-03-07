package httpd

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type sessionJSON struct {
	ID          int64   `json:"id"`
	ChannelID   string  `json:"channelId"`
	ChannelName string  `json:"channelName"`
	Bitrate     int     `json:"bitrate"`
	ContentType string  `json:"contentType"`
	Genre       string  `json:"genre"`
	URL         string  `json:"url"`
	StartedAt   string  `json:"startedAt"`
	EndedAt     *string `json:"endedAt"`
	DurationMin int     `json:"durationMin"`
}

func (s *Server) handleAPIHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 50
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	rows, err := s.sessions.List(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	sessions := make([]sessionJSON, 0, len(rows))
	for _, row := range rows {
		sj := sessionJSON{
			ID:          row.ID,
			ChannelID:   hex.EncodeToString(row.ChannelID),
			ChannelName: row.ChannelName,
			Bitrate:     row.Bitrate,
			ContentType: row.ContentType,
			Genre:       row.Genre,
			URL:         row.URL,
			StartedAt:   row.StartedAt.Format(time.RFC3339),
			DurationMin: row.DurationMin,
		}
		if row.EndedAt != nil {
			t := row.EndedAt.Format(time.RFC3339)
			sj.EndedAt = &t
		}
		sessions = append(sessions, sj)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}
