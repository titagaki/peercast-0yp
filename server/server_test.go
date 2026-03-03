package server_test

import (
	"context"
	"net"
	"testing"
	"time"

	pcp "github.com/titagaki/peercast-pcp/pcp"

	"github.com/megan/peercast-root-shim/channel"
	"github.com/megan/peercast-root-shim/server"
)

// startServer creates a Server, binds a random local port, starts serving,
// and registers cleanup with t. Returns the server, store, and listener.
func startServer(t *testing.T) (*server.Server, *channel.Store, net.Listener) {
	t.Helper()
	store := channel.NewStore()
	srv, err := server.New(store)
	if err != nil {
		t.Fatalf("server.New: %v", err)
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

// sendBcst writes a bcst atom advertising a channel with tracker status.
// recv=true means the tracker is actively receiving (stream is live).
func sendBcst(t *testing.T, conn net.Conn, chanID pcp.GnuID, sessID pcp.GnuID, recv bool) {
	t.Helper()

	var bcid pcp.GnuID
	for i := range bcid {
		bcid[i] = 0xBC
	}

	flg1 := byte(pcp.PCPHostFlags1Tracker)
	if recv {
		flg1 |= byte(pcp.PCPHostFlags1Recv)
	}

	ip := net.ParseIP("127.0.0.1").To4()

	chanAtom := pcp.NewParentAtom(pcp.PCPChan,
		pcp.NewIDAtom(pcp.PCPChanID, chanID),
		pcp.NewIDAtom(pcp.PCPChanBCID, bcid),
		pcp.NewParentAtom(pcp.PCPChanInfo,
			pcp.NewStringAtom(pcp.PCPChanInfoName, "Test Channel"),
			pcp.NewIntAtom(pcp.PCPChanInfoBitrate, 128),
			pcp.NewStringAtom(pcp.PCPChanInfoGenre, "Test"),
			pcp.NewStringAtom(pcp.PCPChanInfoType, "MP3"),
		),
		pcp.NewParentAtom(pcp.PCPChanTrack),
	)

	hostAtom := pcp.NewParentAtom(pcp.PCPHost,
		pcp.NewIDAtom(pcp.PCPHostID, sessID),
		pcp.NewBytesAtom(pcp.PCPHostIP, ip),
		pcp.NewShortAtom(pcp.PCPHostPort, 7144),
		pcp.NewIntAtom(pcp.PCPHostNumListeners, 5),
		pcp.NewIntAtom(pcp.PCPHostNumRelays, 2),
		pcp.NewIntAtom(pcp.PCPHostUptime, 3600),
		pcp.NewByteAtom(pcp.PCPHostFlags1, flg1),
		pcp.NewIDAtom(pcp.PCPHostChanID, chanID),
	)

	bcst := pcp.NewParentAtom(pcp.PCPBcst,
		pcp.NewByteAtom(pcp.PCPBcstGroup, pcp.PCPBcstGroupTrackers),
		pcp.NewByteAtom(pcp.PCPBcstHops, 0),
		pcp.NewByteAtom(pcp.PCPBcstTTL, 7),
		pcp.NewIDAtom(pcp.PCPBcstFrom, sessID),
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

	sendBcst(t, conn, chanID, sessID, true)

	// Poll until the channel appears in the store.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := store.Snapshot()[chanID]; ok {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, ok := store.Snapshot()[chanID]; !ok {
		t.Fatal("channel not registered in store after bcst")
	}
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
	sendBcst(t, conn, chanID, sessID, true)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := store.Snapshot()[chanID]; ok {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, ok := store.Snapshot()[chanID]; !ok {
		t.Fatal("channel not registered after first bcst")
	}

	// Deregister by sending bcst with recv=false.
	sendBcst(t, conn, chanID, sessID, false)

	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := store.Snapshot()[chanID]; !ok {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, ok := store.Snapshot()[chanID]; ok {
		t.Fatal("channel still in store after recv=false bcst")
	}
}
