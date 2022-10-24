package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const publishHeaderLength uint16 = 5

type Publish struct {
	pkts.Header
	PacketV2
	// Flags
	pkts.DUPProperty
	QOS            uint8
	Retain         bool
	TopicAliasType TopicAliasType
	// Fields
	PacketIDProperty
	TopicAlias uint16
	TopicName  string
	Data       []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewPublish(topicAlias uint16, topicName string, data []byte,
	dup bool, qos uint8, retain bool, topicAliasType TopicAliasType) *Publish {
	p := &Publish{
		Header:         *pkts.NewHeader(pkts.PUBLISH, 0),
		DUPProperty:    *pkts.NewDUPProperty(dup),
		QOS:            qos,
		Retain:         retain,
		TopicAliasType: topicAliasType,
		TopicAlias:     topicAlias,
		TopicName:      topicName,
		Data:           data,
	}
	p.computeLength()
	return p
}

func (p *Publish) topicLength() uint16 {
	if p.TopicAliasType == TAT_LONG {
		return uint16(len(p.TopicName))
	}
	return 2
}

func (p *Publish) computeLength() {
	dataLen := uint16(len(p.Data))
	p.Header.SetVarPartLength(publishHeaderLength + p.topicLength() + dataLen)
}

func (p *Publish) encodeFlags() byte {
	var b byte

	if p.DUP() {
		b |= flagsDUPBit
	}
	b |= (p.QOS << 5) & flagsQOSBits
	if p.Retain {
		b |= flagsRetainBit
	}
	b |= uint8(p.TopicAliasType) & flagsTopicAliasTypeBits

	return b
}

func (p *Publish) decodeFlags(b byte) {
	p.SetDUP((b & flagsDUPBit) == flagsDUPBit)
	p.QOS = (b & flagsQOSBits) >> 5
	p.Retain = (b & flagsRetainBit) == flagsRetainBit
	p.TopicAliasType = TopicAliasType(b & flagsTopicAliasTypeBits)
}

func (p *Publish) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint16(p.topicLength()))
	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))

	if p.TopicAliasType == TAT_LONG {
		_, _ = buf.Write([]byte(p.TopicName))
	} else {
		_, _ = buf.Write(pkts.EncodeUint16(p.TopicAlias))
	}

	_, _ = buf.Write(p.Data)

	return buf.Bytes(), nil
}

func (p *Publish) Unpack(buf []byte) error {
	if len(buf) < int(publishHeaderLength) {
		return fmt.Errorf("bad PUBLISH2 packet length: expected >=%d, got %d",
			publishHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	topicLength := binary.BigEndian.Uint16(buf[1:3])
	p.packetID = binary.BigEndian.Uint16(buf[3:5])

	if len(buf) < int(publishHeaderLength+topicLength) {
		return fmt.Errorf("bad PUBLISH2 packet length: expected >=%d, got %d",
			publishHeaderLength+topicLength, len(buf))
	}

	if p.TopicAliasType == TAT_LONG {
		// MQTT 5 explicitly forbids zero-length topic:
		// It is a Protocol Error if the Topic Name is zero length and
		// there is no Topic Alias.
		// (MQTT 5.0 specification, 3.3.2.1 Topic Name)
		if topicLength == 0 {
			return fmt.Errorf("bad PUBLISH2 topic length: expected >0, got %d",
				topicLength)
		}
		p.TopicName = string(buf[publishHeaderLength : publishHeaderLength+topicLength])
		p.TopicAlias = 0
		p.Data = buf[publishHeaderLength+topicLength:]
	} else {
		if topicLength != 2 {
			return fmt.Errorf("bad PUBLISH2 topic length: expected 2, got %d",
				topicLength)
		}
		p.TopicName = ""
		p.TopicAlias = binary.BigEndian.Uint16(buf[publishHeaderLength : publishHeaderLength+2])
		p.Data = buf[publishHeaderLength+2:]
	}

	return nil
}

func (p Publish) String() string {
	var topicStr string
	switch p.TopicAliasType {
	case TAT_NORMAL:
		fallthrough
	case TAT_PREDEFINED:
		topicStr = fmt.Sprintf("TopicAlias(%s)=%d",
			p.TopicAliasType, p.TopicAlias)
	case TAT_SHORT:
		topicStr = fmt.Sprintf("Topic(%s)=%q",
			p.TopicAliasType, pkts.DecodeShortTopic(p.TopicAlias))
	case TAT_LONG:
		topicStr = fmt.Sprintf("Topic(%s)=%q",
			p.TopicAliasType, p.TopicName)
	default:
		topicStr = fmt.Sprintf("Topic=%q, TopicAlias=%d, TopicAliasType=%s",
			p.TopicName, p.TopicAlias, p.TopicAliasType)
	}
	return fmt.Sprintf("PUBLISH2(%s, Data=%#v, QOS=%d, Retain=%t, PacketID=%d, Dup=%t)",
		topicStr, string(p.Data), p.QOS, p.Retain, p.packetID, p.DUP())
}
