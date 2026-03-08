package pcp_test

import (
	"context"
	"net"
	"testing"
	"time"

	pcp "github.com/titagaki/peercast-pcp/pcp"

	"github.com/titagaki/peercast-0yp/internal/channel"
	svr "github.com/titagaki/peercast-0yp/internal/pcp"
)

// startServer creates a Server, binds a random local port, starts serving,
// and registers cleanup with t. Returns the server, store, and listener.
func startServer(t *testing.T) (*svr.Server, *channel.Store, net.Listener) {
	t.Helper()
	store := channel.NewStore()
	srv, err := svr.New(store, svr.DefaultConfig())
	if err != nil {
		t.Fatalf("pcp.New: %v", err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		ln.Close()
	})
	go srv.Serve(ctx, ln)
	return srv, store, ln
}

// dial opens a TCP connection to the server's listener address.
func dial(t *testing.T, ln net.Listener) net.Conn {
	t.Helper()
	conn, err := net.DialTCP("tcp", nil, ln.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

// doHandshake performs the PCP connection handshake up through the first
// response atom after the root informational packet. The handshake sequence is:
//  1. Write pcp\n atom (PCPConnect, version=1)
//  2. Write helo container (agnt, ver, id)
//  3. Read oleh  → returned as atom #1
//  4. Read root  → returned as atom #2 (informational, no upd)
//  5. Read ok or quit → returned as atom #3 (the actual result)
//
// On success the caller should read one more root>upd atom.
func doHandshake(t *testing.T, conn net.Conn, sessID pcp.GnuID, ver uint32) *pcp.Atom {
	t.Helper()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// pcp\n
	if err := pcp.NewIntAtom(pcp.PCPConnect, 1).Write(conn); err != nil {
		t.Fatalf("write PCPConnect: %v", err)
	}
	// helo
	helo := pcp.NewParentAtom(pcp.PCPHelo,
		pcp.NewStringAtom(pcp.PCPHeloAgent, "TestClient/1.0"),
		pcp.NewIntAtom(pcp.PCPHeloVersion, ver),
		pcp.NewIDAtom(pcp.PCPHeloSessionID, sessID),
		pcp.NewShortAtom(pcp.PCPHeloPort, 7144),
	)
	if err := helo.Write(conn); err != nil {
		t.Fatalf("write helo: %v", err)
	}

	// Read oleh
	a1, err := pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read oleh: %v", err)
	}
	if a1.Tag != pcp.PCPOleh {
		t.Fatalf("expected oleh, got %v", a1.Tag)
	}

	// Read root (informational)
	a2, err := pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read root: %v", err)
	}
	if a2.Tag != pcp.PCPRoot {
		t.Fatalf("expected root, got %v", a2.Tag)
	}

	// Read ok or quit
	a3, err := pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read ok/quit: %v", err)
	}
	return a3
}

// doHandshakeFull performs the same handshake as doHandshake but returns all
// three atoms (oleh, root, ok-or-quit) for detailed inspection.
func doHandshakeFull(t *testing.T, conn net.Conn, sessID pcp.GnuID, ver uint32) (oleh, root, final *pcp.Atom) {
	t.Helper()
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	if err := pcp.NewIntAtom(pcp.PCPConnect, 1).Write(conn); err != nil {
		t.Fatalf("write PCPConnect: %v", err)
	}
	helo := pcp.NewParentAtom(pcp.PCPHelo,
		pcp.NewStringAtom(pcp.PCPHeloAgent, "TestClient/1.0"),
		pcp.NewIntAtom(pcp.PCPHeloVersion, ver),
		pcp.NewIDAtom(pcp.PCPHeloSessionID, sessID),
		pcp.NewShortAtom(pcp.PCPHeloPort, 7144),
	)
	if err := helo.Write(conn); err != nil {
		t.Fatalf("write helo: %v", err)
	}
	var err error
	oleh, err = pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read oleh: %v", err)
	}
	root, err = pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read root: %v", err)
	}
	final, err = pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read final: %v", err)
	}
	return
}

// sendQuit writes a quit atom on conn.
func sendQuit(t *testing.T, conn net.Conn) {
	t.Helper()
	if err := pcp.NewIntAtom(pcp.PCPQuit, 0).Write(conn); err != nil {
		t.Fatalf("write quit: %v", err)
	}
}

