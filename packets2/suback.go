package packets2

import (
	"encoding/binary"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const subackVarPartLength uint16 = 6

type Suback struct {
	pkts.Header
	PacketV2
	// Flags
	GrantedQOS     uint8
	TopicAliasType TopicAliasType
	// Fields
	TopicAlias uint16
	PacketIDProperty
	ReasonCode ReasonCode
}

func NewSuback(alias uint16, reasonCode ReasonCode, grantedQOS uint8, aliasType TopicAliasType) *Suback {
	return &Suback{
		Header:         *pkts.NewHeader(pkts.SUBACK, subackVarPartLength),
		GrantedQOS:     grantedQOS,
		TopicAliasType: aliasType,
		TopicAlias:     alias,
		ReasonCode:     reasonCode,
	}
}

func (p *Suback) encodeFlags() byte {
	var b byte

	b |= (p.GrantedQOS << 5) & flagsQOSBits
	b |= uint8(p.TopicAliasType) & flagsTopicAliasTypeBits

	return b
}

func (p *Suback) decodeFlags(b byte) {
	p.GrantedQOS = (b & flagsQOSBits) >> 5
	p.TopicAliasType = TopicAliasType(b & flagsTopicAliasTypeBits)
}

func (p *Suback) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(pkts.EncodeUint16(p.TopicAlias))
	_, _ = buf.Write(pkts.EncodeUint16(p.packetID))
	_ = buf.WriteByte(byte(p.ReasonCode))

	return buf.Bytes(), nil
}

func (p *Suback) Unpack(buf []byte) error {
	if len(buf) != int(subackVarPartLength) {
		return fmt.Errorf("bad SUBACK2 packet length: expected %d, got %d",
			subackVarPartLength, len(buf))
	}

	p.decodeFlags(buf[0])

	p.TopicAlias = binary.BigEndian.Uint16(buf[1:3])
	p.packetID = binary.BigEndian.Uint16(buf[3:5])
	p.ReasonCode = ReasonCode(buf[5])

	return nil
}

func (p Suback) String() string {
	return fmt.Sprintf("SUBACK2(Alias(%s)=%d, ReasonCode=%d, QOS=%d, PacketID=%d)",
		p.TopicAliasType, p.TopicAlias, p.ReasonCode, p.GrantedQOS, p.packetID)
}
