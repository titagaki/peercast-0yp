# 令和のYP 0yp（れいわいぴー）

**peercast-0yp** は PeerCast ルートサーバ（YP / Yellow Page）の Go 実装です。
PCP（PeerCast Protocol）バイナリプロトコルでクライアントからのチャンネル登録を受け付け、
現在放送中のチャンネル一覧を配信します。

## 機能

- **PCP ルートサーバ**: TCP/7144 で PCP バイナリプロトコルを話す
- **チャンネルリスト配信**: PeerCast プレイヤー向け `index.txt`（YP4G 互換フォーマット）
- **REST API**: 現在放送中チャンネルおよび配信履歴の JSON API
- **アーカイブ**: 配信セッション・リスナー数推移を MySQL に記録

## 動作環境

- Go 1.26+
- MySQL 8.0+

## クイックスタート

### ビルド

```bash
go build ./...
```

### 設定

```bash
# 設定ファイルを編集（ポート、接続数など）
cp peercast-0yp.toml peercast-0yp.toml.bak  # バックアップ（任意）

# データベース接続情報を環境変数で設定
export DATABASE_DSN="user:pass@tcp(localhost:3306)/peercast0yp?parseTime=true&loc=Local"
```

### 起動

```bash
./peercast-0yp -config peercast-0yp.toml
```

PCP サーバが `:7144`、HTTP サーバが `:80` で起動します。

### PeerCast クライアントの設定

PeerCast の設定画面で「Root Server」（ルートサーバ / YP アドレス）に
このサーバのホスト名を設定してください。デフォルトポートは 7144 です。

## テスト

```bash
go test ./...
go vet ./...
```

## ドキュメント

| ドキュメント | 内容 |
|---|---|
| [docs/architecture.md](docs/architecture.md) | システムアーキテクチャ・コンポーネント構成 |
| [docs/configuration.md](docs/configuration.md) | 設定ファイルリファレンス |
| [docs/HTTP_API.md](docs/HTTP_API.md) | HTTP API エンドポイント仕様 |

### プロトコル仕様（PeerCast）

| ドキュメント | 内容 |
|---|---|
| [docs/protocol/player.md](docs/protocol/player.md) | PeerCast プレイヤーの YP 連携仕様（index.txt・URL 導出） |
| [docs/protocol/YP_CHANNEL_REGISTRATION.md](docs/protocol/YP_CHANNEL_REGISTRATION.md) | YP 登録プロトコル詳細（PCP） |
| [docs/protocol/PCP_SPEC.md](docs/protocol/PCP_SPEC.md) | PCP プロトコル仕様 |

### 設計文書（実装者向け）

| ドキュメント | 内容 |
|---|---|
| [docs/database.md](docs/database.md) | MySQL スキーマ設計 |
| [docs/design/httpd_architecture.md](docs/design/httpd_architecture.md) | httpd 設計詳細 |
| [docs/design/references.md](docs/design/references.md) | 参考資料一覧 |