// BcstOptions controls the content of a bcst atom built by sendBcst.
// Zero values use sensible defaults so callers only need to set what matters.
type BcstOptions struct {
	ChanID       pcp.GnuID // required
	SessID       pcp.GnuID // required
	BCID         pcp.GnuID // zero → filled with 0xBC
	Name         string    // default: "Test Channel"
	Bitrate      uint32    // default: 128
	NumListeners uint32    // default: 5
	NumRelays    uint32    // default: 2
	Port         uint16    // default: 7144
	Recv         bool
}

// sendBcst writes a bcst atom advertising a channel with tracker status.
func sendBcst(t *testing.T, conn net.Conn, opts BcstOptions) {
	t.Helper()

	bcid := opts.BCID
	if bcid.IsEmpty() {
		for i := range bcid {
			bcid[i] = 0xBC
		}
	}
	name := opts.Name
	if name == "" {
		name = "Test Channel"
	}
	bitrate := opts.Bitrate
	if bitrate == 0 {
		bitrate = 128
	}
	numListeners := opts.NumListeners
	if numListeners == 0 {
		numListeners = 5
	}
	numRelays := opts.NumRelays
	if numRelays == 0 {
		numRelays = 2
	}
	port := opts.Port
	if port == 0 {
		port = 7144
	}

	flg1 := byte(pcp.PCPHostFlags1Tracker)
	if opts.Recv {
		flg1 |= byte(pcp.PCPHostFlags1Recv)
	}

	ip := net.ParseIP("127.0.0.1").To4()

	chanAtom := pcp.NewParentAtom(pcp.PCPChan,
		pcp.NewIDAtom(pcp.PCPChanID, opts.ChanID),
		pcp.NewIDAtom(pcp.PCPChanBCID, bcid),
		pcp.NewParentAtom(pcp.PCPChanInfo,
			pcp.NewStringAtom(pcp.PCPChanInfoName, name),
			pcp.NewIntAtom(pcp.PCPChanInfoBitrate, bitrate),
			pcp.NewStringAtom(pcp.PCPChanInfoGenre, "ypTest"),
			pcp.NewStringAtom(pcp.PCPChanInfoType, "MP3"),
		),
		pcp.NewParentAtom(pcp.PCPChanTrack),
	)

	hostAtom := pcp.NewParentAtom(pcp.PCPHost,
		pcp.NewIDAtom(pcp.PCPHostID, opts.SessID),
		pcp.NewBytesAtom(pcp.PCPHostIP, ip),
		pcp.NewShortAtom(pcp.PCPHostPort, port),
		pcp.NewIntAtom(pcp.PCPHostNumListeners, numListeners),
		pcp.NewIntAtom(pcp.PCPHostNumRelays, numRelays),
		pcp.NewIntAtom(pcp.PCPHostUptime, 3600),
		pcp.NewByteAtom(pcp.PCPHostFlags1, flg1),
		pcp.NewIDAtom(pcp.PCPHostChanID, opts.ChanID),
	)

	bcst := pcp.NewParentAtom(pcp.PCPBcst,
		pcp.NewByteAtom(pcp.PCPBcstGroup, pcp.PCPBcstGroupTrackers),
		pcp.NewByteAtom(pcp.PCPBcstHops, 0),
		pcp.NewByteAtom(pcp.PCPBcstTTL, 7),
		pcp.NewIDAtom(pcp.PCPBcstFrom, opts.SessID),
		pcp.NewIntAtom(pcp.PCPBcstVersion, 1218),
		pcp.NewIntAtom(pcp.PCPBcstVersionVP, 27),
		chanAtom,
		hostAtom,
	)

	if err := bcst.Write(conn); err != nil {
		t.Fatalf("write bcst: %v", err)
	}
}

// makeID creates a GnuID filled with the given byte value.
func makeID(b byte) pcp.GnuID {
	var id pcp.GnuID
	for i := range id {
		id[i] = b
	}
	return id
}

// waitForChannel polls store until chanID appears or timeout elapses.
func waitForChannel(t *testing.T, store *channel.Store, chanID pcp.GnuID, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, ok := store.Snapshot()[chanID]; ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("channel %v not in store after %v", chanID, timeout)
}

