package httpd

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"
)

type activityJSON struct {
	Date    string `json:"date"`    // "YYYY-MM-DD"
	Minutes int    `json:"minutes"` // total broadcast minutes for this date
}

func (s *Server) handleAPIActivity(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name parameter required", http.StatusBadRequest)
		return
	}

	intervals, err := s.sessions.ListIntervalsByName(r.Context(), name)
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
