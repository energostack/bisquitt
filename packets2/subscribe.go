package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const subscribeHeaderLength uint16 = 3

type Subscribe struct {
	pkts.Header
	PacketV2
	// Flags
	NoLocal           bool
	QOS               uint8
	RetainAsPublished bool
	RetainHandling    RetainHandling
	TopicAliasType    TopicAliasType
	// Fields
	PacketIDProperty
	TopicAlias  uint16
	TopicFilter string
}

// Retain handling constants.
type RetainHandling uint8

const (
	RH_SEND_AT_SUBSCRIBE RetainHandling = iota
	RH_SEND_AT_NEW_SUBSCRIBE
	RH_DONT_SEND
)

func (rh RetainHandling) String() string {
	switch rh {
	case RH_SEND_AT_SUBSCRIBE:
		return "send_at_subscribe"
	case RH_SEND_AT_NEW_SUBSCRIBE:
		return "send_at_new_subscribe"
	case RH_DONT_SEND:
		return "dont_send"
	default:
		return fmt.Sprintf("illegal(%d)", rh)
	}
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewSubscribe(topicAlias uint16, topicFilter string, noLocal bool, qos uint8,
	retainedAsPublished bool, retainHandling RetainHandling,
	topicAliasType TopicAliasType) *Subscribe {
	p := &Subscribe{
		Header:            *pkts.NewHeader(pkts.SUBSCRIBE, 0),
		NoLocal:           noLocal,
		QOS:               qos,
		RetainAsPublished: retainedAsPublished,
		RetainHandling:    retainHandling,
		TopicAliasType:    topicAliasType,
		TopicAlias:        topicAlias,
		TopicFilter:       topicFilter,
	}
	p.computeLength()
	return p
}

func (p *Subscribe) computeLength() {
	var topicLength uint16
	switch p.TopicAliasType {
	case TAT_NORMAL, TAT_PREDEFINED, TAT_SHORT:
		topicLength = 2
	case TAT_LONG:
		topicLength = uint16(len(p.TopicFilter))
	default:
		panic(fmt.Sprintf("invalid TopicAliasType value: %d", p.TopicAliasType))
	}
	p.Header.SetVarPartLength(subscribeHeaderLength + topicLength)
}

func (p *Subscribe) encodeFlags() byte {
	var b byte

	if p.NoLocal {
		b |= flagsNoLocalBit
	}
	b |= (p.QOS << 5) & flagsQOSBits
	if p.RetainAsPublished {
		b |= flagsRetainAsPublishedBit
	}
	b |= (uint8(p.RetainHandling) << 2) & flagsRetainHandlingBits
	b |= uint8(p.TopicAliasType) & flagsTopicAliasTypeBits

	return b
}

func (p *Subscribe) decodeFlags(b byte) {
	p.NoLocal = (b & flagsNoLocalBit) != 0
	p.QOS = (b & flagsQOSBits) >> 5
	p.RetainAsPublished = (b & flagsRetainAsPublishedBit) != 0
	p.RetainHandling = RetainHandling((b & flagsRetainHandlingBits) >> 2)
	p.TopicAliasType = TopicAliasType(b & flagsTopicAliasTypeBits)
}

func (p *Subscribe) Pack() ([]byte, error) {
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

func (p *Subscribe) Unpack(buf []byte) error {
	if len(buf) <= int(subscribeHeaderLength) {
		return fmt.Errorf("bad SUBSCRIBE2 packet length: expected >%d, got %d",
			subscribeHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.packetID = binary.BigEndian.Uint16(buf[1:3])

	switch p.TopicAliasType {
	case TAT_NORMAL, TAT_PREDEFINED, TAT_SHORT:
		if len(buf) != int(subscribeHeaderLength+2) {
			return fmt.Errorf("bad SUBSCRIBE2 packet length: expected %d, got %d",
				subscribeHeaderLength+2, len(buf))
		}
		p.TopicFilter = ""
		p.TopicAlias = binary.BigEndian.Uint16(buf[subscribeHeaderLength:])
	case TAT_LONG:
		p.TopicFilter = string(buf[subscribeHeaderLength:])
		p.TopicAlias = 0
	default:
		// Should never be reached because TAT is only 2 bits long.
		panic(fmt.Sprintf("invalid TopicAliasType value: %d", p.TopicAliasType))
	}

	return nil
}

func (p Subscribe) String() string {
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
	return fmt.Sprintf("SUBSCRIBE2(%s, QOS=%d, PacketID=%d, NoLocal=%t, RAP=%t, RH=%s)",
		topicStr, p.QOS, p.packetID, p.NoLocal, p.RetainAsPublished, p.RetainHandling)
}
