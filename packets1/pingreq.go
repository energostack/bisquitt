package packets1

import (
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

type Pingreq struct {
	pkts.Header
	// Fields
	ClientID []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewPingreq(clientID []byte) *Pingreq {
	p := &Pingreq{
		Header:   *pkts.NewHeader(pkts.PINGREQ, 0),
		ClientID: clientID,
	}
	p.computeLength()
	return p
}

func (p *Pingreq) computeLength() {
	length := len(p.ClientID)
	p.Header.SetVarPartLength(uint16(length))
}

func (p *Pingreq) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_, _ = buf.Write(p.ClientID)

	return buf.Bytes(), nil
}

func (p *Pingreq) Unpack(buf []byte) error {
	p.ClientID = buf
	return nil
}

func (p Pingreq) String() string {
	return fmt.Sprintf("PINGREQ(ClientID=%#v)", string(p.ClientID))
}
