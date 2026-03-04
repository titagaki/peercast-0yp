# E2E Test Plan: PCP Channel Registration Flow

## 対象範囲

YPサーバを実際に起動し、TCPクライアントがPCPプロトコルで接続・bcstを送信して
channel情報がStoreに積まれるまでの一連の流れを検証する。
テストファイルは `server/server_test.go`（`package server_test`）に記述する。

## テスト基盤

`server_test.go` に既にある共通ヘルパー：

| ヘルパー | 役割 |
|---|---|
| `startServer(t)` | サーバ起動・ランダムポートでListen・t.Cleanup登録 |
| `dial(t, ln)` | TCPコネクション確立 |
| `doHandshake(t, conn, sessID, ver)` | pcp\n → helo → read(oleh/root/ok or quit) |
| `sendBcst(t, conn, chanID, sessID, recv)` | bcst(chan+host)送信 |
| `makeID(b)` | テスト用GnuID生成 |

---

## グループ H: ハンドシェイク

### H-01 正常接続 ✅ 実装済み
**目的**: 有効なhelo送信でoleh/root/ok/root>updが返ること

手順:
1. `startServer` → `dial`
2. `doHandshake(sessID=0x11, ver=1218)`
3. ok atomを受信
4. root>upd atomを受信

期待: ok tag = `PCPOk`、続くatom tag = `PCPRoot`

---

### H-02 バージョン不足 ✅ 実装済み
**目的**: ver < 1200 のクライアントをBadAgentで拒否

手順:
1. `doHandshake(ver=1000)`

期待: quit atom、code = `PCPErrorQuit + PCPErrorBadAgent`

---

### H-03 セッションID空 ✅ 実装済み
**目的**: zero GnuID を NotIdentified で拒否

手順:
1. `doHandshake(sessID=zero, ver=1218)`

期待: quit atom、code = `PCPErrorQuit + PCPErrorNotIdentified`

---

### H-04 重複セッション ✅ 実装済み
**目的**: 同一SIDで2接続目をAlreadyConnectedで拒否

手順:
1. conn1 で `doHandshake(sessID=0x33)` → ok → root>upd を消費（接続維持）
2. conn2 で `doHandshake(sessID=0x33)`

期待: conn2 で quit、code = `PCPErrorQuit + PCPErrorAlreadyConnected`

---

### H-05 ループバック検出 ☐ 未実装
**目的**: サーバ自身のSIDと同じSIDを送るとサイレントクローズされること

前提: `server.Server` にSessionIDを取得するメソッド（`SessionID() GnuID` など）が必要、
または `Server.New` を呼び出した後に `srv.SessionID` フィールドを公開する。

手順:
1. `srv, _, ln := startServer(t)`
2. `doHandshake(sessID=srv.SessionID(), ver=1218)`

期待: oleh/root は届くが、その後 ok も quit も来ずに接続がEOFで閉じられる

---

### H-06 接続数上限 ☐ 未実装
**目的**: `maxCINConnections` 到達後の新規接続がUnavailableで拒否されること

手順:
1. `maxCINConnections`（100）個のコネクションを正常にハンドシェイクしてroot>updまで消費し、接続維持
2. 101番目のコネクションで `doHandshake`

期待: quit atom、code = `PCPErrorQuit + PCPErrorUnavailable`

備考: テストが重いため `maxCINConnections` をテスト用に小さい値（例: 3）に設定
できる構成にするか、`Server` に接続数上書きオプションを追加することを検討する。

---

### H-07 oleh内容の検証 ☐ 未実装
**目的**: oleh の各サブアトムに正しい値がセットされていること

手順:
1. `doHandshake` を改修してoleh atomを返すようにする
2. oleh の子アトムを検査

期待:
- `PCPHeloAgent` = `"PeerCastRoot/0.1 (Go)"`
- `PCPHeloSessionID` = サーバのSID（Server.SessionID()と一致）
- `PCPHeloVersion` = 1218
- `PCPHeloRemoteIP` = `127.0.0.1`（4バイト、ループバック接続なので）
- `PCPHeloPort` = クライアントが送ったport値（7144）

