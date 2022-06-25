package packets1

import (
	"fmt"
	"io"
)

const pubcompVarPartLength uint16 = 2

type Pubcomp struct {
	Header
	MessageIDProperty
}

func NewPubcomp() *Pubcomp {
	return &Pubcomp{
		Header: *NewHeader(PUBCOMP, pubcompVarPartLength),
	}
}

func (p *Pubcomp) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.Write(encodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubcomp) Unpack(r io.Reader) (err error) {
	p.messageID, err = readUint16(r)
	return
}

func (p Pubcomp) String() string {
	return fmt.Sprintf("PUBCOMP(MessageID=%d)", p.messageID)
}
