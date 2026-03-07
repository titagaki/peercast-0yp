// Package repository wraps all MySQL queries for the PeerCast YP.
package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/titagaki/peercast-0yp/internal/channel"
)

// Session is a row from channel_sessions.
type Session struct {
	ID          int64
	ChannelID   []byte
	ChannelName string
	Bitrate     int
	ContentType string
	Genre       string
	URL         string
	StartedAt   time.Time
	EndedAt     *time.Time
	DurationMin int
}

// SessionInterval is a [start, end) pair from channel_sessions.
type SessionInterval struct {
	Start time.Time
	End   time.Time
}

// SessionRepo wraps channel_sessions queries.
type SessionRepo struct {
	db *sql.DB
}

// NewSessionRepo creates a SessionRepo backed by db.
func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// CloseStaleSessions closes sessions left open by a previous crash.
func (r *SessionRepo) CloseStaleSessions(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE channel_sessions SET ended_at = NOW() WHERE ended_at IS NULL`)
	return err
}

// Insert creates a new session row and returns its ID.
func (r *SessionRepo) Insert(ctx context.Context, s channel.ChannelState, now time.Time) (int64, error) {
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

// Close sets ended_at on the session row identified by id.
func (r *SessionRepo) Close(ctx context.Context, id int64, now time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE channel_sessions SET ended_at = ? WHERE id = ?`, now, id)
	return err
}

// List returns up to limit sessions starting from offset, ordered by started_at DESC.
func (r *SessionRepo) List(ctx context.Context, limit, offset int) ([]Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, channel_id, channel_name, bitrate, content_type, genre, url,
		       started_at, ended_at,
		       TIMESTAMPDIFF(MINUTE, started_at, IFNULL(ended_at, NOW())) AS duration_min
		FROM channel_sessions
		ORDER BY started_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		var endedAt sql.NullTime
		if err := rows.Scan(
			&s.ID, &s.ChannelID, &s.ChannelName,
			&s.Bitrate, &s.ContentType, &s.Genre, &s.URL,
			&s.StartedAt, &endedAt, &s.DurationMin,
		); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			t := endedAt.Time
			s.EndedAt = &t
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// ListIntervalsByChannel returns [start, end) pairs for all sessions of the
// given channel in the past 365 days, ordered by started_at.
func (r *SessionRepo) ListIntervalsByChannel(ctx context.Context, chanID []byte) ([]SessionInterval, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT started_at, IFNULL(ended_at, NOW())
		FROM channel_sessions
		WHERE channel_id = ?
		  AND started_at >= DATE_SUB(NOW(), INTERVAL 365 DAY)
		ORDER BY started_at`, chanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var intervals []SessionInterval
	for rows.Next() {
		var iv SessionInterval
		if err := rows.Scan(&iv.Start, &iv.End); err != nil {
			return nil, err
		}
		intervals = append(intervals, iv)
	}
	return intervals, rows.Err()
}

// stripYPPrefix removes a "xx:" YP prefix from genre strings (e.g. "ap:Music" → "Music").
func stripYPPrefix(genre string) string {
	if idx := strings.Index(genre, ":"); idx != -1 {
		return genre[idx+1:]
	}
	return genre
}
