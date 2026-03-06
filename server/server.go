// Package server implements the PeerCast root (YP) server.
// It listens for incoming CIN (Control-In) connections, performs the PCP
// handshake, and maintains the channel registry via periodic tracker updates.
package server

import (
	"context"
	"crypto/rand"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	pcp "github.com/titagaki/peercast-pcp/pcp"

	"github.com/titagaki/peercast-0yp/channel"
)

// Protocol constants mirroring the reference implementation defaults.
const (
	agentString      = "PeerCastRoot/0.1 (Go)"
	serverVersion    = uint32(1218)
	serverVersionVP  = uint32(27)
	minClientVersion = uint32(1200)

	updateInterval    = 120 * time.Second
	hitTimeout        = 180 * time.Second // dead-hit removal threshold
	readTimeout       = updateInterval + 60*time.Second
	maxCINConnections = 100
)

var versionExPrefix = [2]byte{'Y', 'P'}

const versionExNumber = uint16(1)

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

// Server is the PeerCast root (YP) server. It accepts CIN connections,
// handles the PCP handshake, dispatches incoming bcst atoms to the channel
// store, and periodically broadcasts root settings to all connected clients.
type Server struct {
	sessionID pcp.GnuID
	store     *channel.Store

	mu       sync.Mutex
	sessions map[pcp.GnuID]*session
}

// New creates a Server with a random session ID.
func New(store *channel.Store) (*Server, error) {
	var id pcp.GnuID
	if _, err := rand.Read(id[:]); err != nil {
		return nil, err
	}
	return &Server{
		sessionID: id,
		store:     store,
		sessions:  make(map[pcp.GnuID]*session),
	}, nil
}

// SessionID returns the server's own session ID.
func (srv *Server) SessionID() pcp.GnuID { return srv.sessionID }

// Serve accepts connections from ln and serves them until ctx is cancelled.
// Returns nil when the context is done.
func (srv *Server) Serve(ctx context.Context, ln net.Listener) error {
	defer ln.Close()

	go func() {
		<-ctx.Done()
		ln.Close()             // unblock Accept
		srv.closeAllSessions() // terminate existing connections
	}()

	go srv.cleanupLoop(ctx)
	go srv.broadcastLoop(ctx)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				slog.Error("accept error", "err", err)
				continue
			}
		}
		go srv.handleConn(ctx, conn)
	}
}

// ListenAndServe starts a TCP listener on addr and serves connections until
// ctx is cancelled. Returns nil when the context is done.
func (srv *Server) ListenAndServe(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ctx, ln)
}

// closeAllSessions closes all active session connections.
// It collects connections under the lock, then closes them outside the lock
// to avoid deadlocking with the handleConn defer that also acquires the lock.
func (srv *Server) closeAllSessions() {
	srv.mu.Lock()
	conns := make([]net.Conn, 0, len(srv.sessions))
	for _, s := range srv.sessions {
		conns = append(conns, s.conn)
	}
	srv.mu.Unlock()

	for _, c := range conns {
		c.Close() // double-close is harmless
	}
}

// cleanupLoop removes dead hits from the channel store every 500 ms,
// replicating C++ ChanMgr::clearDeadHits called from idleProc.
func (srv *Server) cleanupLoop(ctx context.Context) {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			srv.store.RemoveDeadHits(hitTimeout)
		}
	}
}

// broadcastLoop sends a bcst > root > upd packet to every connected CIN
// session every updateInterval seconds, replicating C++
// ServMgr::broadcastRootSettings called from idleProc.
func (srv *Server) broadcastLoop(ctx context.Context) {
	t := time.NewTicker(updateInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			srv.broadcastRootSettings()
		}
	}
}

// ----------------------------------------------------------------------------
// Per-connection session
// ----------------------------------------------------------------------------

// session holds the state for a single connected CIN client.
type session struct {
	conn     net.Conn
	remoteIP net.IP
	clientID pcp.GnuID

	// sendCh is a buffered channel used to pass atoms to the send goroutine.
	// If the buffer is full the packet is dropped (client is too slow).
	sendCh chan *pcp.Atom
	done   chan struct{} // closed when the session ends
}

