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

→ [docs/index.md](docs/index.md)
