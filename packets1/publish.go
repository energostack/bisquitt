package packets1

import (
	"encoding/binary"
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const publishHeaderLength uint16 = 5

type Publish struct {
	pkts.Header
	MessageIDProperty
	pkts.DUPProperty
	Retain      bool
	QOS         uint8
	TopicIDType uint8
	TopicID     uint16
	Data        []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewPublish(topicID uint16, topicIDType uint8, payload []byte, qos uint8,
	retain bool, dup bool) *Publish {
	p := &Publish{
		Header:      *pkts.NewHeader(pkts.PUBLISH, 0),
		DUPProperty: *pkts.NewDUPProperty(dup),
		TopicID:     topicID,
		TopicIDType: topicIDType,
		Data:        payload,
		QOS:         qos,
		Retain:      retain,
	}
	p.computeLength()
	return p
}

func (p *Publish) computeLength() {
	payloadLen := uint16(len(p.Data))
	p.Header.SetVarPartLength(publishHeaderLength + payloadLen)
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
	b |= p.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (p *Publish) decodeFlags(b byte) {
	p.SetDUP((b & flagsDUPBit) == flagsDUPBit)
	p.QOS = (b & flagsQOSBits) >> 5
	p.Retain = (b & flagsRetainBit) == flagsRetainBit
	p.TopicIDType = b & flagsTopicIDTypeBits
}

func (p *Publish) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.WriteByte(p.encodeFlags())
	buf.Write(pkts.EncodeUint16(p.TopicID))
	buf.Write(pkts.EncodeUint16(p.messageID))
	buf.Write(p.Data)

	_, err := buf.WriteTo(w)
	return err
}

func (p *Publish) Unpack(buf []byte) error {
	if len(buf) < int(publishHeaderLength) {
		return fmt.Errorf("bad PUBLISH packet length: expected >=%d, got %d",
			publishHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.TopicID = binary.BigEndian.Uint16(buf[1:3])
	p.messageID = binary.BigEndian.Uint16(buf[3:5])
	p.Data = buf[5:]

	return nil
}

func (p Publish) String() string {
	var topicIDType string
	switch p.TopicIDType {
	case TIT_REGISTERED:
		topicIDType = "r"
	case TIT_PREDEFINED:
		topicIDType = "p"
	case TIT_SHORT:
		topicIDType = "s"
	}
	return fmt.Sprintf("PUBLISH(TopicID(%s)=%d, Data=%#v, QOS=%d, Retain=%t, MessageID=%d, Dup=%t)",
		topicIDType, p.TopicID, string(p.Data), p.QOS, p.Retain, p.messageID, p.DUP())
}
