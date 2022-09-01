package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const unsubscribeHeaderLength uint16 = 3

type Unsubscribe struct {
	pkts.Header
	MessageIDProperty
	TopicIDType uint8
	TopicID     uint16
	TopicName   []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewUnsubscribe(topicID uint16, topicIDType uint8, topicName []byte) *Unsubscribe {
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

func (p *Unsubscribe) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.WriteByte(p.encodeFlags())
	buf.Write(pkts.EncodeUint16(p.messageID))
	switch p.TopicIDType {
	case TIT_STRING:
		buf.Write(p.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		buf.Write(pkts.EncodeUint16(p.TopicID))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (p *Unsubscribe) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = pkts.ReadByte(r); err != nil {
		return
	}
	p.decodeFlags(flagsByte)

	if p.messageID, err = pkts.ReadUint16(r); err != nil {
		return
	}

	switch p.TopicIDType {
	case TIT_STRING:
		p.TopicID = 0
		p.TopicName = make([]byte, p.VarPartLength()-unsubscribeHeaderLength)
		_, err = io.ReadFull(r, p.TopicName)
	case TIT_PREDEFINED, TIT_SHORT:
		p.TopicName = nil
		p.TopicID, err = pkts.ReadUint16(r)
	default:
		err = fmt.Errorf("invalid TopicIDType: %d", p.TopicIDType)
	}
	return
}

func (p Unsubscribe) String() string {
	return fmt.Sprintf("UNSUBSCRIBE(TopicName=%#v, TopicID=%d, TopicIDType=%d, MessageID=%d)",
		string(p.TopicName), p.TopicID, p.TopicIDType, p.messageID)
}
