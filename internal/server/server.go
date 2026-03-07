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

	"github.com/titagaki/peercast-0yp/internal/channel"
)

// Protocol-level constants that are part of the PCP implementation and
// are not operator-configurable.
const (
	agentString     = "PeerCastRoot/0.1 (Go)"
	serverVersion   = uint32(1218)
	serverVersionVP = uint32(27)
)

var versionExPrefix = [2]byte{'Y', 'P'}

const versionExNumber = uint16(1)

// Config holds operator-configurable server parameters.
type Config struct {
	MaxConnections   int           // maximum simultaneous CIN connections (default 100)
	UpdateInterval   time.Duration // how often to request tracker updates (default 120s)
	HitTimeout       time.Duration // dead-hit removal threshold (default 180s)
	MinClientVersion uint32        // minimum accepted client version (default 1200)
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxConnections:   100,
		UpdateInterval:   120 * time.Second,
		HitTimeout:       180 * time.Second,
		MinClientVersion: 1200,
	}
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

// Server is the PeerCast root (YP) server. It accepts CIN connections,
// handles the PCP handshake, dispatches incoming bcst atoms to the channel
// store, and periodically broadcasts root settings to all connected clients.
type Server struct {
	cfg       Config
	sessionID pcp.GnuID
	store     *channel.Store

	mu       sync.Mutex
	sessions map[pcp.GnuID]*session
}

// New creates a Server with a random session ID.
func New(store *channel.Store, cfg Config) (*Server, error) {
	var id pcp.GnuID
	if _, err := rand.Read(id[:]); err != nil {
		return nil, err
	}
	return &Server{
		cfg:       cfg,
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
func (srv *Server) closeAllSessions() {
	srv.mu.Lock()
	conns := make([]net.Conn, 0, len(srv.sessions))
	for _, s := range srv.sessions {
		conns = append(conns, s.conn)
	}
	srv.mu.Unlock()

	for _, c := range conns {
		c.Close()
	}
}

// cleanupLoop removes dead hits from the channel store every 500 ms.
func (srv *Server) cleanupLoop(ctx context.Context) {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			srv.store.RemoveDeadHits(srv.cfg.HitTimeout)
		}
	}
}

// broadcastLoop sends a bcst > root > upd packet to every connected CIN
// session every UpdateInterval seconds.
func (srv *Server) broadcastLoop(ctx context.Context) {
	t := time.NewTicker(srv.cfg.UpdateInterval)
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

type session struct {
	conn     net.Conn
	remoteIP net.IP
	clientID pcp.GnuID
	sendCh   chan *pcp.Atom
	done     chan struct{}
}

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

func (srv *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	remoteIP := remoteAddr.IP

	readTimeout := srv.cfg.UpdateInterval + 60*time.Second

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

	// ── Step 3: Send OLEH ────────────────────────────────────────────────
	if err := writeOleh(conn, srv.sessionID, remoteIP, clientPort); err != nil {
		return
	}

	// ── Step 4: Send informational root atoms (no upd yet) ───────────────
	if err := writeRootAtoms(conn, srv.cfg, false); err != nil {
		return
	}

	// ── Step 5: Validate ─────────────────────────────────────────────────
	sendQuit := func(code uint32) {
		_ = pcp.NewIntAtom(pcp.PCPQuit, code).Write(conn)
	}

	if clientVer < srv.cfg.MinClientVersion {
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
	if len(srv.sessions) >= srv.cfg.MaxConnections {
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

	// ── Step 6: Send ok(0) + root > upd ──────────────────────────────────
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

func writeOleh(w io.Writer, serverSID pcp.GnuID, remoteIP net.IP, clientPort uint16) error {
	return pcp.NewParentAtom(pcp.PCPOleh,
		pcp.NewStringAtom(pcp.PCPHeloAgent, agentString),
		pcp.NewIDAtom(pcp.PCPHeloSessionID, serverSID),
		pcp.NewIntAtom(pcp.PCPHeloVersion, serverVersion),
		pcp.NewBytesAtom(pcp.PCPHeloRemoteIP, encodeIP(remoteIP)),
		pcp.NewShortAtom(pcp.PCPHeloPort, clientPort),
	).Write(w)
}

func writeRootAtoms(w io.Writer, cfg Config, withUpd bool) error {
	interval := uint32(cfg.UpdateInterval.Seconds())
	children := []*pcp.Atom{
		pcp.NewIntAtom(pcp.PCPRootUpdInt, interval),
		pcp.NewStringAtom(pcp.PCPRootURL, ""),
		pcp.NewIntAtom(pcp.PCPRootCheckVer, cfg.MinClientVersion),
		pcp.NewIntAtom(pcp.PCPRootNext, interval),
		pcp.NewStringAtom(pcp.PCPMesgASCII, ""),
	}
	if withUpd {
		children = append(children, pcp.NewEmptyAtom(pcp.PCPRootUpdate))
	}
	return pcp.NewParentAtom(pcp.PCPRoot, children...).Write(w)
}

func (srv *Server) broadcastRootSettings() {
	interval := uint32(srv.cfg.UpdateInterval.Seconds())

	rootAtom := pcp.NewParentAtom(pcp.PCPRoot,
		pcp.NewIntAtom(pcp.PCPRootUpdInt, interval),
		pcp.NewStringAtom(pcp.PCPRootURL, ""),
		pcp.NewIntAtom(pcp.PCPRootCheckVer, srv.cfg.MinClientVersion),
		pcp.NewIntAtom(pcp.PCPRootNext, interval),
		pcp.NewStringAtom(pcp.PCPMesgASCII, ""),
		pcp.NewEmptyAtom(pcp.PCPRootUpdate),
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
		}
	}
}

// ----------------------------------------------------------------------------
// Incoming bcst processing
// ----------------------------------------------------------------------------

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
