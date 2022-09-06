package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pubrelVarPartLength uint16 = 2

type Pubrel struct {
	pkts.Header
	MessageIDProperty
}

func NewPubrel() *Pubrel {
	return &Pubrel{
		Header: *pkts.NewHeader(pkts.PUBREL, pubrelVarPartLength),
	}
}

func (p *Pubrel) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))

	return buf.Bytes(), nil
}

func (p *Pubrel) Unpack(buf []byte) error {
	if len(buf) != int(pubrelVarPartLength) {
		return fmt.Errorf("bad PUBREL packet length: expected %d, got %d",
			pubrelVarPartLength, len(buf))
	}

	p.messageID = binary.BigEndian.Uint16(buf)

	return nil
}

func (p Pubrel) String() string {
	return fmt.Sprintf("PUBREL(MessageID=%d)", p.messageID)
}
