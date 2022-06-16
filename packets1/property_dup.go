package packets1

type DUPProperty struct {
	dup bool
}

func (m *DUPProperty) SetDUP(dup bool) {
	m.dup = dup
}

func (m *DUPProperty) DUP() bool {
	return m.dup
}

// PacketWithDUP is an interface for all packets which include DUP property.
type PacketWithDUP interface {
	DUP() bool
	SetDUP(bool)
}
