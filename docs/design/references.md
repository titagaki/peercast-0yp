# 参考資料一覧

## `_ref/peercast-yt/` — C++ PeerCast リファレンス実装

- リポジトリ: https://github.com/plonk/peercast-yt
- 用途: PCPプロトコルの詳細・エッジケースの確認
- 主要ファイル:
  - `core/common/pcp.h` — PCPタグ定数定義
  - `core/common/atom.h` / `atom2.h` — atom 読み書きプリミティブ
  - `core/common/servent.h` — 接続ごとのステートマシン
  - `core/common/servmgr.h` — サーバ・接続管理（デフォルトポート: 7144）
