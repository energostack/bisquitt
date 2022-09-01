package packets1

import (
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

func (p *Unsuback) Unpack(r io.Reader) (err error) {
	p.messageID, err = pkts.ReadUint16(r)
	return
}

func (p Unsuback) String() string {
	return fmt.Sprintf("UNSUBACK(MessageID=%d)", p.messageID)
}
