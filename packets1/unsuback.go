package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const unsubackVarPartLength uint16 = 2

type Unsuback struct {
	pkts.Header
	// Fields
	MessageIDProperty
}

func NewUnsuback() *Unsuback {
	return &Unsuback{
		Header: *pkts.NewHeader(pkts.UNSUBACK, unsubackVarPartLength),
	}
}

func (p *Unsuback) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))

	return buf.Bytes(), nil
}

func (p *Unsuback) Unpack(buf []byte) error {
	if len(buf) != int(unsubackVarPartLength) {
		return fmt.Errorf("bad UNSUBACK packet length: expected %d, got %d",
			unsubackVarPartLength, len(buf))
	}

	p.messageID = binary.BigEndian.Uint16(buf)

	return nil
}

func (p Unsuback) String() string {
	return fmt.Sprintf("UNSUBACK(MessageID=%d)", p.messageID)
}
