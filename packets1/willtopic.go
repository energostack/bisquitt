package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicFlagsLength uint16 = 1

type WillTopic struct {
	pkts.Header
	QOS       uint8
	Retain    bool
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

func (p *WillTopic) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	if p.Header.VarPartLength() > 0 {
		buf.WriteByte(p.encodeFlags())
		buf.Write([]byte(p.WillTopic))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillTopic) Unpack(r io.Reader) (err error) {
	if p.VarPartLength() > 0 {
		var flagsByte uint8
		if flagsByte, err = pkts.ReadByte(r); err != nil {
			return
		}
		p.decodeFlags(flagsByte)

		buff := make([]byte, p.VarPartLength()-willTopicFlagsLength)
		if _, err = io.ReadFull(r, buff); err != nil {
			return
		}
		p.WillTopic = string(buff)
	} else {
		p.WillTopic = ""
	}
	return
}

func (p WillTopic) String() string {
	return fmt.Sprintf("WILLTOPIC(WillTopic=%#v, QOS=%d, Retain=%t)", p.WillTopic, p.QOS, p.Retain)
}
