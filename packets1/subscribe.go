package packets1

import (
	"fmt"
	"io"
)

const subscribeHeaderLength uint16 = 3

type Subscribe struct {
	Header
	DUPProperty
	MessageIDProperty
	QOS         uint8
	TopicIDType uint8
	TopicID     uint16
	TopicName   []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewSubscribe(topicID uint16, topicIDType uint8, topicName []byte, qos uint8, dup bool) *Subscribe {
	m := &Subscribe{
		Header:      *NewHeader(SUBSCRIBE, 0),
		DUPProperty: DUPProperty{dup},
		QOS:         qos,
		TopicIDType: topicIDType,
		TopicID:     topicID,
		TopicName:   topicName,
	}
	m.computeLength()
	return m
}

func (m *Subscribe) computeLength() {
	var topicLength uint16
	switch m.TopicIDType {
	case TIT_STRING:
		topicLength = uint16(len(m.TopicName))
	case TIT_PREDEFINED, TIT_SHORT:
		topicLength = 2
	}
	m.Header.SetVarPartLength(subscribeHeaderLength + topicLength)
}

func (m *Subscribe) encodeFlags() byte {
	var b byte
	if m.dup {
		b |= flagsDUPBit
	}
	b |= (m.QOS << 5) & flagsQOSBits
	b |= m.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (m *Subscribe) decodeFlags(b byte) {
	m.dup = (b & flagsDUPBit) == flagsDUPBit
	m.QOS = (b & flagsQOSBits) >> 5
	m.TopicIDType = b & flagsTopicIDTypeBits
}

func (m *Subscribe) Write(w io.Writer) error {
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

func (m *Subscribe) Unpack(r io.Reader) (err error) {
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
		m.TopicName = make([]byte, m.VarPartLength()-subscribeHeaderLength)
		_, err = io.ReadFull(r, m.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		m.TopicName = nil
		m.TopicID, err = readUint16(r)
	default:
		err = fmt.Errorf("invalid TopicIDType: %d", m.TopicIDType)
	}
	return
}

func (m Subscribe) String() string {
	return fmt.Sprintf("SUBSCRIBE(TopicName=%#v, QOS=%d, TopicID=%d, TopicIDType=%d, MessageID=%d, Dup=%t)",
		string(m.TopicName), m.QOS, m.TopicID, m.TopicIDType, m.messageID, m.dup)
}
