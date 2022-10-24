package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const connectHeaderLength uint16 = 10
const connectCleanStartBit uint8 = 1 << 0
const connectWillBit uint8 = 1 << 1
const connectAuthBit uint8 = 1 << 2

type Connect struct {
	pkts.Header
	PacketV2
	// Flags
	Authentication bool
	Will           bool
	CleanStart     bool
	// Fields
	ProtocolVersion       uint8
	KeepAlive             uint16
	SessionExpiryInterval uint32
	MaxPacketSize         uint16
	ClientIdentifier      string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewConnect(keepAlive uint16, sessExpiry uint32, maxPktSize uint16, clientID string,
	auth bool, will bool, cleanStart bool) *Connect {
	p := &Connect{
		Header:                *pkts.NewHeader(pkts.CONNECT, 0),
		Authentication:        auth,
		Will:                  will,
		CleanStart:            cleanStart,
		ProtocolVersion:       0x02,
		KeepAlive:             keepAlive,
		SessionExpiryInterval: sessExpiry,
		MaxPacketSize:         maxPktSize,
		ClientIdentifier:      clientID,
	}
	p.computeLength()
	return p
}

func (p *Connect) computeLength() {
	clientIDLength := uint16(len(p.ClientIdentifier))
	p.Header.SetVarPartLength(connectHeaderLength + clientIDLength)
}

func (p *Connect) decodeFlags(b byte) {
	p.CleanStart = (b & connectCleanStartBit) == connectCleanStartBit
	p.Will = (b & connectWillBit) == connectWillBit
	p.Authentication = (b & connectAuthBit) == connectAuthBit
}

func (p *Connect) encodeFlags() byte {
	var b byte
	if p.CleanStart {
		b |= connectCleanStartBit
	}
	if p.Will {
		b |= connectWillBit
	}
	if p.Authentication {
		b |= connectAuthBit
	}
	return b
}

func (p *Connect) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_ = buf.WriteByte(p.ProtocolVersion)
	_, _ = buf.Write(pkts.EncodeUint16(p.KeepAlive))
	_, _ = buf.Write(pkts.EncodeUint32(p.SessionExpiryInterval))
	_, _ = buf.Write(pkts.EncodeUint16(p.MaxPacketSize))
	_, _ = buf.Write([]byte(p.ClientIdentifier))

	return buf.Bytes(), nil
}

func (p *Connect) Unpack(buf []byte) error {
	if len(buf) <= int(connectHeaderLength) {
		return fmt.Errorf("bad CONNECT2 packet length: expected >%d, got %d",
			connectHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.ProtocolVersion = buf[1]
	if p.ProtocolVersion != 2 {
		return fmt.Errorf("bad CONNECT2 ProtocolVersion: expected 2, got %d", buf[1])
	}
	p.KeepAlive = binary.BigEndian.Uint16(buf[2:4])
	p.SessionExpiryInterval = binary.BigEndian.Uint32(buf[4:8])
	p.MaxPacketSize = binary.BigEndian.Uint16(buf[8:10])
	p.ClientIdentifier = string(buf[10:])

	return nil
}

func (p Connect) String() string {
	return fmt.Sprintf(
		"CONNECT2(ClientID=%q, CleanStart=%t, Will=%t, Auth=%t, SessExpiry=%d, KeepAlive=%d, MaxPktSize=%d)",
		p.ClientIdentifier, p.CleanStart, p.Will, p.Authentication,
		p.SessionExpiryInterval, p.KeepAlive, p.MaxPacketSize,
	)
}
