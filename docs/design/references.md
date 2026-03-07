# 参考資料一覧

## `_ref/peercast-yt/` — C++ PeerCast リファレンス実装

- リポジトリ: https://github.com/plonk/peercast-yt
- 用途: PCPプロトコルの詳細・エッジケースの確認
- 主要ファイル:
  - `core/common/pcp.h` — PCPタグ定数定義
  - `core/common/atom.h` / `atom2.h` — atom 読み書きプリミティブ
  - `core/common/servent.h` — 接続ごとのステートマシン
  - `core/common/servmgr.h` — サーバ・接続管理（デフォルトポート: 7144）

---

## `_ref/peercast-yayp/` — Go YP サーバ（過去実装）

- リポジトリ: https://github.com/titagaki/peercast-yayp
- 用途: **httpd サーバ部分の参考実装**（指示があったときに参照）
- 概要: PeerCast Root Mode と連動して動作する YP サーバ。Go 実装。

---

## `_ref/yp4g-html/` — YP4G 実稼働サーバの HTML 出力

現在稼働中の YP サーバ「YP4G」が実際に出力している HTML のスナップショット。
httpd の出力フォーマット・UI 設計の参考として使用する。

| ファイル | 対応 URL | 内容 |
|---|---|---|
| `index.html` | http://bayonet.ddo.jp/sp/ | チャンネル一覧ページ（YP トップ） |
| `getgmt.html` | http://bayonet.ddo.jp/sp/getgmt.php | チャンネル別統計ページ（リスナー数推移など） |

### index.html の構造概要
- `meta name="generator" content="YP4G"` — YP4G 生成
- チャンネル一覧を XHTML で出力
- CSS: `yp.css`（外部）

### getgmt.html の構造概要
- チャンネルごとの統計（時刻・配信時間・リスナー数・配信詳細）をテーブルで出力
- CSS: `getgmt.css`（外部）
- タイトル形式: `<チャンネル名> - Statistics - <YP名>`
- リスナー数フォーマット: `<合計> / <直接接続>` （例: `25 / 21`）
