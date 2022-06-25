package packets1

import (
	"fmt"
	"io"
)

const unsubackVarPartLength uint16 = 2

type Unsuback struct {
	Header
	MessageIDProperty
}

func NewUnsuback() *Unsuback {
	return &Unsuback{
		Header: *NewHeader(UNSUBACK, unsubackVarPartLength),
	}
}

func (p *Unsuback) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.Write(encodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Unsuback) Unpack(r io.Reader) (err error) {
	p.messageID, err = readUint16(r)
	return
}

func (p Unsuback) String() string {
	return fmt.Sprintf("UNSUBACK(MessageID=%d)", p.messageID)
}
