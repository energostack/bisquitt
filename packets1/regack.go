package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const regackVarPartLength uint16 = 5

type Regack struct {
	pkts.Header
	// Fields
	TopicID uint16
	MessageIDProperty
	ReturnCode ReturnCode
}

func NewRegack(topicID uint16, returnCode ReturnCode) *Regack {
	return &Regack{
		Header:     *pkts.NewHeader(pkts.REGACK, regackVarPartLength),
		TopicID:    topicID,
		ReturnCode: returnCode,
	}
}

func (p *Regack) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.TopicID))
	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))
	_ = buf.WriteByte(byte(p.ReturnCode))

	return buf.Bytes(), nil
}

func (p *Regack) Unpack(buf []byte) error {
	if len(buf) != int(regackVarPartLength) {
		return fmt.Errorf("bad REGACK packet length: expected %d, got %d",
			regackVarPartLength, len(buf))
	}

	p.TopicID = binary.BigEndian.Uint16(buf[0:2])
	p.messageID = binary.BigEndian.Uint16(buf[2:4])
	p.ReturnCode = ReturnCode(buf[4])

	return nil
}

func (p Regack) String() string {
	return fmt.Sprintf("REGACK(TopicID=%d, ReturnCode=%d, MessageID=%d)", p.TopicID,
		p.ReturnCode, p.messageID)
}
