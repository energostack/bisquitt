package packets1

import (
	"fmt"
	"io"
)

const willMsgRespVarPartLength uint16 = 1

type WillMsgRespMessage struct {
	Header
	ReturnCode ReturnCode
}

func NewWillMsgRespMessage(returnCode ReturnCode) *WillMsgRespMessage {
	return &WillMsgRespMessage{
		Header:     *NewHeader(WILLMSGRESP, willMsgRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (m *WillMsgRespMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillMsgRespMessage) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m WillMsgRespMessage) String() string {
	return fmt.Sprintf("WILLMSGRESP(ReturnCode=%d)", m.ReturnCode)
}
