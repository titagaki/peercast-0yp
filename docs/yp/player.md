# PeerCast プレイヤーの YP 連携仕様

PeerCast プレイヤーが YP（Yellow Page）サーバとどのようにやり取りするかを定める。
サーバ側はこの仕様に基づいてエンドポイントを実装する必要がある。

---

## 1. index.txt の取得

プレイヤーは設定された YP URL（例: `http://example.com/index.txt`）を HTTP GET で取得し、
チャンネルリストを解析する。

---

## 2. index.txt フォーマット

- エンコード: UTF-8
- 1行 1チャンネル
- フィールドは `<>` 区切り、**19フィールド固定**

```
Name<>ID<>TrackerIP:Port<>URL<>Genre<>Desc<>Listeners<>Relays<>Bitrate<>ContentType<>Artist<>Album<>Title<>Contact<>NameURLEncoded<>Duration<>click<>Comment<>DirectFlag
```

| # | フィールド | 型 | 内容 |
|---|---|---|---|
| 1 | `Name` | string | チャンネル名 |
| 2 | `ID` | string | チャンネル ID（GnuID、32 文字 hex） |
| 3 | `TrackerIP:Port` | string | トラッカーのグローバル IP とポート（例: `1.2.3.4:7144`） |
| 4 | `URL` | string | 配信元 URL |
| 5 | `Genre` | string | ジャンル。`?` を含む場合はリスナー数を隠蔽 |
| 6 | `Desc` | string | 説明 |
| 7 | `Listeners` | int | リスナー数。ジャンルに `?` があれば `-1` |
| 8 | `Relays` | int | リレー数。ジャンルに `?` があれば `-1` |
| 9 | `Bitrate` | int | ビットレート（kbps） |
| 10 | `ContentType` | string | コンテンツタイプ（例: `FLV`, `MKV`, `MP3`） |
| 11 | `Artist` | string | トラック情報：アーティスト |
| 12 | `Album` | string | トラック情報：アルバム |
| 13 | `Title` | string | トラック情報：タイトル |
| 14 | `Contact` | string | トラック情報：コンタクト URL |
| 15 | `NameURLEncoded` | string | チャンネル名を URL エンコードしたもの |
| 16 | `Duration` | string | 配信時間（`H:MM` 形式、トラッカーの uptime から算出） |
| 17 | `click` | string | 固定文字列 `click`（用途不明、互換性のため出力） |
| 18 | `Comment` | string | コメント（DJメッセージ） |
| 19 | `DirectFlag` | `0`/`1` | Direct 接続可能な Hit が存在する場合 `1`、それ以外 `0` |

---

## 3. プレイヤーによる URL 導出

プレイヤーは `index.txt` の URL を基点として、統計ページ・チャットページの URL を自動導出する。
**いずれも `index.txt` と同じパス階層（同一ディレクトリ）に配置する必要がある。**

### 導出ルール

| ページ | URL パターン | パラメータ |
|---|---|---|
| 統計（リスナー数推移） | `getgmt.php?cn={NameURLEncoded}` | `cn`: チャンネル名を URL エンコードしたもの |
| チャット | `chat.php?cn={NameURLEncoded}` | `cn`: チャンネル名を URL エンコードしたもの |

### 導出例

```
index.txt URL : http://example.com/index.txt
チャンネル名  : やの

→ 統計 URL   : http://example.com/getgmt.php?cn=%E3%82%84%E3%81%AE
→ チャット URL: http://example.com/chat.php?cn=%E3%82%84%E3%81%AE
```

`cn` パラメータには `index.txt` フィールド 15（`NameURLEncoded`）の値がそのまま使われる。

---

## 4. 実装上の注意

- `getgmt.php` と `chat.php` のパスは固定。プレイヤーが URL を組み立てるためサーバ側で変更できない。
- チャンネル名の解決には、ライブ中は Store（インメモリ）、終了済みは DB を参照する。
- `chat.php` はチャット機能の実装が必要な場合のみ有効にする。未実装の場合は 404 を返せばよい。
