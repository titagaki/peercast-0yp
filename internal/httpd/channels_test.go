package httpd

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	pcp "github.com/titagaki/peercast-pcp/pcp"

	"github.com/titagaki/peercast-0yp/internal/channel"
)

func makeID(b byte) pcp.GnuID {
	var id pcp.GnuID
	for i := range id {
		id[i] = b
	}
	return id
}

// addChannel is a test helper that adds a single-hit channel to the store.
func addChannel(store *channel.Store, name string, chanID, bcID, sessID pcp.GnuID, listeners uint32) {
	info := channel.Info{
		ID:          chanID,
		BroadcastID: bcID,
		Name:        name,
		Bitrate:     128,
		ContentType: "MP3",
		Genre:       "Music",
		Track:       channel.Track{Title: "Song", Artist: "Artist"},
	}
	hit := channel.Hit{
		SessionID:    sessID,
		ChanID:       chanID,
		GlobalAddr:   net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 7144},
		NumListeners: listeners,
		Tracker:      true,
		Recv:         true,
	}
	store.AddHit(info, hit)
}

func TestHandleAPIChannels_Empty(t *testing.T) {
	s := &Server{store: channel.NewStore()}
	req := httptest.NewRequest(http.MethodGet, "/api/channels", nil)
	w := httptest.NewRecorder()
	s.handleAPIChannels(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var result []channelJSON
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d entries", len(result))
	}
}

func TestHandleAPIChannels_OneChannel(t *testing.T) {
	store := channel.NewStore()
	chanID := makeID(0x01)
	bcID := makeID(0x02)
	sessID := makeID(0x03)
	addChannel(store, "Test Channel", chanID, bcID, sessID, 10)

	s := &Server{store: store}
	req := httptest.NewRequest(http.MethodGet, "/api/channels", nil)
	w := httptest.NewRecorder()
	s.handleAPIChannels(w, req)

	var result []channelJSON
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(result))
	}
	got := result[0]
	if got.Name != "Test Channel" {
		t.Errorf("Name = %q, want Test Channel", got.Name)
	}
	if got.NumListeners != 10 {
		t.Errorf("NumListeners = %d, want 10", got.NumListeners)
	}
	if got.Tracker.IP != "1.2.3.4" {
		t.Errorf("Tracker.IP = %q, want 1.2.3.4", got.Tracker.IP)
	}
	if got.Tracker.Port != 7144 {
		t.Errorf("Tracker.Port = %d, want 7144", got.Tracker.Port)
	}
}

func TestHandleAPIChannels_SortedByName(t *testing.T) {
	store := channel.NewStore()
	addChannel(store, "Zebra Channel", makeID(0x10), makeID(0x11), makeID(0x12), 1)
	addChannel(store, "Alpha Channel", makeID(0x20), makeID(0x21), makeID(0x22), 2)
	addChannel(store, "Middle Channel", makeID(0x30), makeID(0x31), makeID(0x32), 3)

	s := &Server{store: store}
	req := httptest.NewRequest(http.MethodGet, "/api/channels", nil)
	w := httptest.NewRecorder()
	s.handleAPIChannels(w, req)

	var result []channelJSON
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 channels, got %d", len(result))
	}
	if result[0].Name != "Alpha Channel" || result[1].Name != "Middle Channel" || result[2].Name != "Zebra Channel" {
		t.Errorf("unexpected order: %q, %q, %q", result[0].Name, result[1].Name, result[2].Name)
	}
}
