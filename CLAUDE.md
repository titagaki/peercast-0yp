# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`peercast-root-shim` is a Go implementation of a PeerCast root server (tracker). It speaks the PCP (PeerCast Protocol) binary protocol to act as the root/tracker node in a PeerCast P2P streaming network.

The repository is in early development — currently only the protocol spec and a C++ reference implementation exist. No Go code has been written yet.

## Repository Layout

- `docs/PCP_SPEC.md` — Authoritative protocol specification (canonical reference, written for this Go implementation)
- `_ref/peercast-yt/` — Git submodule: C++ reference implementation (peercast-yt fork). Browse `core/common/` for protocol details. Key files: `pcp.h` (all tag constants), `servent.h`/`servmgr.h` (connection handling), `atom.h`/`atom2.h` (atom I/O)

## Commands

```bash
go build ./...
go test ./...
go test -run TestFoo ./channel/...   # run a single test
go vet ./...
```

## PCP Protocol Architecture

PCP is a binary TLV protocol over TCP, strictly **little-endian**.

### Atom Wire Format

Every atom has an 8-byte header:
- Bytes 0–3: 4-byte ASCII tag (short names zero-padded, e.g. `"id"` → `{0x69,0x64,0x00,0x00}`)
- Bytes 4–7: dual-purpose length field
  - Bit 31 set → **container atom**: low 31 bits = child count, no payload bytes
  - Bit 31 clear → **data atom**: value = payload byte count

### Key Implementation Rules (from `docs/PCP_SPEC.md`)

1. **Unknown tags must be skipped gracefully**, not treated as errors — skip child atoms or payload bytes.
2. **Strings** are null-terminated on the wire — strip trailing `\0` on read, append `\0` on write (include it in the length).
3. **IP fields** (`ip`, `rip`, `upip`): 4 bytes = IPv4, 16 bytes = IPv6. IPv6 is stored in **reversed byte order**.
4. **`host` atom IP/port pairs**: `ip` and `port` sub-atoms come in pairs (up to 2 pairs per `host` atom).
5. Use Go struct tags (`pcp:"tagname"`) to map PCP tags to struct fields.

### Connection Handshake

```
Client → Server: pcp\n atom (INT payload = protocol version)
Client → Server: helo container
Server → Client: oleh container
          ↕ ongoing: bcst, chan, host, root atoms
```

### Root Server Role

The root server (`root` atom, Section 4.6 of spec) sends directives to clients:
- `uint` — update interval (seconds)
- `chkv` — minimum required client version
- `url`  — PeerCast download URL suffix
- `upd`  — trigger immediate tracker update broadcast
- `next` — seconds until next root packet

Clients only process `root` atoms when they are **not** themselves a root server.

### Broadcast Routing (`bcst`)

Wraps content atoms with TTL/hop routing metadata. The shim must decrement `ttl`, increment `hops`, detect loops via `from` (sender session ID), and forward to appropriate `grp` targets (`0x01`=root, `0x02`=trackers, `0x04`=relays, `0xFF`=all).

## Go Code Structure

```
main.go              — entry point, signal handling
channel/channel.go   — Info, Track, Hit, HitList, Store (thread-safe registry)
channel/channel_test.go
server/server.go     — Server, session, full handshake & read loop, bcst parsing
```

The `pcp` library (`github.com/titagaki/peercast-pcp/pcp`) provides `Atom`, `ID4`,
`GnuID`, `ReadAtom`, atom constructors (`NewParentAtom`, `NewIntAtom`, etc.), and all
protocol tag vars (`PCPHelo`, `PCPOleh`, `PCPBcst`, etc.) and constants
(`PCPErrorQuit`, `PCPHostFlags1Recv`, etc.).

### Connection lifecycle (server side)

```
ReadAtom → pcp\n     (tag == PCPConnect, payload = version uint32)
ReadAtom → helo      → parseHelo → agent, sid, port, ver
WriteAtom ← oleh     (agnt, sid, ver, rip, port)
WriteAtom ← root     (uint, url, chkv, next, asci)  — §3.4 informational
[validate: ver≥1200, sid non-zero, no loop/dup]
WriteAtom ← ok(0)
WriteAtom ← root > upd    — triggers first tracker update from client
loop:
  ReadAtom → bcst    → processBcst → store.AddHit / store.DelHit
  ReadAtom → quit    → return
```

### Key design decisions

- **Send goroutine per session**: writes are serialized per connection; the
  main goroutine only reads. Channel buffer size 16; packets are dropped if
  the client is too slow (non-blocking send in `broadcastRootSettings`).
- **BCID immutability**: `Store.AddHit` rejects updates with a mismatched
  `BroadcastID` once one is registered (ownership check from §5.3).
- **IP encoding**: IPv4 → 4 raw bytes; IPv6 → 16 bytes reversed
  (`encodeIP`/`decodeIP` in server package).
- **Dead-hit cleanup**: `cleanupLoop` calls `store.RemoveDeadHits(180s)`
  every 500 ms; `broadcastLoop` sends `bcst > root > upd` every 120 s.

## Reference Implementation Notes

The C++ reference in `_ref/peercast-yt/` is useful for understanding edge cases:
- `core/common/pcp.h` — all PCP tag constant definitions
- `core/common/atom.h` / `atom2.h` — atom reading/writing primitives
- `core/common/servent.h` — per-connection state machine
- `core/common/servmgr.h` — server/connection management (default port: 7144)
