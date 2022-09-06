package packets1

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicUpdateHeaderLength uint16 = 1

type WillTopicUpdate struct {
	pkts.Header
	QOS       uint8
	Retain    bool
	WillTopic []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillTopicUpdate(willTopic []byte, qos uint8, retain bool) *WillTopicUpdate {
	p := &WillTopicUpdate{
		Header:    *pkts.NewHeader(pkts.WILLTOPICUPD, 0),
		QOS:       qos,
		Retain:    retain,
		WillTopic: willTopic,
	}
	p.computeLength()
	return p
}

func (p *WillTopicUpdate) computeLength() {
	topicLength := uint16(len(p.WillTopic))
	p.Header.SetVarPartLength(willTopicUpdateHeaderLength + topicLength)
}

func (p *WillTopicUpdate) encodeFlags() byte {
	var b byte
	b |= (p.QOS << 5) & flagsQOSBits
	if p.Retain {
		b |= flagsRetainBit
	}
	return b
}

func (p *WillTopicUpdate) decodeFlags(b byte) {
	p.QOS = (b & flagsQOSBits) >> 5
	p.Retain = (b & flagsRetainBit) == flagsRetainBit
}

func (p *WillTopicUpdate) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(p.encodeFlags())
	_, _ = buf.Write(p.WillTopic)

	return buf.Bytes(), nil
}

func (p *WillTopicUpdate) Unpack(buf []byte) error {
	if len(buf) <= int(willTopicUpdateHeaderLength) {
		return fmt.Errorf("bad WILLTOPICUPDATE packet length: expected >%d, got %d",
			willTopicUpdateHeaderLength, len(buf))
	}

	p.decodeFlags(buf[0])
	p.WillTopic = buf[1:]

	return nil
}

func (p WillTopicUpdate) String() string {
	return fmt.Sprintf("WILLTOPICUPDATE(WillTopic=%#v, QOS=%d, Retain=%t)",
		p.WillTopic, p.QOS, p.Retain)
}
