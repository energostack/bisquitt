package packets1

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const subackVarPartLength uint16 = 6

type Suback struct {
	pkts.Header
	MessageIDProperty
	QOS        uint8
	ReturnCode ReturnCode
	TopicID    uint16
}

func NewSuback(topicID uint16, returnCode ReturnCode, qos uint8) *Suback {
	return &Suback{
		Header:     *pkts.NewHeader(pkts.SUBACK, subackVarPartLength),
		QOS:        qos,
		ReturnCode: returnCode,
		TopicID:    topicID,
	}
}

func (p *Suback) encodeFlags() byte {
	var b byte
	b |= (p.QOS << 5) & flagsQOSBits
	return b
}

func (p *Suback) decodeFlags(b byte) {
	p.QOS = (b & flagsQOSBits) >> 5
}

func (p *Suback) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint16(p.TopicID))
	_, _ = buf.Write(pkts.EncodeUint16(p.messageID))
	_ = buf.WriteByte(byte(p.ReturnCode))

	return buf.Bytes(), nil
}

func (p *Suback) Unpack(buf []byte) error {
	if len(buf) != int(subackVarPartLength) {
		return fmt.Errorf("bad SUBACK packet length: expected %d, got %d",
			subackVarPartLength, len(buf))
	}

	p.decodeFlags(buf[0])

	p.TopicID = binary.BigEndian.Uint16(buf[1:3])
	p.messageID = binary.BigEndian.Uint16(buf[3:5])
	p.ReturnCode = ReturnCode(buf[5])

	return nil
}

func (p Suback) String() string {
	return fmt.Sprintf("SUBACK(TopicID=%d, MessageID=%d, ReturnCode=%d, QOS=%d)", p.TopicID,
		p.messageID, p.ReturnCode, p.QOS)
}
