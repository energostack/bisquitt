package packets1

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicUpdHeaderLength uint16 = 1

type WillTopicUpd struct {
	pkts.Header
	QOS       uint8
	Retain    bool
	WillTopic []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillTopicUpd(willTopic []byte, qos uint8, retain bool) *WillTopicUpd {
	p := &WillTopicUpd{
		Header:    *pkts.NewHeader(pkts.WILLTOPICUPD, 0),
		QOS:       qos,
		Retain:    retain,
		WillTopic: willTopic,
	}
	p.computeLength()
	return p
}

func (p *WillTopicUpd) computeLength() {
	topicLength := uint16(len(p.WillTopic))
	p.Header.SetVarPartLength(willTopicUpdHeaderLength + topicLength)
}

func (p *WillTopicUpd) encodeFlags() byte {
	var b byte
	b |= (p.QOS << 5) & flagsQOSBits
	if p.Retain {
		b |= flagsRetainBit
	}
	return b
}

func (p *WillTopicUpd) decodeFlags(b byte) {
	p.QOS = (b & flagsQOSBits) >> 5
	p.Retain = (b & flagsRetainBit) == flagsRetainBit
}

func (p *WillTopicUpd) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(p.WillTopic)

	return buf.Bytes(), nil
}

func (p *WillTopicUpd) Unpack(buf []byte) error {
	if len(buf) <= int(willTopicUpdHeaderLength) {
		return fmt.Errorf("bad WILLTOPICUPD packet length: expected >%d, got %d",
			willTopicUpdHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.WillTopic = buf[1:]

	return nil
}

func (p WillTopicUpd) String() string {
	return fmt.Sprintf("WILLTOPICUPD(WillTopic=%#v, QOS=%d, Retain=%t)",
		p.WillTopic, p.QOS, p.Retain)
}
