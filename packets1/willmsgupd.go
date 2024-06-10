package packets1

import (
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

type WillMsgUpd struct {
	pkts.Header
	// Fields
	WillMsg []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillMsgUpd(willMsg []byte) *WillMsgUpd {
	p := &WillMsgUpd{
		Header:  *pkts.NewHeader(pkts.WILLMSGUPD, 0),
		WillMsg: willMsg,
	}
	p.computeLength()
	return p
}

func (p *WillMsgUpd) computeLength() {
	length := len(p.WillMsg)
	p.Header.SetVarPartLength(uint16(length))
}

func (p *WillMsgUpd) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(p.WillMsg)

	return buf.Bytes(), nil
}

func (p *WillMsgUpd) Unpack(buf []byte) error {
	p.WillMsg = buf
	return nil
}

func (p WillMsgUpd) String() string {
	return fmt.Sprintf("WILLMSGUPD(WillMsg=%#v)", string(p.WillMsg))
}
