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
	p := &Subscribe{
		Header:      *NewHeader(SUBSCRIBE, 0),
		DUPProperty: DUPProperty{dup},
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
	if p.dup {
		b |= flagsDUPBit
	}
	b |= (p.QOS << 5) & flagsQOSBits
	b |= p.TopicIDType & flagsTopicIDTypeBits
	return b
}

func (p *Subscribe) decodeFlags(b byte) {
	p.dup = (b & flagsDUPBit) == flagsDUPBit
	p.QOS = (b & flagsQOSBits) >> 5
	p.TopicIDType = b & flagsTopicIDTypeBits
}

func (p *Subscribe) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.pack()
	buf.WriteByte(p.encodeFlags())
	buf.Write(encodeUint16(p.messageID))
	switch p.TopicIDType {
	case TIT_STRING:
		buf.Write(p.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		buf.Write(encodeUint16(p.TopicID))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (p *Subscribe) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = readByte(r); err != nil {
		return
	}
	p.decodeFlags(flagsByte)

	if p.messageID, err = readUint16(r); err != nil {
		return
	}

	switch p.TopicIDType {
	case TIT_STRING:
		p.TopicID = 0
		p.TopicName = make([]byte, p.VarPartLength()-subscribeHeaderLength)
		_, err = io.ReadFull(r, p.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		p.TopicName = nil
		p.TopicID, err = readUint16(r)
	default:
		err = fmt.Errorf("invalid TopicIDType: %d", p.TopicIDType)
	}
	return
}

func (p Subscribe) String() string {
	return fmt.Sprintf("SUBSCRIBE(TopicName=%#v, QOS=%d, TopicID=%d, TopicIDType=%d, MessageID=%d, Dup=%t)",
		string(p.TopicName), p.QOS, p.TopicID, p.TopicIDType, p.messageID, p.dup)
}
