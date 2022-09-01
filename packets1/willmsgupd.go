package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

type WillMsgUpdate struct {
	pkts.Header
	WillMsg []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillMsgUpdate(willMsg []byte) *WillMsgUpdate {
	p := &WillMsgUpdate{
		Header:  *pkts.NewHeader(pkts.WILLMSGUPD, 0),
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

	buf := p.Header.Pack()
	buf.Write(p.WillMsg)

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsgUpdate) Unpack(buf []byte) error {
	p.WillMsg = buf
	return nil
}

func (p WillMsgUpdate) String() string {
	return fmt.Sprintf("WILLMSGUPDATE(WillMsg=%#v)", string(p.WillMsg))
}
