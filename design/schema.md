# MySQLスキーマ設計

## 方針

- DATETIME はすべて JST で保存（単一リージョンサービスのためTZ変換不要）
- `channel_id` は GnuID（16バイト）を `BINARY(16)` で保存
- チャンネルマスタは持たない。名前等は各テーブルに直接持つ

---

## テーブル: `channel_sessions`

チャンネルがインメモリStoreに存在している期間（＝配信セッション）を記録する。
芝生グラフの配信時間合計の集計元。

```sql
CREATE TABLE channel_sessions (
    id           BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    channel_id   BINARY(16)       NOT NULL,
    channel_name VARCHAR(255)     NOT NULL,
    bitrate      SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    content_type VARCHAR(32)      NOT NULL DEFAULT '',
    genre        VARCHAR(255)     NOT NULL DEFAULT '',
    url          VARCHAR(255)     NOT NULL DEFAULT '',
    started_at   DATETIME         NOT NULL,
    ended_at     DATETIME         NULL,        -- NULL = 配信中

    PRIMARY KEY (id),
    INDEX idx_channel_period (channel_id, started_at, ended_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### カラム説明

| カラム | 元フィールド | 説明 |
|---|---|---|
| `channel_id` | `Info.ID` | GnuID |
| `channel_name` | `Info.Name` | セッション開始時点のチャンネル名 |
| `bitrate` | `Info.Bitrate` | ビットレート (kbps) |
| `content_type` | `Info.ContentType` | コンテンツタイプ（FLV, MKV等） |
| `genre` | `Info.Genre`（パース後） | YPプレフィックス除去後のジャンル |
| `url` | `Info.URL` | コンタクトURL |
| `started_at` | — | Storeに初めて出現した時刻 |
| `ended_at` | — | Storeから消えた時刻。配信中は NULL |

### 日付またぎの按分

芝生グラフの集計クエリは、`started_at`〜`ended_at` を日付境界（00:00）で切り分けて日ごとに分配する。
例: `2026-03-07 23:10` 〜 `2026-03-08 01:30` → 3/7に50分、3/8に90分。
純粋なSQLでの実装が複雑になる場合はアプリ側で日付ループして計算する。

---

## テーブル: `channel_snapshots`

1分間隔でインメモリStoreの状態を記録したスナップショット。
リスナー数推移グラフと日別タイムライン表示に使用する。

```sql
CREATE TABLE channel_snapshots (
    id             BIGINT UNSIGNED   NOT NULL AUTO_INCREMENT,
    session_id     BIGINT UNSIGNED   NOT NULL,  -- channel_sessions.id
    channel_id     BINARY(16)        NOT NULL,  -- 検索用に非正規化
    recorded_at    DATETIME          NOT NULL,

    -- リスナー数（全Hitのsum）
    listeners      SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    relays         SMALLINT UNSIGNED NOT NULL DEFAULT 0,

    -- 配信詳細（LAG()による変化検出用）
    name           VARCHAR(255)      NOT NULL DEFAULT '',
    bitrate        SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    content_type   VARCHAR(32)       NOT NULL DEFAULT '',
    genre          VARCHAR(255)      NOT NULL DEFAULT '',
    description    VARCHAR(255)      NOT NULL DEFAULT '',
    url            VARCHAR(255)      NOT NULL DEFAULT '',
    comment        VARCHAR(255)      NOT NULL DEFAULT '',
    hidden_listeners BOOLEAN         NOT NULL DEFAULT 0,
    track_title    VARCHAR(255)      NOT NULL DEFAULT '',
    track_artist   VARCHAR(255)      NOT NULL DEFAULT '',
    track_album    VARCHAR(255)      NOT NULL DEFAULT '',
    track_contact  VARCHAR(255)      NOT NULL DEFAULT '',

    PRIMARY KEY (id),
    INDEX idx_channel_time (channel_id, recorded_at),
    INDEX idx_session      (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### カラム説明

| カラム | 元フィールド | 説明 |
|---|---|---|
| `session_id` | — | 所属セッション（`channel_sessions.id`） |
| `channel_id` | `Info.ID` | タイムライン検索用に非正規化 |
| `listeners` | `Hit.NumListeners` の合計 | 全Hitのリスナー数合計 |
| `relays` | `Hit.NumRelays` の合計 | 全Hitのリレー数合計 |
| `name` | `Info.Name` | チャンネル名 |
| `bitrate` | `Info.Bitrate` | ビットレート (kbps) |
| `content_type` | `Info.ContentType` | コンテンツタイプ |
| `genre` | `Info.Genre`（パース後） | YPプレフィックス除去後のジャンル |
| `description` | `Info.Desc` | 概要 |
| `url` | `Info.URL` | コンタクトURL |
| `comment` | `Info.Comment` | コメント |
| `hidden_listeners` | ジャンル `?` フラグ | リスナー数非表示フラグ |
| `track_title` | `Track.Title` | トラックタイトル |
| `track_artist` | `Track.Artist` | トラックアーティスト |
| `track_album` | `Track.Album` | トラックアルバム |
| `track_contact` | `Track.Contact` | トラックコンタクト |

### タイムライン取得クエリ（概要）

`LAG()` 窓関数で前レコードと比較し、配信詳細が変化した行のみ詳細を返す。

```sql
SELECT
    recorded_at,
    listeners,
    relays,
    name,
    genre,
    description,
    url,
    comment,
    track_title,
    track_artist,
    LAG(name)        OVER w AS prev_name,
    LAG(genre)       OVER w AS prev_genre,
    LAG(description) OVER w AS prev_description,
    LAG(url)         OVER w AS prev_url,
    LAG(comment)     OVER w AS prev_comment,
    LAG(track_title) OVER w AS prev_track_title,
    LAG(track_artist) OVER w AS prev_track_artist
FROM channel_snapshots
WHERE channel_id = ? AND recorded_at >= ? AND recorded_at < ?
WINDOW w AS (PARTITION BY session_id ORDER BY recorded_at)
ORDER BY recorded_at;
```

アプリ側で `name != prev_name` 等を判定して変化行のみ配信詳細を出力する。
