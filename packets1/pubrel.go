package packets1

import (
	"fmt"
	"io"

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

func (p *Pubrel) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubrel) Unpack(r io.Reader) (err error) {
	p.messageID, err = pkts.ReadUint16(r)
	return
}

func (p Pubrel) String() string {
	return fmt.Sprintf("PUBREL(MessageID=%d)", p.messageID)
}
