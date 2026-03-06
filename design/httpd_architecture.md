# httpd サーバ 要件・アーキテクチャ設計

## 概要

`peercast-root-shim`（PCPルートサーバ）を心臓部として、Web UIや
PeerCastプレイヤー向け出力・アーカイブ閲覧を提供する多機能 httpd サーバを
**同一バイナリ**として実装する。

---

## 機能要件

### 機能1: Web ブラウザ用 YP情報一覧
- 現在放送中のチャンネル一覧を表示する React SPA を提供する
- フロントエンド: React（SPAとしてビルドしたものを Go バイナリに埋め込む）
- データ取得: React → JSON API (`GET /api/channels`) → `channel.Store`

### 機能2: PeerCastプレイヤー用チャンネルリスト
- PeerCastプレイヤーが読み込む `index.txt` 形式のテキストを出力する
- React は関与しない（プレイヤーが直接 HTTP で取得するテキストエンドポイント）
- ソース: インメモリの `channel.Store`（リアルタイム）

### 機能3: アーカイブ・統計
- 過去の配信履歴・統計を閲覧できるページを React SPA として提供する
- フロントエンド: 機能1と同一 SPA 内のページ
- データ取得: React → JSON API (`GET /api/history` 等) → MySQL
- 記録する情報:
  - 配信開始・終了時刻（`channel_sessions` テーブル）
  - 1分間隔のリスナー数スナップショット（`channel_snapshots` テーブル）
- チャンネルマスタは持たない。チャンネルID（GnuID）と名前は各テーブルに直接持つ

#### 機能3-a: チャンネル別 配信頻度グラフ（芝生表示）
- GitHub の Contribution Graph のように、日単位の配信有無・頻度をカレンダー状のヒートマップで表示する
- セルの色の濃淡で「その日の配信時間の合計（または配信回数）」を表現する
- 表示単位: 1セル = 1日、横軸 = 週、縦軸 = 曜日（GitHub 形式）
- データ取得: React → `GET /api/channels/{id}/activity` → MySQL 集計 → JSON
- 集計値: **配信時間合計（分）** をセルの濃淡に使用する
- 日付またぎセッションは日付ごとに按分して計上する（例: 23:10〜01:30の配信 → 当日50分・翌日90分）
- 集計クエリは `channel_sessions` の `started_at`/`ended_at` と各日の境界（00:00）で切り分けて計算する
- 芝生グラフのセルをクリックすると、その日の詳細（機能3-b）を表示する

#### 機能3-b: 日別タイムライン表示
- 芝生グラフの日付セルを選択したときに、その日の配信内容をタイムライン形式で表示する
- 表示内容（1行 = 1スナップショット、ただし配信詳細が変化した行のみ明示）:
  - 時刻
  - リスナー数（合計 / 直接接続）
  - 配信詳細（チャンネル名・概要・コメントなど）— 前回から変化したときのみ表示、変化なしは省略
- 参考: `_ref/yp4g-html/getgmt.html`（YP4Gの同等ページ、10分間隔）
- データ取得: React → `GET /api/channels/{id}/timeline?date=YYYYMMDD` → `channel_snapshots` → JSON
- `channel_snapshots` には配信詳細の変化検出に必要なフィールド（名前・概要・コメント・トラック情報）も記録する
- 変化検出は `channel_snapshots` を MySQL の `LAG()` 窓関数でスキャンして行う（変化専用テーブルは持たない）

---

## 非機能要件

- **デプロイ**: 常に単一マシン上で動作する。スケールアウト不要
- **プロセス**: PCPサーバ・HTTPサーバを同一バイナリ・同一プロセスで動かす
- **永続化**: MySQL をアーカイブ・統計の永続化ストアとして使用する
- **リアルタイム状態**: `channel.Store`（インメモリ）を単一の信頼できる情報源とする

---

## アーキテクチャ

### 同一バイナリを選択した理由

| 観点 | 判断 |
|---|---|
| 機能1・2はStoreの直接参照で実装できる | 別プロセス化するとStore公開手段が必要になり複雑化する |
| MySQLが外部にある | 歴史データの共有はDBが解決する。gRPCは不要 |
| 常に同一マシン | 分散の利点がない。gRPCのオーバーヘッドだけが残る |
| 将来の分離 | 要件が変わったとき（別マシン化・独立デプロイ）に分離を検討する |

### コンポーネント構成

