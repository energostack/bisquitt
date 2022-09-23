package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const subscribeHeaderLength uint16 = 3

type Subscribe struct {
	pkts.Header
	pkts.DUPProperty
	MessageIDProperty
	QOS         uint8
	TopicIDType uint8
	TopicID     uint16
	TopicName   string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewSubscribe(topicID uint16, topicIDType uint8, topicName string, qos uint8, dup bool) *Subscribe {
	p := &Subscribe{
		Header:      *pkts.NewHeader(pkts.SUBSCRIBE, 0),
		DUPProperty: *pkts.NewDUPProperty(dup),
		QOS:         qos,
		TopicIDType: topicIDType,
		TopicID:     topicID,
		TopicName:   topicName,
	}
	p.computeLength()
	return p
}

func (p *Subscribe) computeLength() {
	var topicLength uint16
	switch p.TopicIDType {
	case TIT_STRING:
		topicLength = uint16(len(p.TopicName))
	case TIT_PREDEFINED, TIT_SHORT:
		topicLength = 2
	}
	p.Header.SetVarPartLength(subscribeHeaderLength + topicLength)
}

func (p *Subscribe) encodeFlags() byte {
	var b byte
	if p.DUP() {
		b |= flagsDUPBit
	}
	b |= (p.QOS << 5) & flagsQOSBits
	b |= p.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (p *Subscribe) decodeFlags(b byte) {
	p.SetDUP((b & flagsDUPBit) == flagsDUPBit)
	p.QOS = (b & flagsQOSBits) >> 5
	p.TopicIDType = b & flagsTopicIDTypeBits
}

func (p *Subscribe) Pack() ([]byte, error) {
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

func (p *Subscribe) Unpack(buf []byte) error {
	if len(buf) <= int(subscribeHeaderLength) {
		return fmt.Errorf("bad SUBSCRIBE packet length: expected >%d, got %d",
			subscribeHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.messageID = binary.BigEndian.Uint16(buf[1:3])

	switch p.TopicIDType {
	case TIT_STRING:
		p.TopicID = 0
		p.TopicName = string(buf[3:])
	case TIT_PREDEFINED, TIT_SHORT:
		if len(buf) != int(subscribeHeaderLength+2) {
			return fmt.Errorf("bad SUBSCRIBE packet length: expected %d, got %d",
				subscribeHeaderLength+2, len(buf))
		}
		p.TopicName = ""
		p.TopicID = binary.BigEndian.Uint16(buf[3:5])
	default:
		return fmt.Errorf("invalid TopicIDType: %d", p.TopicIDType)
	}

	return nil
}

func (p Subscribe) String() string {
	return fmt.Sprintf("SUBSCRIBE(TopicName=%#v, QOS=%d, TopicID=%d, TopicIDType=%d, MessageID=%d, Dup=%t)",
		string(p.TopicName), p.QOS, p.TopicID, p.TopicIDType, p.messageID, p.DUP())
}
