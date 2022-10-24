package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicUpdFlagsLength uint16 = 1

type WillTopicUpd struct {
	pkts.Header
	PacketV2
	// Flags
	WillQOS uint8
	Retain  bool
	// Fields
	WillTopic string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillTopicUpd(willTopic string, qos uint8, retain bool) *WillTopicUpd {
	p := &WillTopicUpd{
		Header:    *pkts.NewHeader(pkts.WILLTOPICUPD, 0),
		WillQOS:   qos,
		Retain:    retain,
		WillTopic: willTopic,
	}
	p.computeLength()
	return p
}

func (p *WillTopicUpd) computeLength() {
	// An empty WILLTOPICUPD packet is a WILLTOPICUPD packet without Flags
	// and WillTopic field (i.e. it is exactly 2 bytes long).
	// [MQTT-SN 2.0 specification draft WD20, chapter 2.1.25.3 Will Topic]
	if len(p.WillTopic) == 0 {
		p.Header.SetVarPartLength(0)
	} else {
		length := willTopicUpdFlagsLength + uint16(len(p.WillTopic))
		p.Header.SetVarPartLength(length)
	}
}

func (p *WillTopicUpd) encodeFlags() byte {
	var b byte
	b |= (p.WillQOS << 5) & flagsQOSBits
	if p.Retain {
		b |= flagsRetainBit
	}
	return b
}

func (p *WillTopicUpd) decodeFlags(b byte) {
	p.WillQOS = (b & flagsQOSBits) >> 5
	p.Retain = (b & flagsRetainBit) == flagsRetainBit
}

func (p *WillTopicUpd) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	if p.Header.VarPartLength() > 0 {
		_ = buf.WriteByte(p.encodeFlags())
		_, _ = buf.Write([]byte(p.WillTopic))
	}

	return buf.Bytes(), nil
}

func (p *WillTopicUpd) Unpack(buf []byte) error {
	switch len(buf) {
	case 0:
		p.WillTopic = ""
	case 1:
		return fmt.Errorf("bad WILLTOPICUPD2 packet length: expected 0 or >=2, got %d",
			len(buf))
	default:
		p.decodeFlags(buf[0])
		p.WillTopic = string(buf[1:])
	}

	return nil
}

func (p WillTopicUpd) String() string {
	if len(p.WillTopic) == 0 {
		return fmt.Sprintf(`WILLTOPICUPD2(WillTopic="")`)
	}
	return fmt.Sprintf("WILLTOPICUPD2(WillTopic=%#v, QOS=%d, Retain=%t)",
		p.WillTopic, p.WillQOS, p.Retain)
}
