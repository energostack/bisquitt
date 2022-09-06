package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const registerHeaderLength uint16 = 4

type Register struct {
	pkts.Header
	MessageIDProperty
	TopicID   uint16
	TopicName string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewRegister(topicID uint16, topicName string) *Register {
	p := &Register{
		Header:    *pkts.NewHeader(pkts.REGISTER, 0),
		TopicID:   topicID,
		TopicName: topicName,
	}
	p.computeLength()
	return p
}

func (p *Register) computeLength() {
	topicLength := uint16(len(p.TopicName))
	p.Header.SetVarPartLength(registerHeaderLength + topicLength)
}

func (p *Register) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.TopicID))
	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))
	_, _ = buf.Write([]byte(p.TopicName))

	return buf.Bytes(), nil
}

func (p *Register) Unpack(buf []byte) error {
	if len(buf) <= int(registerHeaderLength) {
		return fmt.Errorf("bad REGISTER packet length: expected >%d, got %d",
			registerHeaderLength, len(buf))
	}

	p.TopicID = binary.BigEndian.Uint16(buf[0:2])
	p.messageID = binary.BigEndian.Uint16(buf[2:4])
	p.TopicName = string(buf[4:])

	return nil
}

func (p Register) String() string {
	return fmt.Sprintf("REGISTER(TopicName=%#v, TopicID=%d, MessageID=%d)", string(p.TopicName),
		p.TopicID, p.messageID)
}
