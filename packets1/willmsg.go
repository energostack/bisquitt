package packets1

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

type WillMsg struct {
	pkts.Header
	// Fields
	WillMsg []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillMsg(willMsg []byte) *WillMsg {
	p := &WillMsg{
		Header:  *pkts.NewHeader(pkts.WILLMSG, 0),
		WillMsg: willMsg,
	}
	p.computeLength()
	return p
}

func (p *WillMsg) computeLength() {
	msgLength := len(p.WillMsg)
	p.Header.SetVarPartLength(uint16(msgLength))
}

func (p *WillMsg) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(p.WillMsg)

	return buf.Bytes(), nil
}

func (p *WillMsg) Unpack(buf []byte) error {
	p.WillMsg = buf
	return nil
}

func (p WillMsg) String() string {
	return fmt.Sprintf("WILLMSG(WillMsg=%#v)", string(p.WillMsg))
}
