package packets1

import (
	"io"
)

const willMsgReqVarPartLength uint16 = 0

type WillMsgReqMessage struct {
	Header
}

func NewWillMsgReqMessage() *WillMsgReqMessage {
	return &WillMsgReqMessage{
		Header: *NewHeader(WILLMSGREQ, willMsgReqVarPartLength),
	}
}

func (m *WillMsgReqMessage) Write(w io.Writer) error {
	buf := m.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillMsgReqMessage) Unpack(r io.Reader) error {
	return nil
}

func (m WillMsgReqMessage) String() string {
	return "WILLMSGREQ"
}