// sendLoop drains sendCh and writes each atom to the connection. It exits
// when done is closed or a write error occurs.
func (s *session) sendLoop() {
	for {
		select {
		case <-s.done:
			return
		case a := <-s.sendCh:
			if err := a.Write(s.conn); err != nil {
				return
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Connection handling
// ----------------------------------------------------------------------------

// handleConn manages the full lifecycle of a single CIN connection:
// PCP version exchange → HELO/OLEH handshake → validation → main read loop.
func (srv *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	remoteIP := remoteAddr.IP

	// ── Step 1: Read "pcp\n" version atom ────────────────────────────────
	vAtom, err := pcp.ReadAtom(conn)
	if err != nil {
		return
	}
	if vAtom.Tag != pcp.PCPConnect {
		slog.Debug("unexpected first atom", "tag", vAtom.Tag, "addr", remoteAddr)
		return
	}

	// ── Step 2: Read HELO ────────────────────────────────────────────────
	heloAtom, err := pcp.ReadAtom(conn)
	if err != nil {
		return
	}
	if heloAtom.Tag != pcp.PCPHelo {
		slog.Debug("expected helo", "got", heloAtom.Tag, "addr", remoteAddr)
		return
	}
	clientAgent, clientSID, clientPort, clientVer := parseHelo(heloAtom)
	slog.Info("HELO", "addr", remoteAddr, "agent", clientAgent, "ver", clientVer)

	// ── Step 3: Send OLEH (always, even if we will reject next) ──────────
	if err := writeOleh(conn, srv.sessionID, remoteIP, clientPort); err != nil {
		return
	}

	// ── Step 4: Send informational root atoms (§3.4, no upd yet) ─────────
	if err := writeRootAtoms(conn, false); err != nil {
		return
	}

	// ── Step 5: Validate ─────────────────────────────────────────────────
	sendQuit := func(code uint32) {
		_ = pcp.NewIntAtom(pcp.PCPQuit, code).Write(conn)
	}

	if clientVer < minClientVersion {
		sendQuit(pcp.PCPErrorQuit + pcp.PCPErrorBadAgent)
		return
	}
	if clientSID.IsEmpty() {
		sendQuit(pcp.PCPErrorQuit + pcp.PCPErrorNotIdentified)
		return
	}
	if clientSID == srv.sessionID {
		return // loopback — silently close
	}

	srv.mu.Lock()
	if len(srv.sessions) >= maxCINConnections {
		srv.mu.Unlock()
		sendQuit(pcp.PCPErrorQuit + pcp.PCPErrorUnavailable)
		return
	}
	if _, dup := srv.sessions[clientSID]; dup {
		srv.mu.Unlock()
		sendQuit(pcp.PCPErrorQuit + pcp.PCPErrorAlreadyConnected)
		return
	}
	sess := &session{
		conn:     conn,
		remoteIP: remoteIP,
		clientID: clientSID,
		sendCh:   make(chan *pcp.Atom, 16),
		done:     make(chan struct{}),
	}
	srv.sessions[clientSID] = sess
	srv.mu.Unlock()

	defer func() {
		close(sess.done)
		srv.mu.Lock()
		delete(srv.sessions, clientSID)
		srv.mu.Unlock()
		slog.Info("CIN disconnected", "addr", remoteAddr)
	}()

	slog.Info("CIN connected", "addr", remoteAddr, "agent", clientAgent, "ver", clientVer)

	// ── Step 6: Send ok(0) + root > upd to trigger first tracker update ──
	if err := pcp.NewIntAtom(pcp.PCPOK, 0).Write(conn); err != nil {
		return
	}
	rootUpd := pcp.NewParentAtom(pcp.PCPRoot, pcp.NewEmptyAtom(pcp.PCPRootUpdate))
	if err := rootUpd.Write(conn); err != nil {
		return
	}

	// ── Step 7: Start send goroutine, then enter read loop ───────────────
	go sess.sendLoop()

	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		atom, err := pcp.ReadAtom(conn)
		if err != nil {
			if ctx.Err() == nil {
				slog.Debug("read error", "addr", remoteAddr, "err", err)
			}
			return
		}
		switch atom.Tag {
		case pcp.PCPBcst:
			srv.processBcst(sess, atom)
		case pcp.PCPQuit:
			slog.Info("client quit", "addr", remoteAddr)
			return
		}
	}
}

// ----------------------------------------------------------------------------
// Outgoing packet builders
// ----------------------------------------------------------------------------

// writeOleh sends the OLEH handshake response.
// rip is the client's observed global IP; port is the port the client
// advertised in its HELO (0 if the client is firewalled).
func writeOleh(w io.Writer, serverSID pcp.GnuID, remoteIP net.IP, clientPort uint16) error {
	return pcp.NewParentAtom(pcp.PCPOleh,
		pcp.NewStringAtom(pcp.PCPHeloAgent, agentString),
		pcp.NewIDAtom(pcp.PCPHeloSessionID, serverSID),
		pcp.NewIntAtom(pcp.PCPHeloVersion, serverVersion),
		pcp.NewBytesAtom(pcp.PCPHeloRemoteIP, encodeIP(remoteIP)),
		pcp.NewShortAtom(pcp.PCPHeloPort, clientPort),
	).Write(w)
}

// writeRootAtoms sends the root information packet (§3.4).
// If withUpd is true, a root > upd child is appended to request an
// immediate tracker update from the client.
func writeRootAtoms(w io.Writer, withUpd bool) error {
	interval := uint32(updateInterval.Seconds())
	children := []*pcp.Atom{
		pcp.NewIntAtom(pcp.PCPRootUpdInt, interval),
		pcp.NewStringAtom(pcp.PCPRootURL, ""),
		pcp.NewIntAtom(pcp.PCPRootCheckVer, minClientVersion),
		pcp.NewIntAtom(pcp.PCPRootNext, interval),
		pcp.NewStringAtom(pcp.PCPMesgASCII, ""),
	}
	if withUpd {
		children = append(children, pcp.NewEmptyAtom(pcp.PCPRootUpdate))
	}
	return pcp.NewParentAtom(pcp.PCPRoot, children...).Write(w)
}

// broadcastRootSettings sends a bcst > root > upd packet to every live CIN
// session, replicating C++ ServMgr::broadcastRootSettings(true).
func (srv *Server) broadcastRootSettings() {
	interval := uint32(updateInterval.Seconds())

	rootAtom := pcp.NewParentAtom(pcp.PCPRoot,
		pcp.NewIntAtom(pcp.PCPRootUpdInt, interval),
		pcp.NewStringAtom(pcp.PCPRootURL, ""),
		pcp.NewIntAtom(pcp.PCPRootCheckVer, minClientVersion),
		pcp.NewIntAtom(pcp.PCPRootNext, interval),
		pcp.NewStringAtom(pcp.PCPMesgASCII, ""),
		pcp.NewEmptyAtom(pcp.PCPRootUpdate), // upd: request a fresh tracker update
	)

	bcst := pcp.NewParentAtom(pcp.PCPBcst,
		pcp.NewByteAtom(pcp.PCPBcstGroup, pcp.PCPBcstGroupTrackers),
		pcp.NewByteAtom(pcp.PCPBcstHops, 0),
		pcp.NewByteAtom(pcp.PCPBcstTTL, 7),
		pcp.NewIDAtom(pcp.PCPBcstFrom, srv.sessionID),
		pcp.NewIntAtom(pcp.PCPBcstVersion, serverVersion),
		pcp.NewIntAtom(pcp.PCPBcstVersionVP, serverVersionVP),
		pcp.NewBytesAtom(pcp.PCPBcstVersionExPrefix, versionExPrefix[:]),
		pcp.NewShortAtom(pcp.PCPBcstVersionExNumber, versionExNumber),
		rootAtom,
	)

	srv.mu.Lock()
	sessions := make([]*session, 0, len(srv.sessions))
	for _, s := range srv.sessions {
		sessions = append(sessions, s)
	}
	srv.mu.Unlock()

	for _, s := range sessions {
		select {
		case s.sendCh <- bcst:
		default:
			// Client is too slow to drain its send buffer; skip this cycle.
		}
	}
}

// ----------------------------------------------------------------------------
// Incoming bcst processing
// ----------------------------------------------------------------------------

// processBcst handles a bcst atom from a connected tracker client.
// It parses the chan and host sub-atoms and updates the channel store.
func (srv *Server) processBcst(sess *session, atom *pcp.Atom) {
	var routingChanID pcp.GnuID
	var chanInfo *channel.Info
	var hit *channel.Hit

	for _, child := range atom.Children() {
		switch child.Tag {
		case pcp.PCPBcstChanID:
			routingChanID, _ = child.GetID()
		case pcp.PCPChan:
			info := channel.ParseChanAtom(child, routingChanID)
			chanInfo = &info
		case pcp.PCPHost:
			h := channel.ParseHostAtom(child, routingChanID, sess.remoteIP)
			hit = &h
		}
	}

	if hit == nil {
		return
	}

	if hit.Recv && chanInfo != nil {
		srv.store.AddHit(*chanInfo, *hit)
	} else if !hit.Recv {
		// Client signalling end-of-broadcast (recv=false in flg1).
		cid := hit.ChanID
		if cid.IsEmpty() {
			cid = routingChanID
		}
		srv.store.DelHit(cid, hit.SessionID)
	}
}

// ----------------------------------------------------------------------------
// Atom parsers
// ----------------------------------------------------------------------------

// parseHelo extracts the four key fields from a helo container atom.
func parseHelo(a *pcp.Atom) (agent string, sid pcp.GnuID, port uint16, ver uint32) {
	for _, child := range a.Children() {
		switch child.Tag {
		case pcp.PCPHeloAgent:
			agent = child.GetString()
		case pcp.PCPHeloSessionID:
			sid, _ = child.GetID()
		case pcp.PCPHeloPort:
			port, _ = child.GetShort()
		case pcp.PCPHeloVersion:
			ver, _ = child.GetInt()
		}
	}
	return
}

// ----------------------------------------------------------------------------
// IP address helpers
// ----------------------------------------------------------------------------

// encodeIP converts a net.IP to PCP wire-format bytes.
// IPv4 → 4 bytes as-is.
// IPv6 → 16 bytes in reversed byte order (per PCP spec §6.5).
func encodeIP(ip net.IP) []byte {
	if ip4 := ip.To4(); ip4 != nil {
		b := make([]byte, 4)
		copy(b, ip4)
		return b
	}
	b := make([]byte, 16)
	for i, v := range ip.To16() {
		b[15-i] = v
	}
	return b
}
