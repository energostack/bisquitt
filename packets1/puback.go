package packets1

import (
	"fmt"
	"io"
)

const pubackVarPartLength uint16 = 5

type Puback struct {
	Header
	MessageIDProperty
	TopicID    uint16
	ReturnCode ReturnCode
}

func NewPuback(topicID uint16, returnCode ReturnCode) *Puback {
	return &Puback{
		Header:     *NewHeader(PUBACK, pubackVarPartLength),
		TopicID:    topicID,
		ReturnCode: returnCode,
	}
}

func (m *Puback) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.TopicID))
	buf.Write(encodeUint16(m.messageID))
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Puback) Unpack(r io.Reader) (err error) {
	if m.TopicID, err = readUint16(r); err != nil {
		return
	}

	if m.messageID, err = readUint16(r); err != nil {
		return
	}

	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m Puback) String() string {
	return fmt.Sprintf("PUBACK(TopicID=%d, ReturnCode=%d, MessageID=%d)", m.TopicID,
		m.ReturnCode, m.messageID)
}
