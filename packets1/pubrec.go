package packets1

import (
	"fmt"
	"io"
)

const pubrecVarPartLength uint16 = 2

type Pubrec struct {
	Header
	MessageIDProperty
}

func NewPubrec() *Pubrec {
	return &Pubrec{
		Header: *NewHeader(PUBREC, pubrecVarPartLength),
	}
}

func (p *Pubrec) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.Write(encodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubrec) Unpack(r io.Reader) (err error) {
	p.messageID, err = readUint16(r)
	return
}

func (p Pubrec) String() string {
	return fmt.Sprintf("PUBREC(MessageID=%d)", p.messageID)
}
