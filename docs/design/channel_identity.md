# チャンネル識別と統計の設計方針

## channel_id (GnuID) は配信ごとに変わる

PeerCast の `channel_id`（GnuID、BINARY(16)）はクライアントが配信開始時に生成する値であり、
**同じ配信者が再接続するたびに異なる値になる**。配信者の安定した識別子ではない。

## 統計はチャンネル名単位で集計する

この制約から、履歴・タイムライン・アクティビティの統計はすべて **`channel_name` 単位**で集計する。

- `channel_sessions` の `channel_name` で絞り込む
- `channel_snapshots` は `session_id` → `channel_sessions` を経由して名前で引く
- HTTP API のパラメータも `cn=` (channel name) で受け取る

## 同名の別配信者が混在するリスク

異なる配信者が同じチャンネル名を使った場合、統計が混在する。
これは PeerCast の仕組み上の制約であり、このシステムでは許容している。

## channel_id の位置づけ

`channel_id` は `channel_snapshots` にのみ記録する。

- `channel_sessions` には持たない（配信ごとに変わるため、セッションの識別子として意味を持たない）
- `channel_snapshots.channel_id` は「この時点でどの GnuID が使われていたか」という履歴として保持する

WHERE 句によるフィルタリングには使用しない。インデックスも設けていない。
