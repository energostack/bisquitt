package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pubcompVarPartLength uint16 = 2

type Pubcomp struct {
	pkts.Header
	MessageIDProperty
}

func NewPubcomp() *Pubcomp {
	return &Pubcomp{
		Header: *pkts.NewHeader(pkts.PUBCOMP, pubcompVarPartLength),
	}
}

func (p *Pubcomp) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubcomp) Unpack(r io.Reader) (err error) {
	p.messageID, err = pkts.ReadUint16(r)
	return
}

func (p Pubcomp) String() string {
	return fmt.Sprintf("PUBCOMP(MessageID=%d)", p.messageID)
}
