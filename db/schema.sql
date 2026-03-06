CREATE TABLE IF NOT EXISTS channel_sessions (
    id           BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
    channel_id   BINARY(16)        NOT NULL,
    channel_name VARCHAR(255)      NOT NULL,
    bitrate      SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    content_type VARCHAR(32)       NOT NULL DEFAULT '',
    genre        VARCHAR(255)      NOT NULL DEFAULT '',
    url          VARCHAR(255)      NOT NULL DEFAULT '',
    started_at   DATETIME          NOT NULL,
    ended_at     DATETIME          NULL,        -- NULL = 配信中

    PRIMARY KEY (id),
    INDEX idx_channel_period (channel_id, started_at, ended_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS channel_snapshots (
    id               BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
    session_id       BIGINT UNSIGNED   NOT NULL,  -- channel_sessions.id
    channel_id       BINARY(16)        NOT NULL,  -- 検索用に非正規化
    recorded_at      DATETIME          NOT NULL,
    listeners        SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    relays           SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    name             VARCHAR(255)      NOT NULL DEFAULT '',
    bitrate          SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    content_type     VARCHAR(32)       NOT NULL DEFAULT '',
    genre            VARCHAR(255)      NOT NULL DEFAULT '',
    description      VARCHAR(255)      NOT NULL DEFAULT '',
    url              VARCHAR(255)      NOT NULL DEFAULT '',
    comment          VARCHAR(255)      NOT NULL DEFAULT '',
    hidden_listeners BOOLEAN           NOT NULL DEFAULT 0,
    track_title      VARCHAR(255)      NOT NULL DEFAULT '',
    track_artist     VARCHAR(255)      NOT NULL DEFAULT '',
    track_album      VARCHAR(255)      NOT NULL DEFAULT '',
    track_contact    VARCHAR(255)      NOT NULL DEFAULT '',

    PRIMARY KEY (id),
    INDEX idx_channel_time (channel_id, recorded_at),
    INDEX idx_session      (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
