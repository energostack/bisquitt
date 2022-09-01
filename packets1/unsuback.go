package packets1

import (
	"encoding/binary"
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const unsubackVarPartLength uint16 = 2

type Unsuback struct {
	pkts.Header
	MessageIDProperty
}

func NewUnsuback() *Unsuback {
	return &Unsuback{
		Header: *pkts.NewHeader(pkts.UNSUBACK, unsubackVarPartLength),
	}
}

func (p *Unsuback) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
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
