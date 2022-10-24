package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

type Pingreq struct {
	pkts.Header
	PacketV2
	// Fields
	MaxMessages uint8
	ClientID    string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewPingreq(maxMessages uint8, clientID string) *Pingreq {
	p := &Pingreq{
		Header:      *pkts.NewHeader(pkts.PINGREQ, 0),
		MaxMessages: maxMessages,
		ClientID:    clientID,
	}
	p.computeLength()
	return p
}

func (p *Pingreq) computeLength() {
	length := len(p.ClientID)
	p.Header.SetVarPartLength(uint16(length))
}

// NOTE: The MQTT-SN v. 2.0 specification draft WD20 states that both
// MaxMessages and ClientID are optional (chapter 2.1.22 PINGREQ)
// but there is no means to signalize which one is or is not
// present.
// The specification also says that these properties are used by a sleeping
// client (2.1.22.2 and 2.1.22.3).
// Even when the specification does not state this explicitly, I deduce that
// there are 2 possible uses of the PINGREQ packet:
// 1. an "I am alive!" packet without MaxMessages nor ClientID
// 2. an "I am going from asleep to awake" packet containing both MaxMessages
//    and ClientID
// In Bisquitt, the former use case is signalized by an empty ClientID.
func (p *Pingreq) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	if len(p.ClientID) > 0 {
		_ = buf.WriteByte(p.MaxMessages)
		_, _ = buf.Write([]byte(p.ClientID))
	}

	return buf.Bytes(), nil
}

// NOTE: Both MaxMessages and ClientID must be present or missing, see NOTE for Pack().
func (p *Pingreq) Unpack(buf []byte) error {
	if len(buf) == 1 {
		return fmt.Errorf("bad PINGREQ2 packet length: expected 0 or >=2, got %d",
			len(buf))
	}

	if len(buf) >= 2 {
		p.MaxMessages = buf[0]
		p.ClientID = string(buf[1:])
	}

	return nil
}

func (p Pingreq) String() string {
	if len(p.ClientID) == 0 {
		return "PINGREQ2"
	}
	return fmt.Sprintf("PINGREQ2(MaxMessages=%d, ClientID=%q)",
		p.MaxMessages, p.ClientID)
}
