package messages

type DUPProperty struct {
	dup bool
}

func (m *DUPProperty) SetDUP(dup bool) {
	m.dup = dup
}

func (m *DUPProperty) DUP() bool {
	return m.dup
}

// MessageWithDUP is an interface for all messages which include DUP property.
type MessageWithDUP interface {
	DUP() bool
	SetDUP(bool)
}
