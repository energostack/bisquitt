package packets2

type PacketIDProperty struct {
	packetID uint16
}

func (p *PacketIDProperty) CopyPacketID(m2 PacketWithPacketID) {
	p.packetID = m2.PacketID()
}

func (p *PacketIDProperty) SetPacketID(msgID uint16) {
	p.packetID = msgID
}

func (p *PacketIDProperty) PacketID() uint16 {
	return p.packetID
}

// PacketWithPacketID is an interface for all packets which include PacketID property.
type PacketWithPacketID interface {
	PacketID() uint16
	SetPacketID(msgID uint16)
}