// waitForChannelGone polls store until chanID disappears or timeout elapses.
func waitForChannelGone(t *testing.T, store *channel.Store, chanID pcp.GnuID, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, ok := store.Snapshot()[chanID]; !ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("channel %v still in store after %v", chanID, timeout)
}

// ----------------------------------------------------------------------------
// Integration tests
// ----------------------------------------------------------------------------

func TestHandshake_Success(t *testing.T) {
	_, _, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0x11)
	result := doHandshake(t, conn, sessID, 1218)
	if result.Tag != pcp.PCPOK {
		t.Fatalf("expected ok, got tag %v", result.Tag)
	}

	// Read the root>upd that follows ok on success.
	rootUpd, err := pcp.ReadAtom(conn)
	if err != nil {
		t.Fatalf("read root>upd: %v", err)
	}
	if rootUpd.Tag != pcp.PCPRoot {
		t.Fatalf("expected root atom after ok, got %v", rootUpd.Tag)
	}
}

func TestHandshake_BadVersion(t *testing.T) {
	_, _, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0x22)
	result := doHandshake(t, conn, sessID, 1000) // below minClientVersion=1200
	if result.Tag != pcp.PCPQuit {
		t.Fatalf("expected quit, got tag %v", result.Tag)
	}
	code, _ := result.GetInt()
	want := uint32(pcp.PCPErrorQuit + pcp.PCPErrorBadAgent)
	if code != want {
		t.Errorf("quit code = %d, want %d", code, want)
	}
}

func TestHandshake_NoSessionID(t *testing.T) {
	_, _, ln := startServer(t)
	conn := dial(t, ln)

	var emptySID pcp.GnuID // zero value = empty
	result := doHandshake(t, conn, emptySID, 1218)
	if result.Tag != pcp.PCPQuit {
		t.Fatalf("expected quit, got tag %v", result.Tag)
	}
	code, _ := result.GetInt()
	want := uint32(pcp.PCPErrorQuit + pcp.PCPErrorNotIdentified)
	if code != want {
		t.Errorf("quit code = %d, want %d", code, want)
	}
}

func TestHandshake_DuplicateSession(t *testing.T) {
	_, _, ln := startServer(t)

	sessID := makeID(0x33)

	// First connection — must succeed.
	conn1 := dial(t, ln)
	result1 := doHandshake(t, conn1, sessID, 1218)
	if result1.Tag != pcp.PCPOK {
		t.Fatalf("first connection: expected ok, got %v", result1.Tag)
	}
	// Consume the root>upd so conn1 stays open.
	pcp.ReadAtom(conn1)

	// Second connection with the same session ID — must be rejected.
	conn2 := dial(t, ln)
	result2 := doHandshake(t, conn2, sessID, 1218)
	if result2.Tag != pcp.PCPQuit {
		t.Fatalf("second connection: expected quit, got %v", result2.Tag)
	}
	code, _ := result2.GetInt()
	want := uint32(pcp.PCPErrorQuit + pcp.PCPErrorAlreadyConnected)
	if code != want {
		t.Errorf("quit code = %d, want %d", code, want)
	}
}

func TestChannelRegistration(t *testing.T) {
	_, store, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0x44)
	var chanID pcp.GnuID
	for i := range chanID {
		chanID[i] = byte(i + 1)
	}

	result := doHandshake(t, conn, sessID, 1218)
	if result.Tag != pcp.PCPOK {
		t.Fatalf("handshake: expected ok, got %v", result.Tag)
	}
	pcp.ReadAtom(conn) // consume root>upd

	sendBcst(t, conn, BcstOptions{ChanID: chanID, SessID: sessID, Recv: true})
	waitForChannel(t, store, chanID, 2*time.Second)
}

