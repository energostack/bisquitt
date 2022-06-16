package packets1

import (
	"fmt"
	"io"
)

const subackVarPartLength uint16 = 6

type Suback struct {
	Header
	MessageIDProperty
	QOS        uint8
	ReturnCode ReturnCode
	TopicID    uint16
}

func NewSuback(topicID uint16, qos uint8, returnCode ReturnCode) *Suback {
	return &Suback{
		Header:     *NewHeader(SUBACK, subackVarPartLength),
		QOS:        qos,
		ReturnCode: returnCode,
		TopicID:    topicID,
	}
}

func (m *Suback) encodeFlags() byte {
	var b byte
	b |= (m.QOS << 5) & flagsQOSBits
	return b
}

func (m *Suback) decodeFlags(b byte) {
	m.QOS = (b & flagsQOSBits) >> 5
}

func (m *Suback) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(m.encodeFlags())
	buf.Write(encodeUint16(m.TopicID))
	buf.Write(encodeUint16(m.messageID))
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Suback) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = readByte(r); err != nil {
		return
	}
	m.decodeFlags(flagsByte)

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

func (m Suback) String() string {
	return fmt.Sprintf("SUBACK(TopicID=%d, MessageID=%d, ReturnCode=%d, QOS=%d)", m.TopicID,
		m.messageID, m.ReturnCode, m.QOS)
}
