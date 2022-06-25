package packets1

import (
	"fmt"
	"io"
)

const pubrelVarPartLength uint16 = 2

type Pubrel struct {
	Header
	MessageIDProperty
}

func NewPubrel() *Pubrel {
	return &Pubrel{
		Header: *NewHeader(PUBREL, pubrelVarPartLength),
	}
}

func (p *Pubrel) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.Write(encodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubrel) Unpack(r io.Reader) (err error) {
	p.messageID, err = readUint16(r)
	return
}

func (p Pubrel) String() string {
	return fmt.Sprintf("PUBREL(MessageID=%d)", p.messageID)
}
