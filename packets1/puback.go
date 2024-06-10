package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

const pubackVarPartLength uint16 = 5

type Puback struct {
	pkts.Header
	// Fields
	TopicID uint16
	MessageIDProperty
	ReturnCode ReturnCode
}

func NewPuback(topicID uint16, returnCode ReturnCode) *Puback {
	return &Puback{
		Header:     *pkts.NewHeader(pkts.PUBACK, pubackVarPartLength),
		TopicID:    topicID,
		ReturnCode: returnCode,
	}
}

func (p *Puback) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.TopicID))
	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))
	_ = buf.WriteByte(byte(p.ReturnCode))

	return buf.Bytes(), nil
}

func (p *Puback) Unpack(buf []byte) error {
	if len(buf) != int(pubackVarPartLength) {
		return fmt.Errorf("bad PUBACK packet length: expected %d, got %d",
			pubackVarPartLength, len(buf))
	}

	p.TopicID = binary.BigEndian.Uint16(buf[0:2])
	p.messageID = binary.BigEndian.Uint16(buf[2:4])
	p.ReturnCode = ReturnCode(buf[4])

	return nil
}

func (p Puback) String() string {
	return fmt.Sprintf("PUBACK(TopicID=%d, ReturnCode=%d, MessageID=%d)", p.TopicID,
		p.ReturnCode, p.messageID)
}
