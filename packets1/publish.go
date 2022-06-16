package packets1

import (
	"fmt"
	"io"
)

const publishHeaderLength uint16 = 5

type Publish struct {
	Header
	MessageIDProperty
	DUPProperty
	Retain      bool
	QOS         uint8
	TopicIDType uint8
	TopicID     uint16
	Data        []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewPublish(topicID uint16, topicIDType uint8, payload []byte, qos uint8,
	retain bool, dup bool) *Publish {
	m := &Publish{
		Header:      *NewHeader(PUBLISH, 0),
		DUPProperty: DUPProperty{dup},
		TopicID:     topicID,
		TopicIDType: topicIDType,
		Data:        payload,
		QOS:         qos,
		Retain:      retain,
	}
	m.computeLength()
	return m
}

func (m *Publish) computeLength() {
	payloadLen := uint16(len(m.Data))
	m.Header.SetVarPartLength(publishHeaderLength + payloadLen)
}

func (m *Publish) encodeFlags() byte {
	var b byte
	if m.dup {
		b |= flagsDUPBit
	}
	b |= (m.QOS << 5) & flagsQOSBits
	if m.Retain {
		b |= flagsRetainBit
	}
	b |= m.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (m *Publish) decodeFlags(b byte) {
	m.dup = (b & flagsDUPBit) == flagsDUPBit
	m.QOS = (b & flagsQOSBits) >> 5
	m.Retain = (b & flagsRetainBit) == flagsRetainBit
	m.TopicIDType = b & flagsTopicIDTypeBits
}

func (m *Publish) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.WriteByte(m.encodeFlags())
	buf.Write(encodeUint16(m.TopicID))
	buf.Write(encodeUint16(m.messageID))
	buf.Write(m.Data)

	_, err := buf.WriteTo(w)
	return err
}

func (m *Publish) Unpack(r io.Reader) (err error) {
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

	m.Data = make([]byte, m.VarPartLength()-publishHeaderLength)
	_, err = io.ReadFull(r, m.Data)
	return
}

func (m Publish) String() string {
	var topicIDType string
	switch m.TopicIDType {
	case TIT_REGISTERED:
		topicIDType = "r"
	case TIT_PREDEFINED:
		topicIDType = "p"
	case TIT_SHORT:
		topicIDType = "s"
	}
	return fmt.Sprintf("PUBLISH(TopicID(%s)=%d, Data=%#v, QOS=%d, Retain=%t, MessageID=%d, Dup=%t)",
		topicIDType, m.TopicID, string(m.Data), m.QOS, m.Retain, m.messageID, m.dup)
}
