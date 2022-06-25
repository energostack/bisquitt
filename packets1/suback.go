package packets1

import (
	"fmt"
	"io"
)

const subackVarPartLength uint16 = 6

type Suback struct {
	Header
	MessageIDProperty
	QOS        uint8
	ReturnCode ReturnCode
	TopicID    uint16
}

func NewSuback(topicID uint16, qos uint8, returnCode ReturnCode) *Suback {
	return &Suback{
		Header:     *NewHeader(SUBACK, subackVarPartLength),
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

func (p *Suback) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.WriteByte(p.encodeFlags())
	buf.Write(encodeUint16(p.TopicID))
	buf.Write(encodeUint16(p.messageID))
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Suback) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = readByte(r); err != nil {
		return
	}
	p.decodeFlags(flagsByte)

	if p.TopicID, err = readUint16(r); err != nil {
		return
	}

	if p.messageID, err = readUint16(r); err != nil {
		return
	}

	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	p.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (p Suback) String() string {
	return fmt.Sprintf("SUBACK(TopicID=%d, MessageID=%d, ReturnCode=%d, QOS=%d)", p.TopicID,
		p.messageID, p.ReturnCode, p.QOS)
}
