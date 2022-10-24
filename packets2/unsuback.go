package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const unsubackVarPartLength uint16 = 3

type Unsuback struct {
	pkts.Header
	PacketV2
	// Fields
	PacketIDProperty
	ReasonCode ReasonCode
}

func NewUnsuback(reasonCode ReasonCode) *Unsuback {
	return &Unsuback{
		Header:     *pkts.NewHeader(pkts.UNSUBACK, unsubackVarPartLength),
		ReasonCode: reasonCode,
	}
}

func (p *Unsuback) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))
	_ = buf.WriteByte(uint8(p.ReasonCode))

	return buf.Bytes(), nil
}

func (p *Unsuback) Unpack(buf []byte) error {
	if len(buf) != int(unsubackVarPartLength) {
		return fmt.Errorf("bad UNSUBACK2 packet length: expected %d, got %d",
			unsubackVarPartLength, len(buf))
	}

	p.packetID = binary.BigEndian.Uint16(buf[0:2])
	p.ReasonCode = ReasonCode(buf[2])

	return nil
}

func (p Unsuback) String() string {
	return fmt.Sprintf("UNSUBACK2(ReasonCode=%s, PacketID=%d)", p.ReasonCode, p.packetID)
}