---

### H-08 root atom内容の検証 ☐ 未実装
**目的**: ハンドシェイク中のroot atomに正しいパラメータが含まれていること

手順:
1. `doHandshake` を改修してroot atomを返すようにする
2. root の子アトムを検査

期待:
- `PCPRootUpdInt` = 120（updateInterval秒）
- `PCPRootCheckVer` = 1200（minClientVersion）
- `PCPRootNext` = 120
- `PCPRootUpdate` サブアトムが**ない**（informational rootなのでupd無し）

---

## グループ C: チャンネル登録・管理

### C-01 bcstでチャンネル登録 ✅ 実装済み
**目的**: recv=true の bcst 送信後にStoreにチャンネルが積まれること

手順:
1. 正常ハンドシェイク後、`sendBcst(chanID, sessID, recv=true)`
2. Storeをポーリングしてchanが出現するのを待つ

期待: `store.Snapshot()[chanID]` が存在、Name="Test Channel"

---

### C-02 recv=falseでチャンネル削除 ✅ 実装済み
**目的**: recv=false bcstでhitが削除されること

手順:
1. `sendBcst(recv=true)` でチャンネル登録
2. `sendBcst(recv=false)` で削除
3. Storeをポーリング

期待: `store.Snapshot()[chanID]` が消える

---

### C-03 hit更新（同SIDで再bcst） ☐ 未実装
**目的**: 同じクライアントが同じチャンネルを再度bcstしてもhit数が増えず、情報だけ更新されること

手順:
1. `sendBcst(chanID, sessID, recv=true, numListeners=5)`
2. `sendBcst(chanID, sessID, recv=true, numListeners=10)` （リスナー数が変わった）
3. ポーリング

期待:
- `len(hl.Hits)` = 1（重複しない）
- `hl.Hits[0].NumListeners` = 10（更新されている）

備考: `sendBcst` を `numListeners` を受け取る形に拡張する。

---

### C-04 複数チャンネルの同時登録 ☐ 未実装
**目的**: 1クライアントが複数のchanIDを登録できること

手順:
1. 正常ハンドシェイク後
2. chanID-A, chanID-B, chanID-C それぞれに `sendBcst(recv=true)` を送信
3. ポーリング

期待: Snapshot に3チャンネルすべてが存在

---

### C-05 複数クライアントによる登録 ☐ 未実装
**目的**: 異なるクライアントがそれぞれ別のチャンネルを登録したとき、Storeに全て見えること

手順:
1. clientA: 正常ハンドシェイク → `sendBcst(chanID-A, sessA)`
2. clientB: 正常ハンドシェイク → `sendBcst(chanID-B, sessB)`
3. ポーリング

期待: `Snapshot()[chanID-A]` と `Snapshot()[chanID-B]` の両方が存在

---

### C-06 同一チャンネルに複数ホストが登録される ☐ 未実装
**目的**: 異なるクライアントが同じchanIDを送った場合、2つのhitとして積まれること

手順:
1. clientA: `sendBcst(chanID-X, sessA, bcid=0xBC)` で登録
2. clientB: `sendBcst(chanID-X, sessB, bcid=0xBC)` で登録（同一bcid）
3. ポーリング

期待: `Snapshot()[chanID-X].Hits` の len = 2

---

### C-07 BCID immutability (e2e) ☐ 未実装
**目的**: 別クライアントが異なるBCIDで同一チャンネルを乗っ取ろうとしても拒否されること

手順:
1. clientA: `sendBcst(chanID-X, sessA, bcid=0xAA)` で登録（Name="Original"）
2. clientB: `sendBcst(chanID-X, sessB, bcid=0xBB)` を送信（Name="Hijacked"）
3. ポーリング

期待:
- `Snapshot()[chanID-X].Info.Name` = "Original"（上書きされない）
- `len(Hits)` = 1（clientBのhitが追加されない）

---

### C-08 bcstのchanIDサブアトムでhostのchanIDを解決 ☐ 未実装
**目的**: bcstにPCPBcstChanIDがあり、hostにPCPHostChanIDがない場合でも正しいchanIDに紐付けられること

