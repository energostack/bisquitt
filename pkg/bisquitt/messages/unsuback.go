package messages

import (
	"fmt"
	"io"
)

const unsubackVarPartLength uint16 = 2

type UnsubackMessage struct {
	Header
	MessageIDProperty
}

func NewUnsubackMessage() *UnsubackMessage {
	return &UnsubackMessage{
		Header: *NewHeader(UNSUBACK, unsubackVarPartLength),
	}
}

func (m *UnsubackMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *UnsubackMessage) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m UnsubackMessage) String() string {
	return fmt.Sprintf("UNSUBACK(MessageID=%d)", m.messageID)
}