func TestChannelDeletion(t *testing.T) {
	_, store, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0x55)
	var chanID pcp.GnuID
	for i := range chanID {
		chanID[i] = byte(i + 0x10)
	}

	result := doHandshake(t, conn, sessID, 1218)
	if result.Tag != pcp.PCPOK {
		t.Fatalf("handshake: expected ok, got %v", result.Tag)
	}
	pcp.ReadAtom(conn) // consume root>upd

	// Register the channel.
	sendBcst(t, conn, BcstOptions{ChanID: chanID, SessID: sessID, Recv: true})
	waitForChannel(t, store, chanID, 2*time.Second)

	// Deregister by sending bcst with recv=false.
	sendBcst(t, conn, BcstOptions{ChanID: chanID, SessID: sessID, Recv: false})
	waitForChannelGone(t, store, chanID, 2*time.Second)
}

// C-03: 同クライアントが同チャンネルを再bcstしてもhit数は増えず、情報が更新されること。
func TestHitUpdate_SameClient(t *testing.T) {
	_, store, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0x60)
	chanID := makeID(0x61)

	doHandshake(t, conn, sessID, 1218)
	pcp.ReadAtom(conn) // consume root>upd

	sendBcst(t, conn, BcstOptions{ChanID: chanID, SessID: sessID, Recv: true, NumListeners: 5})
	waitForChannel(t, store, chanID, 2*time.Second)

	sendBcst(t, conn, BcstOptions{ChanID: chanID, SessID: sessID, Recv: true, NumListeners: 10})

	// Poll until NumListeners is updated.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hl, ok := store.Snapshot()[chanID]; ok && hl.Hits[0].NumListeners == 10 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	hl := store.Snapshot()[chanID]
	if len(hl.Hits) != 1 {
		t.Errorf("hit count = %d, want 1 (no dup)", len(hl.Hits))
	}
	if hl.Hits[0].NumListeners != 10 {
		t.Errorf("NumListeners = %d, want 10", hl.Hits[0].NumListeners)
	}
}

// C-04: 1クライアントが複数チャンネルを登録できること。
func TestMultipleChannels_SingleClient(t *testing.T) {
	_, store, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0x70)
	chanA := makeID(0x71)
	chanB := makeID(0x72)
	chanC := makeID(0x73)

	doHandshake(t, conn, sessID, 1218)
	pcp.ReadAtom(conn) // consume root>upd

	sendBcst(t, conn, BcstOptions{ChanID: chanA, SessID: sessID, Recv: true})
	sendBcst(t, conn, BcstOptions{ChanID: chanB, SessID: sessID, Recv: true})
	sendBcst(t, conn, BcstOptions{ChanID: chanC, SessID: sessID, Recv: true})

	waitForChannel(t, store, chanA, 2*time.Second)
	waitForChannel(t, store, chanB, 2*time.Second)
	waitForChannel(t, store, chanC, 2*time.Second)
}

// C-05: 複数クライアントがそれぞれ別チャンネルを登録したとき、両方Storeに見えること。
func TestMultipleChannels_MultipleClients(t *testing.T) {
	_, store, ln := startServer(t)

	sessA := makeID(0x80)
	sessB := makeID(0x81)
	chanA := makeID(0x82)
	chanB := makeID(0x83)

	connA := dial(t, ln)
	doHandshake(t, connA, sessA, 1218)
	pcp.ReadAtom(connA)

	connB := dial(t, ln)
	doHandshake(t, connB, sessB, 1218)
	pcp.ReadAtom(connB)

	sendBcst(t, connA, BcstOptions{ChanID: chanA, SessID: sessA, Recv: true})
	sendBcst(t, connB, BcstOptions{ChanID: chanB, SessID: sessB, Recv: true})

	waitForChannel(t, store, chanA, 2*time.Second)
	waitForChannel(t, store, chanB, 2*time.Second)
}

// C-06: 異なるクライアントが同一チャンネルを登録すると2つのhitとして積まれること。
func TestMultipleHits_SameChannel(t *testing.T) {
	_, store, ln := startServer(t)

	sessA := makeID(0x90)
	sessB := makeID(0x91)
	chanID := makeID(0x92)

	var bcid pcp.GnuID
	for i := range bcid {
		bcid[i] = 0xBC
	}

	connA := dial(t, ln)
	doHandshake(t, connA, sessA, 1218)
	pcp.ReadAtom(connA)

	connB := dial(t, ln)
	doHandshake(t, connB, sessB, 1218)
	pcp.ReadAtom(connB)

	sendBcst(t, connA, BcstOptions{ChanID: chanID, SessID: sessA, BCID: bcid, Port: 7144, Recv: true})
	waitForChannel(t, store, chanID, 2*time.Second)

	sendBcst(t, connB, BcstOptions{ChanID: chanID, SessID: sessB, BCID: bcid, Port: 7145, Recv: true})

	// Poll until 2 hits appear.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hl, ok := store.Snapshot()[chanID]; ok && len(hl.Hits) == 2 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if n := len(store.Snapshot()[chanID].Hits); n != 2 {
		t.Errorf("hit count = %d, want 2", n)
	}
}

