package packets1

import (
	"fmt"
	"io"
)

const pubackVarPartLength uint16 = 5

type Puback struct {
	Header
	MessageIDProperty
	TopicID    uint16
	ReturnCode ReturnCode
}

func NewPuback(topicID uint16, returnCode ReturnCode) *Puback {
	return &Puback{
		Header:     *NewHeader(PUBACK, pubackVarPartLength),
		TopicID:    topicID,
		ReturnCode: returnCode,
	}
}

func (p *Puback) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.Write(encodeUint16(p.TopicID))
	buf.Write(encodeUint16(p.messageID))
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Puback) Unpack(r io.Reader) (err error) {
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

func (p Puback) String() string {
	return fmt.Sprintf("PUBACK(TopicID=%d, ReturnCode=%d, MessageID=%d)", p.TopicID,
		p.ReturnCode, p.messageID)
}
