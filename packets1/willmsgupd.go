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
	m := &WillMsgUpdate{
		Header:  *NewHeader(WILLMSGUPD, 0),
		WillMsg: willMsg,
	}
	m.computeLength()
	return m
}

func (m *WillMsgUpdate) computeLength() {
	length := len(m.WillMsg)
	m.Header.SetVarPartLength(uint16(length))
}

func (m *WillMsgUpdate) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.Write(m.WillMsg)

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillMsgUpdate) Unpack(r io.Reader) (err error) {
	m.WillMsg = make([]byte, m.VarPartLength())
	_, err = io.ReadFull(r, m.WillMsg)
	return
}

func (m WillMsgUpdate) String() string {
	return fmt.Sprintf("WILLMSGUPDATE(WillMsg=%#v)", string(m.WillMsg))
}
