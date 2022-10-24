package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const registerHeaderLength uint16 = 4

type Register struct {
	pkts.Header
	PacketV2
	// Fields
	TopicAlias uint16
	PacketIDProperty
	TopicName string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewRegister(alias uint16, topic string) *Register {
	p := &Register{
		Header:     *pkts.NewHeader(pkts.REGISTER, 0),
		TopicAlias: alias,
		TopicName:  topic,
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

	_, _ = buf.Write(pkts.EncodeUint16(p.TopicAlias))
	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))
	_, _ = buf.Write([]byte(p.TopicName))

	return buf.Bytes(), nil
}

func (p *Register) Unpack(buf []byte) error {
	if len(buf) <= int(registerHeaderLength) {
		return fmt.Errorf("bad REGISTER2 packet length: expected >%d, got %d",
			registerHeaderLength, len(buf))
	}

	p.TopicAlias = binary.BigEndian.Uint16(buf[0:2])
	p.packetID = binary.BigEndian.Uint16(buf[2:4])
	p.TopicName = string(buf[4:])

	return nil
}

func (p Register) String() string {
	return fmt.Sprintf("REGISTER2(Topic=%q, Alias=%d, PacketID=%d)", string(p.TopicName),
		p.TopicAlias, p.packetID)
}
