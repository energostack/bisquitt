package packets1

type MessageIDProperty struct {
	messageID uint16
}

func (m *MessageIDProperty) CopyMessageID(m2 PacketWithID) {
	m.messageID = m2.MessageID()
}

func (m *MessageIDProperty) SetMessageID(msgID uint16) {
	m.messageID = msgID
}

func (m *MessageIDProperty) MessageID() uint16 {
	return m.messageID
}

// PacketWithID is an interface for all packets which include MessageID property.
type PacketWithID interface {
	MessageID() uint16
	SetMessageID(msgID uint16)
}
