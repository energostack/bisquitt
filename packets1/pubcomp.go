package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

const pubcompVarPartLength uint16 = 2

type Pubcomp struct {
	pkts.Header
	// Fields
	MessageIDProperty
}

func NewPubcomp() *Pubcomp {
	return &Pubcomp{
		Header: *pkts.NewHeader(pkts.PUBCOMP, pubcompVarPartLength),
	}
}

func (p *Pubcomp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))

	return buf.Bytes(), nil
}

func (p *Pubcomp) Unpack(buf []byte) error {
	if len(buf) != int(pubcompVarPartLength) {
		return fmt.Errorf("bad PUBCOMP packet length: expected %d, got %d",
			pubcompVarPartLength, len(buf))
	}

	p.messageID = binary.BigEndian.Uint16(buf)

	return nil
}

func (p Pubcomp) String() string {
	return fmt.Sprintf("PUBCOMP(MessageID=%d)", p.messageID)
}
