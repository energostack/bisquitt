package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const regackVarPartLength uint16 = 5

type Regack struct {
	pkts.Header
	MessageIDProperty
	TopicID    uint16
	ReturnCode ReturnCode
}

func NewRegack(topicID uint16, returnCode ReturnCode) *Regack {
	return &Regack{
		Header:     *pkts.NewHeader(pkts.REGACK, regackVarPartLength),
		TopicID:    topicID,
		ReturnCode: returnCode,
	}
}

func (p *Regack) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.TopicID))
	buf.Write(pkts.EncodeUint16(p.messageID))
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Regack) Unpack(r io.Reader) (err error) {
	if p.TopicID, err = pkts.ReadUint16(r); err != nil {
		return
	}
	if p.messageID, err = pkts.ReadUint16(r); err != nil {
		return
	}
	var returnCodeByte uint8
	returnCodeByte, err = pkts.ReadByte(r)
	p.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (p Regack) String() string {
	return fmt.Sprintf("REGACK(TopicID=%d, ReturnCode=%d, MessageID=%d)", p.TopicID,
		p.ReturnCode, p.messageID)
}
