package packets1

import (
	"fmt"
	"io"

	"github.com/energomonitor/bisquitt/packets"
	pkts "github.com/energomonitor/bisquitt/packets"
)

const registerHeaderLength uint16 = 4

type Register struct {
	packets.Header
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

func (p *Register) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.TopicID))
	buf.Write(pkts.EncodeUint16(p.messageID))
	buf.Write([]byte(p.TopicName))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Register) Unpack(r io.Reader) (err error) {
	if p.TopicID, err = pkts.ReadUint16(r); err != nil {
		return
	}

	if p.messageID, err = pkts.ReadUint16(r); err != nil {
		return
	}

	topic := make([]byte, p.VarPartLength()-registerHeaderLength)
	if _, err = io.ReadFull(r, topic); err != nil {
		return
	}
	p.TopicName = string(topic)
	return
}

func (p Register) String() string {
	return fmt.Sprintf("REGISTER(TopicName=%#v, TopicID=%d, MessageID=%d)", string(p.TopicName),
		p.TopicID, p.messageID)
}
