package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pubrecVarPartLength uint16 = 2

type Pubrec struct {
	pkts.Header
	PacketV2
	// Fields
	PacketIDProperty
}

func NewPubrec() *Pubrec {
	return &Pubrec{
		Header: *pkts.NewHeader(pkts.PUBREC, pubrecVarPartLength),
	}
}

func (p *Pubrec) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))

	return buf.Bytes(), nil
}

func (p *Pubrec) Unpack(buf []byte) error {
	if len(buf) != int(pubrecVarPartLength) {
		return fmt.Errorf("bad PUBREC2 packet length: expected %d, got %d",
			pubrecVarPartLength, len(buf))
	}

	p.packetID = binary.BigEndian.Uint16(buf)

	return nil
}

func (p Pubrec) String() string {
	return fmt.Sprintf("PUBREC2(PacketID=%d)", p.packetID)
}
