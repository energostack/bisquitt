package packets1

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicFlagsLength uint16 = 1

type WillTopic struct {
	pkts.Header
	// Flags
	QOS    uint8
	Retain bool
	// Fields
	WillTopic string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillTopic(willTopic string, qos uint8, retain bool) *WillTopic {
	p := &WillTopic{
		Header:    *pkts.NewHeader(pkts.WILLTOPIC, 0),
		QOS:       qos,
		Retain:    retain,
		WillTopic: willTopic,
	}
	p.computeLength()
	return p
}

func (p *WillTopic) computeLength() {
	// An empty WILLTOPIC message is a WILLTOPIC message without Flags and
	// WillTopic field (i.e. it is exactly 2 octets long).
	// [MQTT-SN specification v. 1.2, chapter 5.4.7 WILLTOPIC]
	if len(p.WillTopic) == 0 {
		p.Header.SetVarPartLength(0)
	} else {
		length := willTopicFlagsLength + uint16(len(p.WillTopic))
		p.Header.SetVarPartLength(length)
	}
}

func (p *WillTopic) encodeFlags() byte {
	var b byte

	b |= (p.QOS << 5) & flagsQOSBits
	if p.Retain {
		b |= flagsRetainBit
	}
	return b
}

func (p *WillTopic) decodeFlags(b byte) {
	p.QOS = (b & flagsQOSBits) >> 5
	p.Retain = (b & flagsRetainBit) == flagsRetainBit
}

func (p *WillTopic) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	if p.Header.VarPartLength() > 0 {
		_ = buf.WriteByte(p.encodeFlags())
		_, _ = buf.Write([]byte(p.WillTopic))
	}

	return buf.Bytes(), nil
}

func (p *WillTopic) Unpack(buf []byte) error {
	switch len(buf) {
	case 0:
		p.WillTopic = ""
	case 1:
		return fmt.Errorf("bad WILLTOPIC packet length: expected 0 or >=2, got %d",
			len(buf))
	default:
		p.decodeFlags(buf[0])
		p.WillTopic = string(buf[1:])
	}

	return nil
}

func (p WillTopic) String() string {
	return fmt.Sprintf("WILLTOPIC(WillTopic=%#v, QOS=%d, Retain=%t)", p.WillTopic, p.QOS, p.Retain)
}
