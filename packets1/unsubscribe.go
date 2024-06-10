package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

const unsubscribeHeaderLength uint16 = 3

type Unsubscribe struct {
	pkts.Header
	// Flags
	TopicIDType uint8
	// Fields
	MessageIDProperty
	TopicID   uint16
	TopicName string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewUnsubscribe(topicName string, topicID uint16, topicIDType uint8) *Unsubscribe {
	p := &Unsubscribe{
		Header:      *pkts.NewHeader(pkts.UNSUBSCRIBE, 0),
		TopicIDType: topicIDType,
		TopicID:     topicID,
		TopicName:   topicName,
	}
	p.computeLength()
	return p
}

func (p *Unsubscribe) computeLength() {
	var topicLength uint16
	switch p.TopicIDType {
	case TIT_STRING:
		topicLength = uint16(len(p.TopicName))
	case TIT_PREDEFINED, TIT_SHORT:
		topicLength = 2
	}
	p.Header.SetVarPartLength(unsubscribeHeaderLength + topicLength)
}

func (p *Unsubscribe) encodeFlags() byte {
	var b byte
	b |= p.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (p *Unsubscribe) decodeFlags(b byte) {
	p.TopicIDType = b & flagsTopicIDTypeBits
}

func (p *Unsubscribe) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))
	switch p.TopicIDType {
	case TIT_STRING:
		_, _ = buf.Write([]byte(p.TopicName))
	case TIT_PREDEFINED, TIT_SHORT:
		_, _ = buf.Write(pkts.EncodeUint16(p.TopicID))
	}

	return buf.Bytes(), nil
}

func (p *Unsubscribe) Unpack(buf []byte) error {
	if len(buf) <= int(unsubscribeHeaderLength) {
		return fmt.Errorf("bad UNSUBSCRIBE packet length: expected >%d, got %d",
			unsubscribeHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.messageID = binary.BigEndian.Uint16(buf[1:3])

	switch p.TopicIDType {
	case TIT_STRING:
		p.TopicID = 0
		p.TopicName = string(buf[3:])
	case TIT_PREDEFINED, TIT_SHORT:
		if len(buf) != int(unsubscribeHeaderLength+2) {
			return fmt.Errorf("bad UNSUBSCRIBE packet length: expected %d, got %d",
				unsubscribeHeaderLength+2, len(buf))
		}
		p.TopicName = ""
		p.TopicID = binary.BigEndian.Uint16(buf[3:5])
	default:
		return fmt.Errorf("invalid TopicIDType: %d", p.TopicIDType)
	}

	return nil
}

func (p Unsubscribe) String() string {
	switch p.TopicIDType {
	case TIT_STRING:
		return fmt.Sprintf("UNSUBSCRIBE(TopicName=%#v, MessageID=%d)",
			string(p.TopicName), p.messageID)
	case TIT_PREDEFINED:
		return fmt.Sprintf("UNSUBSCRIBE(TopicID(p)=%d, MessageID=%d)",
			p.TopicID, p.messageID)
	case TIT_SHORT:
		topicName := pkts.DecodeShortTopic(p.TopicID)
		return fmt.Sprintf("UNSUBSCRIBE(TopicName(s)=%#v, MessageID=%d)",
			topicName, p.messageID)
	default:
		return fmt.Sprintf("UNSUBSCRIBE(TopicName=%#v, TopicID=%d, TopicIDType=%d (INVALID!), MessageID=%d)",
			p.TopicName, p.TopicID, p.TopicIDType, p.messageID)
	}
}
