# MySQLスキーマ設計

## 方針

- DATETIME はすべて JST で保存（単一リージョンサービスのためTZ変換不要）
- チャンネルマスタは持たない。名前等は各テーブルに直接持つ

---

## テーブル: `channel_sessions`

チャンネルがインメモリStoreに存在している期間（＝配信セッション）を記録する。
芝生グラフの配信時間合計の集計元。

```sql
CREATE TABLE channel_sessions (
    id           BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    channel_name VARCHAR(255)     NOT NULL,
    genre        VARCHAR(255)     NOT NULL DEFAULT '',
    url          VARCHAR(255)     NOT NULL DEFAULT '',
    description  VARCHAR(255)     NOT NULL DEFAULT '',
    comment      VARCHAR(255)     NOT NULL DEFAULT '',
    content_type VARCHAR(32)      NOT NULL DEFAULT '',
    started_at   DATETIME         NOT NULL,
    ended_at     DATETIME         NULL,        -- NULL = 配信中

    PRIMARY KEY (id),
    INDEX idx_channel_name (channel_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### カラム説明

| カラム | 元フィールド | 説明 |
|---|---|---|
| `channel_name` | `Info.Name` | セッション開始時点のチャンネル名 |
| `genre` | `Info.Genre`（パース後） | YPプレフィックス除去後のジャンル |
| `url` | `Info.URL` | コンタクトURL |
| `description` | `Info.Desc` | 概要 |
| `comment` | `Info.Comment` | コメント |
| `content_type` | `Info.ContentType` | コンテンツタイプ（FLV, MKV等） |
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
    channel_id     BINARY(16)        NOT NULL,  -- 履歴参照用
    recorded_at    DATETIME          NOT NULL,

    -- リスナー数（全Hitのsum）
    listeners      SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    relays         SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    age            MEDIUMINT UNSIGNED NOT NULL DEFAULT 0,  -- tracker hit の UpTime (秒)

    -- 配信詳細（LAG()による変化検出用）
    name           VARCHAR(255)      NOT NULL DEFAULT '',
    bitrate        SMALLINT UNSIGNED NOT NULL DEFAULT 0,
    genre          VARCHAR(255)      NOT NULL DEFAULT '',
    url            VARCHAR(255)      NOT NULL DEFAULT '',
    description    VARCHAR(255)      NOT NULL DEFAULT '',
    comment        VARCHAR(255)      NOT NULL DEFAULT '',
    content_type   VARCHAR(32)       NOT NULL DEFAULT '',
    hidden_listeners BOOLEAN         NOT NULL DEFAULT 0,
    track_title    VARCHAR(255)      NOT NULL DEFAULT '',
    track_artist   VARCHAR(255)      NOT NULL DEFAULT '',
    track_contact  VARCHAR(255)      NOT NULL DEFAULT '',
    track_album    VARCHAR(255)      NOT NULL DEFAULT '',

    PRIMARY KEY (id),
    INDEX idx_session_time (session_id, recorded_at),
    UNIQUE INDEX idx_recorded_at_name (recorded_at, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### カラム説明

| カラム | 元フィールド | 説明 |
|---|---|---|
| `session_id` | — | 所属セッション（`channel_sessions.id`） |
| `channel_id` | `Info.ID` | 配信時点の GnuID（履歴参照用） |
| `listeners` | `Hit.NumListeners` の合計 | 全Hitのリスナー数合計 |
| `relays` | `Hit.NumRelays` の合計 | 全Hitのリレー数合計 |
| `age` | tracker `Hit.UpTime` | 配信者が報告する配信経過秒数 |
| `name` | `Info.Name` | チャンネル名 |
| `bitrate` | `Info.Bitrate` | ビットレート (kbps) |
| `genre` | `Info.Genre`（パース後） | YPプレフィックス除去後のジャンル |
| `url` | `Info.URL` | コンタクトURL |
| `description` | `Info.Desc` | 概要 |
| `comment` | `Info.Comment` | コメント |
| `content_type` | `Info.ContentType` | コンテンツタイプ |
| `hidden_listeners` | ジャンル `?` フラグ | リスナー数非表示フラグ |
| `track_title` | `Track.Title` | トラックタイトル |
| `track_artist` | `Track.Artist` | トラックアーティスト |
| `track_contact` | `Track.Contact` | トラックコンタクト |
| `track_album` | `Track.Album` | トラックアルバム |

### タイムライン取得クエリ（概要）

`LAG()` 窓関数で前レコードと比較し、配信詳細が変化した行のみ詳細を返す。

```sql
SELECT
    ch.recorded_at,
    ch.listeners,
    ch.relays,
    ch.name,
    ch.genre,
    ch.description,
    ch.url,
    ch.comment,
    ch.track_title,
    ch.track_artist,
    LAG(ch.name)         OVER w AS prev_name,
    LAG(ch.genre)        OVER w AS prev_genre,
    LAG(ch.description)  OVER w AS prev_description,
    LAG(ch.url)          OVER w AS prev_url,
    LAG(ch.comment)      OVER w AS prev_comment,
    LAG(ch.track_title)  OVER w AS prev_track_title,
    LAG(ch.track_artist) OVER w AS prev_track_artist
FROM channel_snapshots ch
JOIN channel_sessions cs ON ch.session_id = cs.id
WHERE cs.channel_name = ? AND ch.recorded_at >= ? AND ch.recorded_at < ?
WINDOW w AS (PARTITION BY ch.session_id ORDER BY ch.recorded_at)
ORDER BY ch.recorded_at;
```

アプリ側で `name != prev_name` 等を判定して変化行のみ配信詳細を出力する。
