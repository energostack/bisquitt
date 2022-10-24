package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const pubackVarPartLength uint16 = 3

type Puback struct {
	pkts.Header
	PacketV2
	// Fields
	PacketIDProperty
	ReasonCode ReasonCode
}

func NewPuback(reasonCode ReasonCode) *Puback {
	return &Puback{
		Header:     *pkts.NewHeader(pkts.PUBACK, pubackVarPartLength),
		ReasonCode: reasonCode,
	}
}

func (p *Puback) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))
	_ = buf.WriteByte(byte(p.ReasonCode))

	return buf.Bytes(), nil
}

func (p *Puback) Unpack(buf []byte) error {
	if len(buf) != int(pubackVarPartLength) {
		return fmt.Errorf("bad PUBACK2 packet length: expected %d, got %d",
			pubackVarPartLength, len(buf))
	}

	p.packetID = binary.BigEndian.Uint16(buf[0:2])
	p.ReasonCode = ReasonCode(buf[2])

	return nil
}

func (p Puback) String() string {
	return fmt.Sprintf("PUBACK2(ReasonCode=%s, PacketID=%d)",
		p.ReasonCode, p.packetID)
}
