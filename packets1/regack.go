package packets1

import (
	"fmt"
	"io"
)

const regackVarPartLength uint16 = 5

type Regack struct {
	Header
	MessageIDProperty
	TopicID    uint16
	ReturnCode ReturnCode
}

func NewRegack(topicID uint16, returnCode ReturnCode) *Regack {
	return &Regack{
		Header:     *NewHeader(REGACK, regackVarPartLength),
		TopicID:    topicID,
		ReturnCode: returnCode,
	}
}

func (m *Regack) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.Write(encodeUint16(m.TopicID))
	buf.Write(encodeUint16(m.messageID))
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Regack) Unpack(r io.Reader) (err error) {
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

func (m Regack) String() string {
	return fmt.Sprintf("REGACK(TopicID=%d, ReturnCode=%d, MessageID=%d)", m.TopicID,
		m.ReturnCode, m.messageID)
}