```
main.go
  ├── channel.Store          ← 共有インメモリ状態（PCPで更新、APIで参照）
  ├── server.Server          ← PCPルートサーバ (port 7144)
  │     └── channel.Store への AddHit / DelHit
  ├── archive.Recorder       ← Store の変化を観察し MySQL に記録
  └── httpd.Server           ← HTTPサーバ (port 未定)
        ├── GET /api/channels → channel.Store 参照 → JSON（機能1用API）
        ├── GET /api/history  → MySQL 参照 → JSON（機能3用API）
        ├── GET /index.txt    → channel.Store 参照 → index.txt（機能2）
        └── GET /*            → React SPA の静的ファイル（機能1・3 UI）
```

React SPA のビルド成果物は `//go:embed` でバイナリに埋め込み、単一バイナリで完結させる。

### データフロー

```
PeerCastクライアント          Webブラウザ（React SPA）
  │  PCP (TCP)                  │  HTTP
  ▼                             ▼
server.Server             httpd.Server
  │ AddHit / DelHit         ├── GET /api/channels → channel.Store（リアルタイム）
  ▼                         ├── GET /api/history  → MySQL（履歴）
channel.Store               └── GET /index.txt    → channel.Store（機能2）
  │
  │ 変化通知（イベントまたはポーリング）
  ▼
archive.Recorder
  │ INSERT / UPDATE
  ▼
MySQL
```

### archive.Recorder の実装方針（未決定）

Store の変化を MySQL に記録する手段として以下の2案がある：

**A. イベント通知方式**
- `channel.Store` に `AddHit`/`DelHit` のフック（channel や callback）を追加
- Recorder がイベントを受け取ってその都度 INSERT/UPDATE
- リアルタイム性が高い。Store の実装変更が必要

**B. ポーリング差分方式**
- Recorder が一定間隔（例: 1s）で `Store.Snapshot()` を取得し、前回との差分を検出
- Store の変更不要。実装がシンプル
- 精度は polling 間隔に依存する

---

## 未決定事項

| 項目 | 候補・備考 |
|---|---|
| HTTPサーバのポート番号 | **80** |
| `archive.Recorder` の実装方式 | イベント通知 vs ポーリング差分 |
| MySQLのスキーマ設計 | `channel_sessions` / `channel_snapshots` |
| React SPA のビルド・配置方法 | `go:embed` でバイナリ埋め込み（予定） |
| フロントエンドのビルドツール | Vite? Create React App? |
| CORS ポリシー | 開発時は React dev server (localhost:5173 等) からのリクエストを許可する必要あり |
| JSON API の認証・認可 | 公開YPなら不要？管理画面があるなら要検討 |
| `index.txt` フォーマットの詳細仕様 | 下記「index.txt フォーマット仕様」参照（確定） |

---

## index.txt フォーマット仕様

出典: `_ref/peercast-yayp/docs/api.md`

### Content-Type

`text/plain`

### レコード形式

各チャンネルが **1行**、フィールドは `<>` 区切り（全19フィールド）：

```
チャンネル名<>チャンネルID<>トラッカーIP<>コンタクトURL<>ジャンル<>概要<>リスナー数<>リレー数<>ビットレート<>コンテンツタイプ<>トラックアーティスト<>トラックアルバム<>トラックタイトル<>トラックコンタクト<>URLエンコード済みチャンネル名<>配信時間(H:MM)<>click<>コメント<>直接接続可否(0/1)
```

| # | フィールド | 備考 |
|---|---|---|
| 1 | チャンネル名 | |
| 2 | チャンネルID | 32文字 hex（GnuID） |
| 3 | トラッカーIP | |
| 4 | コンタクトURL | |
| 5 | ジャンル | `?` を含む場合はリスナー非表示 |
| 6 | 概要 | |
| 7 | リスナー数 | ジャンルに `?` があれば `-1` |
| 8 | リレー数 | ジャンルに `?` があれば `-1` |
| 9 | ビットレート | |
| 10 | コンテンツタイプ | 例: `FLV`, `MP3` |
| 11 | トラックアーティスト | |
| 12 | トラックアルバム | |
| 13 | トラックタイトル | |
| 14 | トラックコンタクト | |
| 15 | URLエンコード済みチャンネル名 | |
| 16 | 配信時間 | `H:MM` 形式 |
| 17 | click | （用途不明、互換性のため出力） |
| 18 | コメント | |
| 19 | 直接接続可否 | `0` or `1` |

### 末尾のお知らせ行

チャンネルレコードの末尾に、お知らせ情報を同形式で追加できる（peercast-yayp では `information` テーブルから取得）。本実装での要否は未決定。
