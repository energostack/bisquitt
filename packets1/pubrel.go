package packets1

import (
	"fmt"
	"io"
)

const pubrelVarPartLength uint16 = 2

type PubrelMessage struct {
	Header
	MessageIDProperty
}

func NewPubrelMessage() *PubrelMessage {
	return &PubrelMessage{
		Header: *NewHeader(PUBREL, pubrelVarPartLength),
	}
}

func (m *PubrelMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *PubrelMessage) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m PubrelMessage) String() string {
	return fmt.Sprintf("PUBREL(MessageID=%d)", m.messageID)
}
