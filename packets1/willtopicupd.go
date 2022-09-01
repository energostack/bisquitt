package packets1

import (
	"fmt"
	"io"

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

func (p *WillTopicUpdate) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.WriteByte(p.encodeFlags())
	buf.Write(p.WillTopic)

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillTopicUpdate) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = pkts.ReadByte(r); err != nil {
		return
	}
	p.decodeFlags(flagsByte)

	p.WillTopic = make([]byte, p.VarPartLength()-willTopicUpdateHeaderLength)
	_, err = io.ReadFull(r, p.WillTopic)
	return
}

func (p WillTopicUpdate) String() string {
	return fmt.Sprintf("WILLTOPICUPDATE(WillTopic=%#v, QOS=%d, Retain=%t)",
		p.WillTopic, p.QOS, p.Retain)
}
