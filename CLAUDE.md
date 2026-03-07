# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`peercast-0yp` is a Go implementation of a PeerCast root server (YP / Yellow Page). It speaks the PCP binary protocol to accept channel registrations from PeerCast clients and serves channel listings over HTTP.

## Commands

```bash
go build ./...
go test ./...
go test -run TestFoo ./internal/server/...
go vet ./...
```

## Code Structure

```
main.go                  — entry point, wires all components
internal/channel/        — Info, Track, Hit, HitList, Store (thread-safe registry)
internal/server/         — PCP root server: handshake, bcst parsing, session management
internal/archive/        — Recorder: polls Store every 1s, writes sessions/snapshots to MySQL
internal/httpd/          — chi HTTP server: index.txt, /api/* endpoints
internal/repository/     — MySQL access (SessionRepo, SnapshotRepo)
internal/config/         — TOML config loader
```

## Key Rules

- **Genre `yp` prefix**: channels whose `Genre` does not start with `yp` are not registered to this YP — exclude silently from `Store.AddHit` and `index.txt`. See `docs/protocol/genre.md`.
- **BCID immutability**: `Store.AddHit` rejects a mismatched `BroadcastID` once one is set (channel ownership check).
- **IP encoding**: IPv4 → 4 raw bytes; IPv6 → 16 bytes reversed. See `encodeIP`/`decodeIP` in `internal/server/`.
- **`index.txt` ordering**: must use `Store.SnapshotOrdered()` (registration order), not `Snapshot()`.

## Unimplemented

HTTP endpoints specified in `docs/HTTP_API.md` but not yet routed in `internal/httpd/server.go`:

- `GET /getgmt.php?cn=` — per-channel stats HTML page
- `GET /chat.php?cn=` — chat page (return 404 until implemented)
- React SPA frontend (planned as `go:embed`)

## Documentation

| 内容 | 場所 |
|---|---|
| システム構成・データフロー | `docs/architecture.md` |
| 設定リファレンス | `docs/configuration.md` |
| HTTP API 仕様 | `docs/HTTP_API.md` |
| DBスキーマ | `docs/database.md` |
| PCP プロトコル仕様 | `docs/protocol/PCP_SPEC.md` |
| YP 登録フロー詳細 | `docs/protocol/YP_CHANNEL_REGISTRATION.md` |
| プレイヤー連携仕様（index.txt・URL導出） | `docs/protocol/player.md` |
| ジャンルフォーマット | `docs/protocol/genre.md` |
| チャンネル識別と統計の設計方針 | `docs/design/channel_identity.md` |
| 設計意思決定・参考資料 | `docs/design/` |

## Reference Implementation

`_ref/peercast-yt/` — C++ PeerCast (git submodule). Key files:
- `core/common/pcp.h` — all PCP tag constants
- `core/common/atom.h` / `atom2.h` — atom I/O primitives
- `core/common/servent.h` — per-connection state machine
