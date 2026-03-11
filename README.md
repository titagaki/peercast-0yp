# 令和のYP 0yp（れいわいぴー）

**peercast-0yp** は PeerCast ルートサーバ（YP / Yellow Page）の Go 実装です。
PCP（PeerCast Protocol）バイナリプロトコルでクライアントからのチャンネル登録を受け付け、
現在放送中のチャンネル一覧を配信します。

## 機能

- **PCP ルートサーバ**: TCP/7144 で PCP バイナリプロトコルを話す
- **チャンネルリスト配信**: PeerCast プレイヤー向け `index.txt`（YP4G 互換フォーマット）
- **REST API**: 現在放送中チャンネルおよび配信履歴の JSON API
- **アーカイブ**: 配信セッション・リスナー数推移を MySQL に記録

## クイックスタート

```bash
# 起動
docker compose up -d

# ログ確認
docker compose logs -f app

# 再起動（設定変更後など）
docker compose restart app

# 停止
docker compose down
```

PCP サーバが `:7145`、HTTP サーバが `:8080` で起動します。

### 設定

`peercast-0yp.toml` を編集してください。Docker 起動時はホストの設定ファイルがコンテナにマウントされます。

```toml
[pcp]
port = 7145

[http]
port = 8080
yp_name = "0yp"
yp_url  = "https://example.com"
```

データベース接続情報は `docker-compose.yml` の `DATABASE_DSN` 環境変数で設定します。

### PeerCast クライアントの設定

PeerCast の設定画面で「Root Server」（ルートサーバ / YP アドレス）に
このサーバのホスト名を設定してください。デフォルトポートは 7145 です。

## 開発

```bash
go build ./...
go test ./...
go vet ./...
```

### フロントエンドの開発サーバ

```bash
cd web && npm run dev
```

`http://localhost:5173` で起動します。APIリクエストは `http://localhost:8080` にプロキシされます。

### フロントエンドの変更を反映する

フロントエンド（`web/`）は `go:embed` でGoバイナリに埋め込まれます。
変更を反映するにはフロントエンドをビルドしてからDockerイメージを再ビルドしてください。

```bash
cd web && npm run build && cd ..
docker compose build app && docker compose up -d app
```

## ドキュメント

→ [docs/index.md](docs/index.md)

## License

This project is licensed under the GNU General Public License v3.0. Portions of this software are Copyright (C) 2026 ITAGAKI Takayuki
