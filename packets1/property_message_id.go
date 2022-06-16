package packets1

type MessageIDProperty struct {
	messageID uint16
}

func (m *MessageIDProperty) CopyMessageID(m2 MessageWithID) {
	m.messageID = m2.MessageID()
}

func (m *MessageIDProperty) SetMessageID(msgID uint16) {
	m.messageID = msgID
}

func (m *MessageIDProperty) MessageID() uint16 {
	return m.messageID
}

// MessageWithID is an interface for all messages which include MessageID property.
type MessageWithID interface {
	MessageID() uint16
	SetMessageID(msgID uint16)
}
