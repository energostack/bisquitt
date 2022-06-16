package packets1

import (
	"fmt"
	"io"
)

const unsubscribeHeaderLength uint16 = 3

type UnsubscribeMessage struct {
	Header
	MessageIDProperty
	TopicIDType uint8
	TopicID     uint16
	TopicName   []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewUnsubscribeMessage(topicID uint16, topicIDType uint8, topicName []byte) *UnsubscribeMessage {
	m := &UnsubscribeMessage{
		Header:      *NewHeader(UNSUBSCRIBE, 0),
		TopicIDType: topicIDType,
		TopicID:     topicID,
		TopicName:   topicName,
	}
	m.computeLength()
	return m
}

func (m *UnsubscribeMessage) computeLength() {
	var topicLength uint16
	switch m.TopicIDType {
	case TIT_STRING:
		topicLength = uint16(len(m.TopicName))
	case TIT_PREDEFINED, TIT_SHORT:
		topicLength = 2
	}
	m.Header.SetVarPartLength(unsubscribeHeaderLength + topicLength)
}

func (m *UnsubscribeMessage) encodeFlags() byte {
	var b byte
	b |= m.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (m *UnsubscribeMessage) decodeFlags(b byte) {
	m.TopicIDType = b & flagsTopicIDTypeBits
}

func (m *UnsubscribeMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.WriteByte(m.encodeFlags())
	buf.Write(encodeUint16(m.messageID))
	switch m.TopicIDType {
	case TIT_STRING:
		buf.Write(m.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		buf.Write(encodeUint16(m.TopicID))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (m *UnsubscribeMessage) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = readByte(r); err != nil {
		return
	}
	m.decodeFlags(flagsByte)

	if m.messageID, err = readUint16(r); err != nil {
		return
	}

	switch m.TopicIDType {
	case TIT_STRING:
		m.TopicID = 0
		m.TopicName = make([]byte, m.VarPartLength()-unsubscribeHeaderLength)
		_, err = io.ReadFull(r, m.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		m.TopicName = nil
		m.TopicID, err = readUint16(r)
	default:
		err = fmt.Errorf("invalid TopicIDType: %d", m.TopicIDType)
	}
	return
}

func (m UnsubscribeMessage) String() string {
	return fmt.Sprintf("UNSUBSCRIBE(TopicName=%#v, TopicID=%d, TopicIDType=%d, MessageID=%d)",
		string(m.TopicName), m.TopicID, m.TopicIDType, m.messageID)
}
