# HTTP API 仕様

## エンドポイント一覧

### `GET /index.txt`

PeerCast プレイヤーが読み込むチャンネルリスト。YP4G 互換フォーマット。

**レスポンス形式（1行1チャンネル、`<>` 区切り、19フィールド）:**

```
Name<>ID<>TrackerIP:Port<>URL<>Genre<>Desc<>Listeners<>Relays<>Bitrate<>ContentType<>Artist<>Album<>Title<>Contact<>NameURLEncoded<>Duration<>click<>Comment<>DirectFlag
```

| フィールド | 内容 |
|---|---|
| Name | チャンネル名 |
| ID | チャンネルID（32文字hex） |
| TrackerIP:Port | トラッカーのIPアドレスとポート |
| URL | 配信元URL |
| Genre | ジャンル（`?` を含む場合リスナー数を隠蔽） |
| Desc | 説明 |
| Listeners | リスナー数（隠蔽時は `-1`） |
| Relays | リレー数（隠蔽時は `-1`） |
| Bitrate | ビットレート（kbps） |
| ContentType | コンテンツタイプ（例: `FLV`, `MP3`） |
| Artist | トラック情報：アーティスト |
| Album | トラック情報：アルバム |
| Title | トラック情報：タイトル |
| Contact | トラック情報：コンタクトURL |
| NameURLEncoded | チャンネル名をURLエンコードしたもの |
| Duration | 配信時間（`H:MM` 形式） |
| `click` | 固定文字列 |
| Comment | コメント |
| DirectFlag | Direct接続可能なhitが存在する場合 `1`、それ以外 `0` |

---

### `GET /getgmt.php?cn={channel_name}`

**統計ページ（YP4G互換URL）。**

PeerCast プレイヤーは `/index.txt` のベースURLとチャンネル名から統計URLを導出する。
`getgmt.php?cn=` という形式はプレイヤーとの互換性のために固定。

**URL導出ルール（プレイヤー側の動作）:**

```
index.txt URL:  http://example.com/index.txt
チャンネル名:   やの
→ 統計URL:      http://example.com/getgmt.php?cn=%E3%82%84%E3%81%AE
```

| パラメータ | 内容 |
|---|---|
| `cn` | チャンネル名（URLエンコード） |

チャンネル名から Store（ライブ中）または DB（終了済み）でチャンネル ID を解決し、
当日の統計データを HTML でレンダリングして返す。

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
