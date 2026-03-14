# 令和のYP 0yp（れいわいぴー）

**peercast-0yp** は、既存の PeerCast ルートサーバに依存しない、自己完結型の YP（Yellow Page）サーバです。
Go で実装されており、PCP（PeerCast Protocol）バイナリプロトコルによるチャンネル登録の受付から
チャンネルリストの配信・放送履歴の公開まで、YP として必要な機能をすべて内包しています。

## 機能

- **独立した PCP ルートサーバ**
  外部の PeerCast ルートサーバに依存せず、TCP/7144 で PCP バイナリプロトコルを直接処理します。
- **チャンネルリスト配信**
  PeerCast プレイヤー向けに `index.txt`（YP4G 互換フォーマット）を配信します。
- **放送履歴ページ**
  配信セッションやリスナー数の推移を記録し、他の YP と同様に放送履歴を閲覧できるページを提供します。

## クイックスタート

```bash
# 起動（開発: MariaDB コンテナを含む）
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# 起動（本番: ホストの mysqld に接続）
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# ログ確認
docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f app

# 再起動（設定変更後など）
docker compose -f docker-compose.yml -f docker-compose.dev.yml restart app

# 停止
docker compose -f docker-compose.yml -f docker-compose.dev.yml down
```

PCP サーバが `:7144`、HTTP（Caddy）が `:80` で起動します。
YP は `/yp/` 配下で配信されます（例: `http://example.com/yp/`）。

### 設定

`peercast-0yp.toml` を編集してください。Docker 起動時はホストの設定ファイルがコンテナにマウントされます。

```toml
[pcp]
port = 7144

[http]
port = 80
yp_name      = "0yp"
yp_url       = "https://example.com/yp/"
yp_index_url = "https://example.com/yp/index.txt"
pcp_address  = "pcp://example.com/"
```

データベース接続情報やポート番号は `.env` ファイルで設定します。

```bash
cp .env.example .env
# .env を編集して各値を設定してください
```

`.env` の `SITE_DOMAIN` に公開ドメインを設定すると、Caddy が Let's Encrypt で HTTPS 証明書を自動取得します。

### PeerCast クライアントの設定

PeerCast の設定画面で「Root Server」（ルートサーバ / YP アドレス）に
このサーバのホスト名を設定してください。デフォルトポートは 7144 です。

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

`http://localhost:5173/yp/` で起動します。APIリクエストは `http://localhost:8080` にプロキシされます。

### フロントエンドの変更を反映する

フロントエンド（`web/`）は `go:embed` でGoバイナリに埋め込まれます。
変更を反映するにはフロントエンドをビルドしてからDockerイメージを再ビルドしてください。

```bash
cd web && npm run build && cd ..
docker compose -f docker-compose.yml -f docker-compose.dev.yml build app
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d app
```

## ドキュメント

→ [docs/index.md](docs/index.md)

## License

This project is licensed under the GNU General Public License v3.0. Portions of this software are Copyright (C) 2026 ITAGAKI Takayuki
