package packets2

import (
	"encoding/binary"
	"fmt"

	"github.com/energomonitor/bisquitt/packets"
	pkts "github.com/energomonitor/bisquitt/packets"
)

// TODO: Current specification draft (WD20) says that all fields are optional but
// it's unclear how missing fields should be signalized in the packet. Therefore
// I implement all the fields as mandatory for now.

const disconnectHeaderLength uint16 = 5

type Disconnect struct {
	pkts.Header
	PacketV2
	// Fields
	ReasonCode            ReasonCode
	SessionExpiryInterval uint32
	ReasonString          string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewDisconnect(reasonCode ReasonCode, sessExpiry uint32, reasonString string) *Disconnect {
	p := &Disconnect{
		Header:                *pkts.NewHeader(pkts.DISCONNECT, 0),
		ReasonCode:            reasonCode,
		SessionExpiryInterval: sessExpiry,
		ReasonString:          reasonString,
	}
	p.computeLength()
	return p
}

func (p *Disconnect) computeLength() {
	p.Header.SetVarPartLength(disconnectHeaderLength + uint16(len(p.ReasonString)))
}

func (p *Disconnect) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(uint8(p.ReasonCode))

	expiryEncoded := packets.EncodeUint32(p.SessionExpiryInterval)
	_, _ = buf.Write(expiryEncoded)
	_, _ = buf.Write([]byte(p.ReasonString))

	return buf.Bytes(), nil
}

func (p *Disconnect) Unpack(buf []byte) error {
	if len(buf) < int(disconnectHeaderLength) {
		return fmt.Errorf("bad DISCONNECT2 packet length: expected >=%d, got %d",
			disconnectHeaderLength, len(buf))
	}

	p.ReasonCode = ReasonCode(buf[0])
	p.SessionExpiryInterval = binary.BigEndian.Uint32(buf[1:5])
	p.ReasonString = string(buf[5:])

	return nil
}

func (p Disconnect) String() string {
	return fmt.Sprintf("DISCONNECT2(ReasonCode=%s, Reason=%q, SessExpiry=%d)",
		p.ReasonCode, p.ReasonString, p.SessionExpiryInterval)
}
