package packets1

import (
	"fmt"
	"io"
)

const connackVarPartLength uint16 = 1

type Connack struct {
	Header
	ReturnCode ReturnCode
}

func NewConnack(returnCode ReturnCode) *Connack {
	return &Connack{
		Header:     *NewHeader(CONNACK, connackVarPartLength),
		ReturnCode: returnCode,
	}
}

func (m *Connack) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Connack) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m Connack) String() string {
	return fmt.Sprintf("CONNACK(ReturnCode=%d)", m.ReturnCode)
}
