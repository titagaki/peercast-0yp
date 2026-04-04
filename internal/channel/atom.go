package channel

import (
	"net"

	pcp "github.com/titagaki/peercast-pcp/pcp"
)

// ParseChanAtom extracts an Info from a chan container atom.
// defaultID is used as the channel ID if no id sub-atom is present.
func ParseChanAtom(a *pcp.Atom, defaultID pcp.GnuID) Info {
	info := Info{ID: defaultID}
	for _, child := range a.Children() {
		switch child.Tag {
		case pcp.PCPChanID:
			if id, err := child.GetID(); err == nil {
				info.ID = id
			}
		case pcp.PCPChanBCID:
			info.BroadcastID, _ = child.GetID()
		case pcp.PCPChanInfo:
			parseInfoAtom(child, &info)
		case pcp.PCPChanTrack:
			parseTrackAtom(child, &info.Track)
		}
	}
	return info
}

// ParseHostAtom extracts a Hit from a host container atom.
// fallbackIP is used as the global IP when no ip sub-atom is present.
func ParseHostAtom(a *pcp.Atom, chanID pcp.GnuID, fallbackIP net.IP) Hit {
	hit := Hit{ChanID: chanID}
	var ips [2]net.IP
	var ports [2]uint16
	ipIdx := 0
	portIdx := 0

	for _, child := range a.Children() {
		switch child.Tag {
		case pcp.PCPHostID:
			hit.SessionID, _ = child.GetID()
		case pcp.PCPHostIP:
			if ipIdx < 2 {
				ips[ipIdx] = decodeIP(child.Data())
				ipIdx++
			}
		case pcp.PCPHostPort:
			if portIdx < 2 {
				ports[portIdx], _ = child.GetShort()
				portIdx++
			}
		case pcp.PCPHostNumListeners:
			hit.NumListeners, _ = child.GetInt()
		case pcp.PCPHostNumRelays:
			hit.NumRelays, _ = child.GetInt()
		case pcp.PCPHostUptime:
			hit.UpTime, _ = child.GetInt()
		case pcp.PCPHostVersion:
			hit.Version, _ = child.GetInt()
		case pcp.PCPHostVersionVP:
			hit.VersionVP, _ = child.GetInt()
		case pcp.PCPHostVersionExPrefix:
			if d := child.Data(); len(d) >= 2 {
				copy(hit.VersionExPfx[:], d[:2])
			}
		case pcp.PCPHostVersionExNumber:
			hit.VersionExNum, _ = child.GetShort()
		case pcp.PCPHostFlags1:
			b, _ := child.GetByte()
			hit.Tracker = b&pcp.PCPHostFlags1Tracker != 0
			hit.Relay = b&pcp.PCPHostFlags1Relay != 0
			hit.Direct = b&pcp.PCPHostFlags1Direct != 0
			hit.Firewalled = b&pcp.PCPHostFlags1Push != 0
			hit.Recv = b&pcp.PCPHostFlags1Recv != 0
			hit.CIN = b&pcp.PCPHostFlags1CIN != 0
		case pcp.PCPHostOldPos:
			hit.OldPos, _ = child.GetInt()
		case pcp.PCPHostNewPos:
			hit.NewPos, _ = child.GetInt()
		case pcp.PCPHostChanID:
			if id, err := child.GetID(); err == nil {
				hit.ChanID = id
			}
		}
	}

	if ips[0] == nil {
		ips[0] = fallbackIP
	}
	hit.GlobalAddr = net.TCPAddr{IP: ips[0], Port: int(ports[0])}
	hit.LocalAddr = net.TCPAddr{IP: ips[1], Port: int(ports[1])}
	return hit
}

func parseInfoAtom(a *pcp.Atom, info *Info) {
	for _, child := range a.Children() {
		switch child.Tag {
		case pcp.PCPChanInfoName:
			info.Name = child.GetString()
		case pcp.PCPChanInfoBitrate:
			info.Bitrate, _ = child.GetInt()
		case pcp.PCPChanInfoGenre:
			info.Genre = child.GetString()
		case pcp.PCPChanInfoURL:
			info.URL = child.GetString()
		case pcp.PCPChanInfoDesc:
			info.Desc = child.GetString()
		case pcp.PCPChanInfoComment:
			info.Comment = child.GetString()
		case pcp.PCPChanInfoType:
			info.ContentType = child.GetString()
		case pcp.PCPChanInfoStreamType:
			info.MIMEType = child.GetString()
		case pcp.PCPChanInfoStreamExt:
			info.StreamExt = child.GetString()
		}
	}
}

func parseTrackAtom(a *pcp.Atom, t *Track) {
	for _, child := range a.Children() {
		switch child.Tag {
		case pcp.PCPChanTrackTitle:
			t.Title = child.GetString()
		case pcp.PCPChanTrackCreator:
			t.Artist = child.GetString()
		case pcp.PCPChanTrackURL:
			t.Contact = child.GetString()
		case pcp.PCPChanTrackAlbum:
			t.Album = child.GetString()
		}
	}
}

// decodeIP converts PCP wire-format bytes to a net.IP.
// 4 bytes → IPv4 (via pcp.DecodeIPv4). 16 bytes → IPv6 (reversed on the wire, per PCP spec §6.5).
func decodeIP(data []byte) net.IP {
	switch len(data) {
	case 4:
		return pcp.DecodeIPv4(data)
	case 16:
		ip := make(net.IP, 16)
		for i, b := range data {
			ip[15-i] = b
		}
		return ip
	}
	return nil
}
