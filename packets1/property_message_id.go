package packets1

type MessageIDProperty struct {
	messageID uint16
}

func (p *MessageIDProperty) CopyMessageID(m2 PacketWithID) {
	p.messageID = m2.MessageID()
}

func (p *MessageIDProperty) SetMessageID(msgID uint16) {
	p.messageID = msgID
}

func (p *MessageIDProperty) MessageID() uint16 {
	return p.messageID
}

// PacketWithID is an interface for all packets which include MessageID property.
type PacketWithID interface {
	MessageID() uint16
	SetMessageID(msgID uint16)
}
