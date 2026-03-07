package httpd

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
)

type activityJSON struct {
	Date    string `json:"date"`    // "YYYY-MM-DD"
	Minutes int    `json:"minutes"` // total broadcast minutes for this date
}

func (s *Server) handleAPIActivity(w http.ResponseWriter, r *http.Request) {
	chanID, err := parseChannelID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid channel id", http.StatusBadRequest)
		return
	}

	intervals, err := s.sessions.ListIntervalsByChannel(r.Context(), chanID)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	acc := make(map[string]int)
	for _, iv := range intervals {
		splitAcrossDays(iv.Start, iv.End, acc)
	}

	result := make([]activityJSON, 0, len(acc))
	for date, minutes := range acc {
		result = append(result, activityJSON{Date: date, Minutes: minutes})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Date < result[j].Date })

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// splitAcrossDays distributes a [start, end) interval across calendar days.
func splitAcrossDays(start, end time.Time, acc map[string]int) {
	for cur := start; cur.Before(end); {
		dayEnd := time.Date(cur.Year(), cur.Month(), cur.Day()+1, 0, 0, 0, 0, cur.Location())
		if dayEnd.After(end) {
			dayEnd = end
		}
		acc[cur.Format("2006-01-02")] += int(dayEnd.Sub(cur).Minutes())
		cur = dayEnd
	}
}

// parseChannelID parses a 32-character hex GnuID string into a []byte.
func parseChannelID(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
