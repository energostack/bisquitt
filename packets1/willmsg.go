package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

type WillMsg struct {
	pkts.Header
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

func (p *WillMsg) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.Write(p.WillMsg)

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsg) Unpack(r io.Reader) (err error) {
	p.WillMsg = make([]byte, p.VarPartLength())
	_, err = io.ReadFull(r, p.WillMsg)
	return
}

func (p WillMsg) String() string {
	return fmt.Sprintf("WILLMSG(WillMsg=%#v)", string(p.WillMsg))
}
