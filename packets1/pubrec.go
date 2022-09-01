package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pubrecVarPartLength uint16 = 2

type Pubrec struct {
	pkts.Header
	MessageIDProperty
}

func NewPubrec() *Pubrec {
	return &Pubrec{
		Header: *pkts.NewHeader(pkts.PUBREC, pubrecVarPartLength),
	}
}

func (p *Pubrec) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubrec) Unpack(r io.Reader) (err error) {
	p.messageID, err = pkts.ReadUint16(r)
	return
}

func (p Pubrec) String() string {
	return fmt.Sprintf("PUBREC(MessageID=%d)", p.messageID)
}