手順:
1. bcstに `PCPBcstChanID = chanID-X` を含め、hostアトムには `PCPHostChanID` を含めない
2. chanアトムも含めて送信
3. ポーリング

期待: `Snapshot()[chanID-X]` が存在

---

## グループ D: 切断・タイムアウト

### D-01 quit atomで接続が閉じられること ☐ 未実装
**目的**: クライアントがquit atomを送ると接続がサーバ側から閉じられること

手順:
1. 正常ハンドシェイク（root>updまで消費）
2. `pcp.NewIntAtom(pcp.PCPQuit, 0).Write(conn)`
3. 次の `pcp.ReadAtom(conn)` を試みる

期待: ReadAtom が io.EOF または net.Conn closed エラーを返す

---

### D-02 切断後のhit残留 ☐ 未実装
**目的**: TCPコネクションが切断されてもhitはStoreに残ること（hitTimeoutまで）

手順:
1. 正常ハンドシェイク → `sendBcst(recv=true)`
2. chanがStoreに出現するのを確認
3. `conn.Close()` でTCPを強制切断
4. 少し待ってからSnapshot確認

期待: `Snapshot()[chanID]` はまだ存在（180s経つまでは消えない）

---

### D-03 dead hit cleanup ☐ 未実装
**目的**: cleanupLoopが期限切れのhitを定期的に削除すること

手順:
1. `sendBcst(recv=true)` で登録
2. Storeに直接アクセスして `hit.LastSeen` を古い時刻に書き換える
   （または `RemoveDeadHits(0)` を直接呼ぶ）
3. 500ms待ってSnapshot確認

期待: hitが削除されている

備考: cleanupLoopのtickは500msでhitTimeoutは180sのため、実際の経過時間でテストすると
遅くなる。Store直接操作によるショートカットを検討する。

---

## グループ B: サーバからのブロードキャスト

### B-01 接続成功後にroot>updが届く ✅ 実装済み（H-01の一部）
**目的**: ok の直後にサーバから root>upd が届くこと

手順: H-01参照

期待: ok の次の atom の tag = `PCPRoot`、かつ子に `PCPRootUpdate` を含む

---

### B-02 broadcastLoopからの定期root>upd ☐ 未実装
**目的**: updateInterval（120s）ごとにサーバから bcst > root > upd が届くこと

手順:
1. 正常ハンドシェイク後、root>updを消費
2. (テスト用に短い) updateIntervalを設定してサーバを起動
3. 次のatom受信を待つ

期待: 受信したatom tag = `PCPBcst`、内部に `PCPRoot > PCPRootUpdate` が含まれる

備考: 実際の120sを待つのはCIに向かないため、`Server` にupdateIntervalを注入可能に
する構成変更が必要。現実的なInterval（例: 200ms）でテストする。

---

## まとめ: 実装優先順位

| 優先度 | ID | 理由 |
|---|---|---|
| 高 | C-03 | リグレッション防止：重複bcstでhitが膨らむバグを防ぐ |
| 高 | C-04 | 基本ユースケース：1クライアント複数チャンネル |
| 高 | C-05 | 基本ユースケース：マルチクライアント |
| 高 | C-06 | 基本ユースケース：同チャンネルに複数ホスト |
| 高 | C-07 | セキュリティ：BCID乗っ取り防止の確認 |
| 中 | H-05 | プロトコル正確性：ループバック検出 |
| 中 | H-07 | プロトコル正確性：oleh内容の正しさ |
| 中 | H-08 | プロトコル正確性：root atom内容の正しさ |
| 中 | D-01 | 切断処理：quit atomハンドリング |
| 中 | D-02 | 切断処理：hit残留の確認 |
| 低 | H-06 | 運用：接続数上限（構成変更が必要） |
| 低 | C-08 | エッジケース：bcstChanIDルーティング |
| 低 | B-02 | 定期ブロードキャスト（構成変更が必要） |
| 低 | D-03 | dead hit cleanup（Store内部操作が必要） |
