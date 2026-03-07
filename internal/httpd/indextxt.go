package httpd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/titagaki/peercast-0yp/internal/channel"
)

func (s *Server) handleIndexTxt(w http.ResponseWriter, r *http.Request) {
	states := s.store.SnapshotOrdered()

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	for _, cs := range states {
		writeIndexLine(w, cs)
	}
}

// genreDisplay strips the YP control prefix from a genre string and returns
// only the display portion. Format: yp[NS:][?][@@@]genre
func genreDisplay(genre string) string {
	s := strings.TrimPrefix(genre, "yp")
	// strip optional namespace (alphanum chars followed by ":")
	if i := strings.IndexByte(s, ':'); i >= 0 {
		s = s[i+1:]
	}
	// strip listener-hide flag and port-check flags
	s = strings.TrimLeft(s, "?@")
	return s
}

func writeIndexLine(w io.Writer, cs channel.ChannelState) {
	info := cs.Info
	track := info.Track

	// Field 19: 1 if any Hit has Direct=true (PCPHostFlags1Direct).
	directFlag := "0"
	for _, h := range cs.Hits {
		if h.Direct {
			directFlag = "1"
			break
		}
	}

	// Field 7/8: mask listeners/relays if genre contains "?".
	listeners := fmt.Sprintf("%d", cs.Listeners)
	relays := fmt.Sprintf("%d", cs.Relays)
	if strings.Contains(info.Genre, "?") {
		listeners = "-1"
		relays = "-1"
	}

	// Field 3: tracker IP:port from first Hit with Tracker=true.
	trackerAddr := ""
	uptime := uint32(0)
	for _, h := range cs.Hits {
		if h.Tracker && h.GlobalAddr.IP != nil {
			trackerAddr = fmt.Sprintf("%s:%d", h.GlobalAddr.IP, h.GlobalAddr.Port)
			uptime = h.UpTime
			break
		}
	}

	// Field 16: H:MM from uptime seconds.
	minutes := uptime / 60
	duration := fmt.Sprintf("%d:%02d", minutes/60, minutes%60)

	fmt.Fprintf(w, "%s<>%s<>%s<>%s<>%s<>%s<>%s<>%s<>%d<>%s<>%s<>%s<>%s<>%s<>%s<>%s<>click<>%s<>%s\n",
		info.Name,
		fmt.Sprintf("%x", info.ID[:]),
		trackerAddr,
		info.URL,
		genreDisplay(info.Genre),
		info.Desc,
		listeners,
		relays,
		info.Bitrate,
		info.ContentType,
		track.Artist,
		track.Album,
		track.Title,
		track.Contact,
		url.QueryEscape(info.Name),
		duration,
		info.Comment,
		directFlag,
	)
}