// C-07: 別クライアントが異なるBCIDで同一チャンネルを上書きしようとしても拒否されること。
func TestBCIDImmutability(t *testing.T) {
	_, store, ln := startServer(t)

	sessA := makeID(0xA0)
	sessB := makeID(0xA1)
	chanID := makeID(0xA2)

	var bcidA, bcidB pcp.GnuID
	for i := range bcidA {
		bcidA[i] = 0xAA
		bcidB[i] = 0xBB
	}

	connA := dial(t, ln)
	doHandshake(t, connA, sessA, 1218)
	pcp.ReadAtom(connA)

	connB := dial(t, ln)
	doHandshake(t, connB, sessB, 1218)
	pcp.ReadAtom(connB)

	sendBcst(t, connA, BcstOptions{ChanID: chanID, SessID: sessA, BCID: bcidA, Name: "Original", Recv: true})
	waitForChannel(t, store, chanID, 2*time.Second)

	sendBcst(t, connB, BcstOptions{ChanID: chanID, SessID: sessB, BCID: bcidB, Name: "Hijacked", Recv: true})

	// Give the server time to process connB's bcst.
	time.Sleep(50 * time.Millisecond)

	hl := store.Snapshot()[chanID]
	if hl.Info.Name == "Hijacked" {
		t.Error("BCID immutability violated: channel name was overwritten")
	}
	if len(hl.Hits) != 1 {
		t.Errorf("hit count = %d, want 1 (hijacker hit must be rejected)", len(hl.Hits))
	}
}

// H-05: サーバ自身のSIDと同じSIDでのループバック接続はサイレントクローズされること。
func TestHandshake_Loopback(t *testing.T) {
	srv, _, ln := startServer(t)
	conn := dial(t, ln)

	conn.SetDeadline(time.Now().Add(5 * time.Second))
	pcp.NewIntAtom(pcp.PCPConnect, 1).Write(conn)
	pcp.NewParentAtom(pcp.PCPHelo,
		pcp.NewStringAtom(pcp.PCPHeloAgent, "TestClient/1.0"),
		pcp.NewIntAtom(pcp.PCPHeloVersion, 1218),
		pcp.NewIDAtom(pcp.PCPHeloSessionID, srv.SessionID()),
		pcp.NewShortAtom(pcp.PCPHeloPort, 7144),
	).Write(conn)

	// Server sends oleh + root even for loopback, then silently closes.
	if a, err := pcp.ReadAtom(conn); err != nil || a.Tag != pcp.PCPOleh {
		t.Fatalf("expected oleh, got tag=%v err=%v", a.Tag, err)
	}
	if a, err := pcp.ReadAtom(conn); err != nil || a.Tag != pcp.PCPRoot {
		t.Fatalf("expected root, got tag=%v err=%v", a.Tag, err)
	}

	// No ok or quit — connection must close.
	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if a, err := pcp.ReadAtom(conn); err == nil {
		t.Fatalf("expected EOF after loopback, got atom tag=%v", a.Tag)
	}
}

