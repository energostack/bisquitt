package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const regackVarPartLength uint16 = 6

const ()

type Regack struct {
	pkts.Header
	PacketV2
	// Flags
	TopicAliasType TopicAliasType
	// Fields
	TopicAlias uint16
	PacketIDProperty
	ReasonCode ReasonCode
}

func NewRegack(alias uint16, reasonCode ReasonCode, aliasType TopicAliasType) *Regack {
	return &Regack{
		Header:         *pkts.NewHeader(pkts.REGACK, regackVarPartLength),
		TopicAliasType: aliasType,
		TopicAlias:     alias,
		ReasonCode:     reasonCode,
	}
}

func (p *Regack) encodeFlags() byte {
	return byte(p.TopicAliasType) & flagsTopicAliasTypeBits
}

func (p *Regack) decodeFlags(b byte) {
	p.TopicAliasType = TopicAliasType(b & flagsTopicAliasTypeBits)
}

func (p *Regack) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint16(p.TopicAlias))
	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))
	_ = buf.WriteByte(byte(p.ReasonCode))

	return buf.Bytes(), nil
}

func (p *Regack) Unpack(buf []byte) error {
	if len(buf) != int(regackVarPartLength) {
		return fmt.Errorf("bad REGACK2 packet length: expected %d, got %d",
			regackVarPartLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.TopicAlias = binary.BigEndian.Uint16(buf[1:3])
	p.packetID = binary.BigEndian.Uint16(buf[3:5])
	p.ReasonCode = ReasonCode(buf[5])

	return nil
}

func (p Regack) String() string {
	return fmt.Sprintf("REGACK2(Alias(%s)=%d, ReasonCode=%d, PacketID=%d)",
		p.TopicAliasType, p.TopicAlias, p.ReasonCode, p.packetID)
}
