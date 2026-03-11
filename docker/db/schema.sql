CREATE TABLE IF NOT EXISTS channel_sessions (
    id           BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
    channel_name VARCHAR(255)      NOT NULL,
    genre        VARCHAR(255)      NOT NULL DEFAULT '',
    url          VARCHAR(255)      NOT NULL DEFAULT '',
    description  VARCHAR(255)      NOT NULL DEFAULT '',
    comment      VARCHAR(255)      NOT NULL DEFAULT '',
    content_type VARCHAR(32)       NOT NULL DEFAULT '',
    started_at   DATETIME          NOT NULL,
    ended_at     DATETIME          NULL,        -- NULL = 配信中

    PRIMARY KEY (id),
    INDEX idx_channel_name (channel_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS channel_snapshots (
    id               BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
    session_id       BIGINT UNSIGNED   NOT NULL,  -- channel_sessions.id
    channel_id       BINARY(16)        NOT NULL,  -- 履歴参照用
    recorded_at      DATETIME          NOT NULL,
    listeners        SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    relays           SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    age              MEDIUMINT UNSIGNED NOT NULL DEFAULT 0,  -- tracker hit の UpTime (秒)
    name             VARCHAR(255)      NOT NULL DEFAULT '',
    bitrate          SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    genre            VARCHAR(255)      NOT NULL DEFAULT '',
    url              VARCHAR(255)      NOT NULL DEFAULT '',
    description      VARCHAR(255)      NOT NULL DEFAULT '',
    comment          VARCHAR(255)      NOT NULL DEFAULT '',
    content_type     VARCHAR(32)       NOT NULL DEFAULT '',
    hidden_listeners BOOLEAN           NOT NULL DEFAULT 0,
    track_title      VARCHAR(255)      NOT NULL DEFAULT '',
    track_artist     VARCHAR(255)      NOT NULL DEFAULT '',
    track_contact    VARCHAR(255)      NOT NULL DEFAULT '',
    track_album      VARCHAR(255)      NOT NULL DEFAULT '',

    PRIMARY KEY (id),
    INDEX idx_session_time (session_id, recorded_at),
    UNIQUE INDEX idx_recorded_at_name (recorded_at, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
