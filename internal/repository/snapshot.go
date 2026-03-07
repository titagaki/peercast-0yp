package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/titagaki/peercast-0yp/internal/channel"
)

// SnapshotRow is a row from channel_snapshots with change detection applied.
type SnapshotRow struct {
	RecordedAt  time.Time
	Listeners   int
	Relays      int
	Changed     bool
	Name        string
	Genre       string
	Description string
	URL         string
	Comment     string
	TrackTitle  string
	TrackArtist string
}

// SnapshotRepo wraps channel_snapshots queries.
type SnapshotRepo struct {
	db *sql.DB
}

// NewSnapshotRepo creates a SnapshotRepo backed by db.
func NewSnapshotRepo(db *sql.DB) *SnapshotRepo {
	return &SnapshotRepo{db: db}
}

// Insert writes a snapshot row for the given session.
func (r *SnapshotRepo) Insert(ctx context.Context, sessionID int64, s channel.ChannelState, t time.Time) error {
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

// ListByChannelAndDate returns snapshot rows for the given channel and date range,
// with Changed set to true when any metadata field differs from the previous snapshot.
func (r *SnapshotRepo) ListByChannelAndDate(ctx context.Context, chanID []byte, dayStart, dayEnd time.Time) ([]SnapshotRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			recorded_at, listeners, relays,
			name, genre, description, url, comment, track_title, track_artist,
			LAG(name)         OVER w AS prev_name,
			LAG(genre)        OVER w AS prev_genre,
			LAG(description)  OVER w AS prev_description,
			LAG(url)          OVER w AS prev_url,
			LAG(comment)      OVER w AS prev_comment,
			LAG(track_title)  OVER w AS prev_track_title,
			LAG(track_artist) OVER w AS prev_track_artist
		FROM channel_snapshots
		WHERE channel_id = ? AND recorded_at >= ? AND recorded_at < ?
		WINDOW w AS (PARTITION BY session_id ORDER BY recorded_at)
		ORDER BY recorded_at`,
		chanID, dayStart, dayEnd,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SnapshotRow
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
			return nil, err
		}

		changed := !prevName.Valid ||
			name != prevName.String ||
			genre != prevGenre.String ||
			desc != prevDesc.String ||
			url != prevURL.String ||
			comment != prevComment.String ||
			trackTitle != prevTrackTitle.String ||
			trackArtist != prevTrackArtist.String

		row := SnapshotRow{
			RecordedAt: recordedAt,
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
	return result, rows.Err()
}