// H-07: oleh の各サブアトムに正しい値がセットされていること。
func TestHandshake_OlehContents(t *testing.T) {
	srv, _, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0xD0)
	oleh, _, final := doHandshakeFull(t, conn, sessID, 1218)

	if final.Tag != pcp.PCPOK {
		t.Fatalf("expected ok, got %v", final.Tag)
	}

	got := map[string]any{}
	for _, child := range oleh.Children() {
		switch child.Tag {
		case pcp.PCPHeloAgent:
			got["agent"] = child.GetString()
		case pcp.PCPHeloSessionID:
			id, _ := child.GetID()
			got["sid"] = id
		case pcp.PCPHeloVersion:
			v, _ := child.GetInt()
			got["ver"] = v
		case pcp.PCPHeloRemoteIP:
			got["rip"] = child.Data()
		case pcp.PCPHeloPort:
			p, _ := child.GetShort()
			got["port"] = p
		}
	}

	if got["agent"] != "PeerCastRoot/0.1 (Go)" {
		t.Errorf("agent = %q, want PeerCastRoot/0.1 (Go)", got["agent"])
	}
	if got["sid"] != srv.SessionID() {
		t.Errorf("sid mismatch: got %v, want %v", got["sid"], srv.SessionID())
	}
	if got["ver"] != uint32(1218) {
		t.Errorf("ver = %v, want 1218", got["ver"])
	}
	wantRIP := []byte{127, 0, 0, 1}
	if rip, ok := got["rip"].([]byte); !ok || len(rip) != 4 || rip[0] != 127 {
		t.Errorf("remoteIP = %v, want %v", got["rip"], wantRIP)
	}
	if got["port"] != uint16(7144) {
		t.Errorf("port = %v, want 7144", got["port"])
	}
}

// H-08: ハンドシェイク中の root atom に正しいパラメータが含まれていること。
// informational root なので PCPRootUpdate サブアトムは含まれない。
func TestHandshake_RootContents(t *testing.T) {
	_, _, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0xE0)
	_, root, final := doHandshakeFull(t, conn, sessID, 1218)

	if final.Tag != pcp.PCPOK {
		t.Fatalf("expected ok, got %v", final.Tag)
	}

	got := map[string]uint32{}
	hasUpd := false
	for _, child := range root.Children() {
		switch child.Tag {
		case pcp.PCPRootUpdInt:
			got["updInt"], _ = child.GetInt()
		case pcp.PCPRootCheckVer:
			got["chkv"], _ = child.GetInt()
		case pcp.PCPRootNext:
			got["next"], _ = child.GetInt()
		case pcp.PCPRootUpdate:
			hasUpd = true
		}
	}

	if got["updInt"] != 120 {
		t.Errorf("updInt = %d, want 120", got["updInt"])
	}
	if got["chkv"] != 1200 {
		t.Errorf("checkVer = %d, want 1200", got["chkv"])
	}
	if got["next"] != 120 {
		t.Errorf("next = %d, want 120", got["next"])
	}
	if hasUpd {
		t.Error("informational root must not contain PCPRootUpdate")
	}
}

// D-01: クライアントがquit atomを送ると接続がサーバ側から閉じられること。
func TestClientQuit(t *testing.T) {
	_, _, ln := startServer(t)
	conn := dial(t, ln)

	sessID := makeID(0xF0)
	result := doHandshake(t, conn, sessID, 1218)
	if result.Tag != pcp.PCPOK {
		t.Fatalf("handshake: expected ok, got %v", result.Tag)
	}
	pcp.ReadAtom(conn) // consume root>upd

	sendQuit(t, conn)

	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	if a, err := pcp.ReadAtom(conn); err == nil {
		t.Fatalf("expected connection closed after quit, got atom tag=%v", a.Tag)
	}
}

// D-02: TCP接続が切断されてもhitはStoreに残ること（hitTimeoutまで）。
func TestHitPersistsAfterDisconnect(t *testing.T) {
	_, store, ln := startServer(t)

	sessID := makeID(0xF1)
	chanID := makeID(0xF2)

	conn, err := net.DialTCP("tcp", nil, ln.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	result := doHandshake(t, conn, sessID, 1218)
	if result.Tag != pcp.PCPOK {
		t.Fatalf("handshake: expected ok, got %v", result.Tag)
	}
	pcp.ReadAtom(conn) // consume root>upd

	sendBcst(t, conn, BcstOptions{ChanID: chanID, SessID: sessID, Recv: true})
	waitForChannel(t, store, chanID, 2*time.Second)

	conn.Close() // forcibly disconnect (no quit atom)
	time.Sleep(50 * time.Millisecond)

	if _, ok := store.Snapshot()[chanID]; !ok {
		t.Fatal("hit was removed immediately after disconnect; should persist until hitTimeout")
	}
}
