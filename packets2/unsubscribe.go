package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const unsubscribeHeaderLength uint16 = 3

type Unsubscribe struct {
	pkts.Header
	PacketV2
	// Flags
	TopicAliasType TopicAliasType
	// Fields
	PacketIDProperty
	TopicAlias  uint16
	TopicFilter string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewUnsubscribe(topicAlias uint16, topicFilter string, topicAliasType TopicAliasType) *Unsubscribe {
	p := &Unsubscribe{
		Header:         *pkts.NewHeader(pkts.UNSUBSCRIBE, 0),
		TopicAliasType: topicAliasType,
		TopicAlias:     topicAlias,
		TopicFilter:    topicFilter,
	}
	p.computeLength()
	return p
}

func (p *Unsubscribe) computeLength() {
	var topicLength uint16
	switch p.TopicAliasType {
	case TAT_NORMAL, TAT_PREDEFINED, TAT_SHORT:
		topicLength = 2
	case TAT_LONG:
		topicLength = uint16(len(p.TopicFilter))
	default:
		panic(fmt.Sprintf("invalid TopicAliasType value: %d", p.TopicAliasType))
	}
	p.Header.SetVarPartLength(unsubscribeHeaderLength + topicLength)
}

func (p *Unsubscribe) encodeFlags() byte {
	var b byte
	b |= uint8(p.TopicAliasType) & flagsTopicAliasTypeBits
	return b
}

func (p *Unsubscribe) decodeFlags(b byte) {
	p.TopicAliasType = TopicAliasType(b & flagsTopicAliasTypeBits)
}

func (p *Unsubscribe) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))
	switch p.TopicAliasType {
	case TAT_NORMAL, TAT_PREDEFINED, TAT_SHORT:
		_, _ = buf.Write(pkts.EncodeUint16(p.TopicAlias))
	case TAT_LONG:
		_, _ = buf.Write([]byte(p.TopicFilter))
	default:
		// Should never be reached because TAT is already checked in computeLength().
		panic(fmt.Sprintf("invalid TopicAliasType value: %d", p.TopicAliasType))
	}

	return buf.Bytes(), nil
}

func (p *Unsubscribe) Unpack(buf []byte) error {
	if len(buf) <= int(unsubscribeHeaderLength) {
		return fmt.Errorf("bad UNSUBSCRIBE2 packet length: expected >%d, got %d",
			unsubscribeHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.packetID = binary.BigEndian.Uint16(buf[1:3])

	switch p.TopicAliasType {
	case TAT_NORMAL, TAT_PREDEFINED, TAT_SHORT:
		if len(buf) != int(unsubscribeHeaderLength+2) {
			return fmt.Errorf("bad UNSUBSCRIBE2 packet length: expected %d, got %d",
				unsubscribeHeaderLength+2, len(buf))
		}
		p.TopicFilter = ""
		p.TopicAlias = binary.BigEndian.Uint16(buf[subscribeHeaderLength:])
	case TAT_LONG:
		p.TopicAlias = 0
		p.TopicFilter = string(buf[subscribeHeaderLength:])
	default:
		// Should never be reached because TAT is only 2 bits long.
		panic(fmt.Sprintf("invalid TopicAliasType value: %d", p.TopicAliasType))
	}

	return nil
}

func (p Unsubscribe) String() string {
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
			p.TopicAliasType, p.TopicFilter)
	default:
		topicStr = fmt.Sprintf("Topic=%q, TopicAlias=%d, TopicAliasType=%s",
			p.TopicFilter, p.TopicAlias, p.TopicAliasType)
	}
	return fmt.Sprintf("UNSUBSCRIBE2(%s, PacketID=%d)", topicStr, p.packetID)
}
