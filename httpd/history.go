package httpd

import (
	"database/sql"
	"encoding/json"
	"encoding/hex"
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

	query := `
		SELECT id, channel_id, channel_name, bitrate, content_type, genre, url,
		       started_at, ended_at,
		       TIMESTAMPDIFF(MINUTE, started_at, IFNULL(ended_at, NOW())) AS duration_min
		FROM channel_sessions
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?`

	rows, err := s.db.QueryContext(r.Context(), query, limit, offset)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sessions []sessionJSON
	for rows.Next() {
		var sess sessionJSON
		var chanID []byte
		var startedAt time.Time
		var endedAt sql.NullTime

		if err := rows.Scan(
			&sess.ID, &chanID, &sess.ChannelName,
			&sess.Bitrate, &sess.ContentType, &sess.Genre, &sess.URL,
			&startedAt, &endedAt, &sess.DurationMin,
		); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}

		sess.ChannelID = hex.EncodeToString(chanID)
		sess.StartedAt = startedAt.Format(time.RFC3339)
		if endedAt.Valid {
			s := endedAt.Time.Format(time.RFC3339)
			sess.EndedAt = &s
		}
		sessions = append(sessions, sess)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}
