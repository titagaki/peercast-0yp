// Package archive records channel session and snapshot data to MySQL.
package archive

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"

	pcp "github.com/titagaki/peercast-pcp/pcp"

	"github.com/titagaki/peercast-0yp/channel"
)

// Recorder polls channel.Store and writes sessions/snapshots to MySQL.
type Recorder struct {
	db    *sql.DB
	store *channel.Store
	log   *slog.Logger

	active   map[pcp.GnuID]sessionRecord // channel_id → open session
	lastSnap time.Time                   // last time snapshots were written
}

type sessionRecord struct {
	id        int64
	startedAt time.Time
}

// New returns a Recorder. Call Start to begin polling.
func New(db *sql.DB, store *channel.Store, log *slog.Logger) *Recorder {
	return &Recorder{
		db:    db,
		store: store,
		log:   log,
		active: make(map[pcp.GnuID]sessionRecord),
	}
}

// Start runs the poll loop until ctx is cancelled.
// It first closes any sessions left open from a previous crash.
func (r *Recorder) Start(ctx context.Context) {
	if err := r.closeStaleSessions(ctx); err != nil {
		r.log.Error("archive: closeStaleSessions", "err", err)
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			r.poll(ctx)
		}
	}
}

// closeStaleSessions closes sessions left open by a previous crash.
func (r *Recorder) closeStaleSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE channel_sessions SET ended_at = NOW() WHERE ended_at IS NULL`)
	return err
}

func (r *Recorder) poll(ctx context.Context) {
	now := time.Now()
	states := r.store.SnapshotOrdered()

	// Build set of currently active channel IDs.
	currentIDs := make(map[pcp.GnuID]struct{}, len(states))
	for _, s := range states {
		currentIDs[s.Info.ID] = struct{}{}
	}

	// New channels → INSERT channel_sessions.
	for _, s := range states {
		id := s.Info.ID
		if _, exists := r.active[id]; !exists {
			sessionID, err := r.insertSession(ctx, s, now)
			if err != nil {
				r.log.Error("archive: insertSession", "channel", s.Info.Name, "err", err)
				continue
			}
			r.active[id] = sessionRecord{id: sessionID, startedAt: now}
		}
	}

	// Gone channels → UPDATE ended_at.
	for id, sess := range r.active {
		if _, exists := currentIDs[id]; !exists {
			if err := r.closeSession(ctx, sess.id, now); err != nil {
				r.log.Error("archive: closeSession", "session_id", sess.id, "err", err)
			}
			delete(r.active, id)
		}
	}

	// Every 1 minute: INSERT snapshots for all active channels.
	if now.Sub(r.lastSnap) >= time.Minute {
		snapTime := now.Truncate(time.Minute)
		for _, s := range states {
			sess, ok := r.active[s.Info.ID]
			if !ok {
				continue
			}
			if err := r.insertSnapshot(ctx, sess.id, s, snapTime); err != nil {
				r.log.Error("archive: insertSnapshot", "channel", s.Info.Name, "err", err)
			}
		}
		r.lastSnap = now
	}
}

func (r *Recorder) insertSession(ctx context.Context, s channel.ChannelState, now time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO channel_sessions
			(channel_id, channel_name, bitrate, content_type, genre, url, started_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.Info.ID[:],
		s.Info.Name,
		s.Info.Bitrate,
		s.Info.ContentType,
		stripYPPrefix(s.Info.Genre),
		s.Info.URL,
		now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Recorder) closeSession(ctx context.Context, id int64, now time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE channel_sessions SET ended_at = ? WHERE id = ?`, now, id)
	return err
}

func (r *Recorder) insertSnapshot(ctx context.Context, sessionID int64, s channel.ChannelState, t time.Time) error {
	hidden := strings.Contains(s.Info.Genre, "?")
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO channel_snapshots
			(session_id, channel_id, recorded_at,
			 listeners, relays,
			 name, bitrate, content_type, genre, description, url, comment,
			 hidden_listeners,
			 track_title, track_artist, track_album, track_contact)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID,
		s.Info.ID[:],
		t,
		s.Listeners,
		s.Relays,
		s.Info.Name,
		s.Info.Bitrate,
		s.Info.ContentType,
		stripYPPrefix(s.Info.Genre),
		s.Info.Desc,
		s.Info.URL,
		s.Info.Comment,
		hidden,
		s.Info.Track.Title,
		s.Info.Track.Artist,
		s.Info.Track.Album,
		s.Info.Track.Contact,
	)
	return err
}

// stripYPPrefix removes a "xx:" YP prefix from genre strings (e.g. "ap:Music" → "Music").
func stripYPPrefix(genre string) string {
	if idx := strings.Index(genre, ":"); idx != -1 {
		return genre[idx+1:]
	}
	return genre
}
