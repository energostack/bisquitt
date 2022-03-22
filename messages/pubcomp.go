package messages

import (
	"fmt"
	"io"
)

const pubcompVarPartLength uint16 = 2

type PubcompMessage struct {
	Header
	MessageIDProperty
}

func NewPubcompMessage() *PubcompMessage {
	return &PubcompMessage{
		Header: *NewHeader(PUBCOMP, pubcompVarPartLength),
	}
}

func (m *PubcompMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (m *PubcompMessage) Unpack(r io.Reader) (err error) {
	m.messageID, err = readUint16(r)
	return
}

func (m PubcompMessage) String() string {
	return fmt.Sprintf("PUBCOMP(MessageID=%d)", m.messageID)
}
