package packets1

import (
	"fmt"
	"io"
)

const registerHeaderLength uint16 = 4

type RegisterMessage struct {
	Header
	MessageIDProperty
	TopicID   uint16
	TopicName string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewRegisterMessage(topicID uint16, topicName string) *RegisterMessage {
	m := &RegisterMessage{
		Header:    *NewHeader(REGISTER, 0),
		TopicID:   topicID,
		TopicName: topicName,
	}
	m.computeLength()
	return m
}

func (m *RegisterMessage) computeLength() {
	topicLength := uint16(len(m.TopicName))
	m.Header.SetVarPartLength(registerHeaderLength + topicLength)
}

func (m *RegisterMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.Write(encodeUint16(m.TopicID))
	buf.Write(encodeUint16(m.messageID))
	buf.Write([]byte(m.TopicName))

	_, err := buf.WriteTo(w)
	return err
}

func (m *RegisterMessage) Unpack(r io.Reader) (err error) {
	if m.TopicID, err = readUint16(r); err != nil {
		return
	}

	if m.messageID, err = readUint16(r); err != nil {
		return
	}

	topic := make([]byte, m.VarPartLength()-registerHeaderLength)
	if _, err = io.ReadFull(r, topic); err != nil {
		return
	}
	m.TopicName = string(topic)
	return
}

func (m RegisterMessage) String() string {
	return fmt.Sprintf("REGISTER(TopicName=%#v, TopicID=%d, MessageID=%d)", string(m.TopicName),
		m.TopicID, m.messageID)
}
