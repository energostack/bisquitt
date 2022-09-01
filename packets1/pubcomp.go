package packets1

import (
	"encoding/binary"
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pubcompVarPartLength uint16 = 2

type Pubcomp struct {
	pkts.Header
	MessageIDProperty
}

func NewPubcomp() *Pubcomp {
	return &Pubcomp{
		Header: *pkts.NewHeader(pkts.PUBCOMP, pubcompVarPartLength),
	}
}

func (p *Pubcomp) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.Write(pkts.EncodeUint16(p.messageID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Pubcomp) Unpack(buf []byte) error {
	if len(buf) != int(pubcompVarPartLength) {
		return fmt.Errorf("bad PUBCOMP packet length: expected %d, got %d",
			pubcompVarPartLength, len(buf))
	}

	p.messageID = binary.BigEndian.Uint16(buf)

	return nil
}

func (p Pubcomp) String() string {
	return fmt.Sprintf("PUBCOMP(MessageID=%d)", p.messageID)
}
