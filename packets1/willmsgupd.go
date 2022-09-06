package packets1

import (
	"fmt"

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

func (p *WillMsgUpdate) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(p.WillMsg)

	return buf.Bytes(), nil
}

func (p *WillMsgUpdate) Unpack(buf []byte) error {
	p.WillMsg = buf
	return nil
}

func (p WillMsgUpdate) String() string {
	return fmt.Sprintf("WILLMSGUPDATE(WillMsg=%#v)", string(p.WillMsg))
}
