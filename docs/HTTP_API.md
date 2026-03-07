# HTTP API 仕様

## エンドポイント一覧

### `GET /index.txt`

PeerCast プレイヤーが読み込むチャンネルリスト。YP4G 互換フォーマット。

フォーマット仕様（19フィールド `<>` 区切り）・プレイヤーによる URL 導出ルールは
[protocol/player.md](protocol/player.md) を参照。

---

### `GET /getgmt.php?cn={channel_name}` ⚠️ 未実装

チャンネルの当日統計を HTML で返す。YP4G 互換 URL。
プレイヤーによる URL 導出ルールは [protocol/player.md](protocol/player.md) を参照。

| パラメータ | 内容 |
|---|---|
| `cn` | チャンネル名（URL エンコード） |

**チャンネル解決:**
チャンネル名で Store（ライブ中）を先に検索し、なければ DB（終了済み）から当日分を検索する。
どちらにも見つからなければ 404。

**レスポンス:** HTML ページ（`text/html`）

表示内容（参考: `_ref/yp4g-html/getgmt.html`）:

| 項目 | 内容 |
|---|---|
| タイトル | `{チャンネル名} - Statistics - 0yp` |
| 各行 | 時刻・リスナー数（合計 / 直接接続）・配信詳細 |
| 配信詳細 | 前のスナップショットから変化した項目のみ表示（名前・概要・コメント・トラック情報） |
| データ粒度 | 1分間隔（`channel_snapshots` テーブル） |

---

### `GET /chat.php?cn={channel_name}` ⚠️ 未実装

チャットページ。プレイヤーが `index.txt` の URL から導出して開く。
実装するまでは 404 を返す。

| パラメータ | 内容 |
|---|---|
| `cn` | チャンネル名（URL エンコード） |

---

### `GET /api/channels`

現在放送中のチャンネル一覧をJSONで返す。

**レスポンス（JSON配列）:**

```json
[
  {
    "id": "0102...",
    "name": "チャンネル名",
    "genre": "Music",
    "desc": "説明",
    "url": "http://...",
    "comment": "コメント",
    "bitrate": 128,
    "contentType": "MP3",
    "track": {
      "title": "曲名",
      "artist": "アーティスト",
      "album": "アルバム",
      "contact": "http://..."
    },
    "tracker": {
      "ip": "1.2.3.4",
      "port": 7144,
      "firewalled": false
    },
    "numListeners": 5,
    "numRelays": 2,
    "upTime": 3600
  }
]
```

---

### `GET /api/history?limit={n}&offset={n}`

過去の配信セッション一覧。

| パラメータ | デフォルト | 最大 |
|---|---|---|
| `limit` | 50 | 200 |
| `offset` | 0 | — |

---

### `GET /api/channels/activity?name={channel_name}`

指定チャンネルの過去365日間の日別放送時間。

| パラメータ | 内容 |
|---|---|
| `name` | チャンネル名 |

---

### `GET /api/channels/timeline?name={channel_name}&date={YYYYMMDD}`

指定チャンネルの特定日のスナップショット履歴（1分刻み）。

| パラメータ | 内容 |
|---|---|
| `name` | チャンネル名 |
| `date` | 日付（`YYYYMMDD` 形式） |
