package httpd

import (
	"database/sql"
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

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT
			recorded_at, listeners, relays,
			name, genre, description, url, comment, track_title, track_artist,
			LAG(name)        OVER w AS prev_name,
			LAG(genre)       OVER w AS prev_genre,
			LAG(description) OVER w AS prev_description,
			LAG(url)         OVER w AS prev_url,
			LAG(comment)     OVER w AS prev_comment,
			LAG(track_title) OVER w AS prev_track_title,
			LAG(track_artist) OVER w AS prev_track_artist
		FROM channel_snapshots
		WHERE channel_id = ? AND recorded_at >= ? AND recorded_at < ?
		WINDOW w AS (PARTITION BY session_id ORDER BY recorded_at)
		ORDER BY recorded_at`,
		chanID, dayStart, dayEnd,
	)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []timelineRowJSON
	for rows.Next() {
		var recordedAt time.Time
		var listeners, relays int
		var name, genre, desc, url, comment, trackTitle, trackArtist string
		var prevName, prevGenre, prevDesc, prevURL, prevComment, prevTrackTitle, prevTrackArtist sql.NullString

		if err := rows.Scan(
			&recordedAt, &listeners, &relays,
			&name, &genre, &desc, &url, &comment, &trackTitle, &trackArtist,
			&prevName, &prevGenre, &prevDesc, &prevURL, &prevComment, &prevTrackTitle, &prevTrackArtist,
		); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}

		changed := !prevName.Valid ||
			name != prevName.String ||
			genre != prevGenre.String ||
			desc != prevDesc.String ||
			url != prevURL.String ||
			comment != prevComment.String ||
			trackTitle != prevTrackTitle.String ||
			trackArtist != prevTrackArtist.String

		row := timelineRowJSON{
			RecordedAt: recordedAt.Format(time.RFC3339),
			Listeners:  listeners,
			Relays:     relays,
			Changed:    changed,
		}
		if changed {
			row.Name = name
			row.Genre = genre
			row.Description = desc
			row.URL = url
			row.Comment = comment
			row.TrackTitle = trackTitle
			row.TrackArtist = trackArtist
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
