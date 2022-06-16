package packets1

import (
	"fmt"
	"io"
)

const connackVarPartLength uint16 = 1

type ConnackMessage struct {
	Header
	ReturnCode ReturnCode
}

func NewConnackMessage(returnCode ReturnCode) *ConnackMessage {
	return &ConnackMessage{
		Header:     *NewHeader(CONNACK, connackVarPartLength),
		ReturnCode: returnCode,
	}
}

func (m *ConnackMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *ConnackMessage) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m ConnackMessage) String() string {
	return fmt.Sprintf("CONNACK(ReturnCode=%d)", m.ReturnCode)
}
