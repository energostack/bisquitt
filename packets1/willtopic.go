package packets1

import (
	"fmt"
	"io"
)

const willTopicFlagsLength uint16 = 1

type WillTopicMessage struct {
	Header
	QOS       uint8
	Retain    bool
	WillTopic string
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillTopicMessage(willTopic string, qos uint8, retain bool) *WillTopicMessage {
	m := &WillTopicMessage{
		Header:    *NewHeader(WILLTOPIC, 0),
		QOS:       qos,
		Retain:    retain,
		WillTopic: willTopic,
	}
	m.computeLength()
	return m
}

func (m *WillTopicMessage) computeLength() {
	// An empty WILLTOPIC message is a WILLTOPIC message without Flags and
	// WillTopic field (i.e. it is exactly 2 octets long).
	// [MQTT-SN specification v. 1.2, chapter 5.4.7 WILLTOPIC]
	if len(m.WillTopic) == 0 {
		m.Header.SetVarPartLength(0)
	} else {
		length := willTopicFlagsLength + uint16(len(m.WillTopic))
		m.Header.SetVarPartLength(length)
	}
}

func (m *WillTopicMessage) encodeFlags() byte {
	var b byte

	b |= (m.QOS << 5) & flagsQOSBits
	if m.Retain {
		b |= flagsRetainBit
	}
	return b
}

func (m *WillTopicMessage) decodeFlags(b byte) {
	m.QOS = (b & flagsQOSBits) >> 5
	m.Retain = (b & flagsRetainBit) == flagsRetainBit
}

func (m *WillTopicMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	if m.Header.VarPartLength() > 0 {
		buf.WriteByte(m.encodeFlags())
		buf.Write([]byte(m.WillTopic))
	}

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillTopicMessage) Unpack(r io.Reader) (err error) {
	if m.VarPartLength() > 0 {
		var flagsByte uint8
		if flagsByte, err = readByte(r); err != nil {
			return
		}
		m.decodeFlags(flagsByte)

		buff := make([]byte, m.VarPartLength()-willTopicFlagsLength)
		if _, err = io.ReadFull(r, buff); err != nil {
			return
		}
		m.WillTopic = string(buff)
	} else {
		m.WillTopic = ""
	}
	return
}

func (m WillTopicMessage) String() string {
	return fmt.Sprintf("WILLTOPIC(WillTopic=%#v, QOS=%d, Retain=%t)", m.WillTopic, m.QOS, m.Retain)
}
