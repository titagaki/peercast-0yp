# 設定リファレンス

## 設定ファイル

`peercast-0yp.toml`（デフォルトパス）を TOML 形式で記述します。
起動時に `-config <path>` で別パスを指定できます。

```bash
./peercast-0yp -config /etc/peercast-0yp.toml
```

---

## `[pcp]` — PCP ルートサーバ設定

| キー | 型 | デフォルト | 説明 |
|---|---|---|---|
| `port` | int | `7144` | PCP サーバのリッスンポート |
| `max_connections` | int | `100` | 最大同時 CIN 接続数。超過した場合は `UNAVAILABLE` で切断 |
| `update_interval` | int（秒） | `120` | クライアントへの更新要求間隔。クライアントはこの間隔でチャンネル情報を再送する |
| `hit_timeout` | int（秒） | `180` | ヒットのタイムアウト。最終更新からこの秒数が経過したノードはリストから自動削除 |
| `min_client_version` | uint32 | `1200` | 受け入れる PeerCast クライアントの最小バージョン番号。未満の場合は `BADAGENT` で切断 |

---

## `[http]` — HTTP サーバ設定

| キー | 型 | デフォルト | 説明 |
|---|---|---|---|
| `port` | int | `80` | HTTP サーバのリッスンポート |
| `cors_origins` | []string | `[]` | CORS で許可するオリジンのリスト |
| `yp_name` | string | `""` | YP 名。設定すると `index.txt` 末尾にステータス行を出力 |
| `yp_url` | string | `""` | ステータス行のリンク先 URL |

`cors_origins` は開発時に React の開発サーバからアクセスする際に設定します。
本番環境では `[]`（空）のまま運用します（same-origin のため CORS 不要）。

`yp_name` を設定すると `index.txt` の末尾に以下の形式でステータス行が追加されます：

```
0yp◆Status<>000...000<><>https://example.com<><>稼働中<>-9<>-9<>0<>RAW<><><><><>...<>0:00<>click<>Uptime=3:45:21<>0
```

```toml
# 開発時の設定例
cors_origins = ["http://localhost:5173"]

# YP ステータス行
yp_name = "0yp"
yp_url  = "https://example.com"
```

---

## `[database]`（環境変数）

データベース認証情報は TOML ファイルには書かず、環境変数で設定します。
環境変数は TOML の値より優先されます。

| 環境変数 | 説明 |
|---|---|
| `DATABASE_DSN` | MySQL 接続 DSN 文字列 |

**DSN フォーマット:**

```
user:pass@tcp(host:port)/dbname?parseTime=true&loc=Local
```

**設定例:**

```bash
export DATABASE_DSN="peercast:secret@tcp(localhost:3306)/peercast0yp?parseTime=true&loc=Local"
```

> `parseTime=true` と `loc=Local` は必須です。
> `loc=Local` を省略すると DATETIME カラムが UTC として解釈され、
> JST 保存のスキーマと時刻がずれます。

---

## 設定ファイルのサンプル

```toml
[pcp]
port = 7145
max_connections = 100
update_interval = 120  # seconds
hit_timeout = 180      # seconds
min_client_version = 1200

[http]
port = 8080
cors_origins = []  # 本番環境では空に設定
yp_name = "0yp"
yp_url  = "https://example.com"

# Database credentials are read from the DATABASE_DSN environment variable.
```
