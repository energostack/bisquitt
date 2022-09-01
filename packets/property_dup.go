package packets

type DUPProperty struct {
	dup bool
}

func NewDUPProperty(dup bool) *DUPProperty {
	return &DUPProperty{
		dup: dup,
	}
}

func (p *DUPProperty) SetDUP(dup bool) {
	p.dup = dup
}

func (p *DUPProperty) DUP() bool {
	return p.dup
}

// PacketWithDUP is an interface for all packets which include DUP property.
type PacketWithDUP interface {
	DUP() bool
	SetDUP(bool)
}
