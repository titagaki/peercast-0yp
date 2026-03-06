// Package api provides the HTTP API for the PeerCast root server.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/megan/peercast-0yp/channel"
)

// Handler serves HTTP API endpoints backed by the channel store.
type Handler struct {
	store *channel.Store
}

// NewHandler creates a Handler backed by store.
func NewHandler(store *channel.Store) *Handler {
	return &Handler{store: store}
}

// Register attaches the API routes to mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/channels", h.handleJSONChannels)
	mux.HandleFunc("GET /channels.txt", h.handleTextChannels)
}

// channelJSON is the JSON representation of a single channel.
type channelJSON struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Genre        string    `json:"genre"`
	Desc         string    `json:"desc"`
	URL          string    `json:"url"`
	Comment      string    `json:"comment"`
	Bitrate      uint32    `json:"bitrate"`
	ContentType  string    `json:"contentType"`
	Track        trackJSON `json:"track"`
	Tracker      addrJSON  `json:"tracker"`
	NumListeners uint32    `json:"numListeners"`
	NumRelays    uint32    `json:"numRelays"`
	UpTime       uint32    `json:"upTime"`
}

type trackJSON struct {
	Title   string `json:"title"`
	Artist  string `json:"artist"`
	Album   string `json:"album"`
	Contact string `json:"contact"`
}

type addrJSON struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Firewalled bool   `json:"firewalled"`
}

func buildChannelJSON(hl channel.HitList) channelJSON {
	info := hl.Info
	out := channelJSON{
		ID:          fmt.Sprintf("%x", info.ID[:]),
		Name:        info.Name,
		Genre:       info.Genre,
		Desc:        info.Desc,
		URL:         info.URL,
		Comment:     info.Comment,
		Bitrate:     info.Bitrate,
		ContentType: info.ContentType,
		Track: trackJSON{
			Title:   info.Track.Title,
			Artist:  info.Track.Artist,
			Album:   info.Track.Album,
			Contact: info.Track.Contact,
		},
	}

	for _, hit := range hl.Hits {
		if hit.Tracker {
			ip := ""
			if hit.GlobalAddr.IP != nil {
				ip = hit.GlobalAddr.IP.String()
			}
			out.Tracker = addrJSON{
				IP:         ip,
				Port:       hit.GlobalAddr.Port,
				Firewalled: hit.Firewalled,
			}
			out.NumListeners = hit.NumListeners
			out.NumRelays = hit.NumRelays
			out.UpTime = hit.UpTime
			break
		}
	}

	return out
}

func (h *Handler) handleJSONChannels(w http.ResponseWriter, r *http.Request) {
	snap := h.store.Snapshot()

	entries := make([]channelJSON, 0, len(snap))
	for _, hl := range snap {
		entries = append(entries, buildChannelJSON(hl))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (h *Handler) handleTextChannels(w http.ResponseWriter, r *http.Request) {
	snap := h.store.Snapshot()

	entries := make([]channelJSON, 0, len(snap))
	for _, hl := range snap {
		entries = append(entries, buildChannelJSON(hl))
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, entry := range entries {
		addr := ""
		if entry.Tracker.IP != "" {
			addr = fmt.Sprintf("%s:%d", entry.Tracker.IP, entry.Tracker.Port)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\t%s\n",
			entry.Name,
			entry.ID,
			addr,
			entry.Genre,
			entry.Desc,
			entry.NumListeners,
			entry.NumRelays,
			entry.Bitrate,
			entry.ContentType,
			entry.Comment,
		)
	}
}
