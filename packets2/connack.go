package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const connackHeaderLength uint16 = 6
const connackSessionPresentBit uint8 = 1

type Connack struct {
	pkts.Header
	PacketV2
	// Flags
	SessionPresent bool
	// Fields
	ReasonCode               ReasonCode
	SessionExpiryInterval    uint32
	AssignedClientIdentifier string
}

func NewConnack(reasonCode ReasonCode, sessExpiry uint32, assignedClientID string,
	sessPresent bool) *Connack {
	p := &Connack{
		Header:                   *pkts.NewHeader(pkts.CONNACK, 0),
		SessionPresent:           sessPresent,
		ReasonCode:               reasonCode,
		SessionExpiryInterval:    sessExpiry,
		AssignedClientIdentifier: assignedClientID,
	}
	p.computeLength()
	return p
}

func (p *Connack) computeLength() {
	assignedClientIDLength := uint16(len(p.AssignedClientIdentifier))
	p.Header.SetVarPartLength(connackHeaderLength + assignedClientIDLength)
}

func (p *Connack) decodeFlags(b byte) {
	p.SessionPresent = (b & connackSessionPresentBit) == connackSessionPresentBit
}

func (p *Connack) encodeFlags() byte {
	var b byte
	if p.SessionPresent {
		b |= connackSessionPresentBit
	}
	return b
}

// NOTE: The MQTT-SN v. 2.0 specification draft WD20 states that both
// SessionExpiryInterval and AssignedClientIdentifier are optional (chapter
// 2.1.5 CONNACK) but there is no means to signalize which one is or is not
// present. I consider this a bug in the specification. I'm fixing it
// temporarily by just considering SessionExpiryInterval mandatory.
func (p *Connack) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReasonCode))
	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint32(p.SessionExpiryInterval))
	if p.AssignedClientIdentifier != "" {
		_, _ = buf.Write([]byte(p.AssignedClientIdentifier))
	}

	return buf.Bytes(), nil
}

// NOTE: SessionExpiryInterval considered mandatory here, see NOTE for Pack().
func (p *Connack) Unpack(buf []byte) error {
	if len(buf) < int(connackHeaderLength) {
		return fmt.Errorf("bad CONNACK2 packet length: expected >=%d, got %d",
			connackHeaderLength, len(buf))
	}

	p.ReasonCode = ReasonCode(buf[0])
	p.decodeFlags(buf[1])
	p.SessionExpiryInterval = binary.BigEndian.Uint32(buf[2:6])
	p.AssignedClientIdentifier = string(buf[6:])

	return nil
}

func (p Connack) String() string {
	return fmt.Sprintf("CONNACK2(ReasonCode=%d,AssClientId=%q,SessExpiry=%d,SessPresent=%t)",
		p.ReasonCode, p.AssignedClientIdentifier, p.SessionExpiryInterval, p.SessionPresent)
}
