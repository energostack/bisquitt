package packets1

import (
	"fmt"
	"io"
)

const pubrecVarPartLength uint16 = 2

type PubrecMessage struct {
	Header
	MessageIDProperty
}

func NewPubrecMessage() *PubrecMessage {
	return &PubrecMessage{
		Header: *NewHeader(PUBREC, pubrecVarPartLength),
	}
}

func (m *PubrecMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *PubrecMessage) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m PubrecMessage) String() string {
	return fmt.Sprintf("PUBREC(MessageID=%d)", m.messageID)
}
