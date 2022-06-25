package packets1

import (
	"fmt"
	"io"
)

type WillMsgUpdate struct {
	Header
	WillMsg []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillMsgUpdate(willMsg []byte) *WillMsgUpdate {
	p := &WillMsgUpdate{
		Header:  *NewHeader(WILLMSGUPD, 0),
		WillMsg: willMsg,
	}
	p.computeLength()
	return p
}

func (p *WillMsgUpdate) computeLength() {
	length := len(p.WillMsg)
	p.Header.SetVarPartLength(uint16(length))
}

func (p *WillMsgUpdate) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.pack()
	buf.Write(p.WillMsg)

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsgUpdate) Unpack(r io.Reader) (err error) {
	p.WillMsg = make([]byte, p.VarPartLength())
	_, err = io.ReadFull(r, p.WillMsg)
	return
}

func (p WillMsgUpdate) String() string {
	return fmt.Sprintf("WILLMSGUPDATE(WillMsg=%#v)", string(p.WillMsg))
}
