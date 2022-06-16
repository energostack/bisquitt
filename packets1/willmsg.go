package packets1

import (
	"fmt"
	"io"
)

type WillMsgMessage struct {
	Header
	WillMsg []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillMsgMessage(willMsg []byte) *WillMsgMessage {
	m := &WillMsgMessage{
		Header:  *NewHeader(WILLMSG, 0),
		WillMsg: willMsg,
	}
	m.computeLength()
	return m
}

func (m *WillMsgMessage) computeLength() {
	msgLength := len(m.WillMsg)
	m.Header.SetVarPartLength(uint16(msgLength))
}

func (m *WillMsgMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.Write(m.WillMsg)

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillMsgMessage) Unpack(r io.Reader) (err error) {
	m.WillMsg = make([]byte, m.VarPartLength())
	_, err = io.ReadFull(r, m.WillMsg)
	return
}

func (m WillMsgMessage) String() string {
	return fmt.Sprintf("WILLMSG(WillMsg=%#v)", string(m.WillMsg))
}
