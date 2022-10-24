package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pingrespHeaderLength uint16 = 0

type Pingresp struct {
	pkts.Header
	PacketV2
	// Fields
	MessagesRemaining uint8
	// Auxiliary
	MessagesRemainingPresent bool
}

func NewPingresp(msgsRemaining uint8, msgsRemainingPresent bool) *Pingresp {
	p := &Pingresp{
		Header:                   *pkts.NewHeader(pkts.PINGRESP, 0),
		MessagesRemaining:        msgsRemaining,
		MessagesRemainingPresent: msgsRemainingPresent,
	}
	p.computeLength()
	return p
}

func (p *Pingresp) computeLength() {
	if p.MessagesRemainingPresent {
		p.Header.SetVarPartLength(pingrespHeaderLength + 1)
	} else {
		p.Header.SetVarPartLength(pingrespHeaderLength)
	}
}

func (p *Pingresp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	if p.MessagesRemainingPresent {
		_ = buf.WriteByte(p.MessagesRemaining)
	}

	return buf.Bytes(), nil
}

func (p *Pingresp) Unpack(buf []byte) error {
	switch len(buf) {
	case 0:
		p.MessagesRemainingPresent = false
		return nil
	case 1:
		p.MessagesRemainingPresent = true
		p.MessagesRemaining = buf[0]
		return nil
	default:
		return fmt.Errorf("bad PINGRESP2 packet length: expected <=1, got %d",
			len(buf))
	}
}

func (p Pingresp) String() string {
	if p.MessagesRemainingPresent {
		return fmt.Sprintf("PINGRESP2(MsgsRemaining=%d)", p.MessagesRemaining)
	}
	return "PINGRESP2"
}
