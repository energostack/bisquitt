package packets1

import (
	"fmt"
	"io"
)

const willTopicUpdateHeaderLength uint16 = 1

type WillTopicUpdateMessage struct {
	Header
	QOS       uint8
	Retain    bool
	WillTopic []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewWillTopicUpdateMessage(willTopic []byte, qos uint8, retain bool) *WillTopicUpdateMessage {
	m := &WillTopicUpdateMessage{
		Header:    *NewHeader(WILLTOPICUPD, 0),
		QOS:       qos,
		Retain:    retain,
		WillTopic: willTopic,
	}
	m.computeLength()
	return m
}

func (m *WillTopicUpdateMessage) computeLength() {
	topicLength := uint16(len(m.WillTopic))
	m.Header.SetVarPartLength(willTopicUpdateHeaderLength + topicLength)
}

func (m *WillTopicUpdateMessage) encodeFlags() byte {
	var b byte
	b |= (m.QOS << 5) & flagsQOSBits
	if m.Retain {
		b |= flagsRetainBit
	}
	return b
}

func (m *WillTopicUpdateMessage) decodeFlags(b byte) {
	m.QOS = (b & flagsQOSBits) >> 5
	m.Retain = (b & flagsRetainBit) == flagsRetainBit
}

func (m *WillTopicUpdateMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.WriteByte(m.encodeFlags())
	buf.Write(m.WillTopic)

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillTopicUpdateMessage) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = readByte(r); err != nil {
		return
	}
	m.decodeFlags(flagsByte)

	m.WillTopic = make([]byte, m.VarPartLength()-willTopicUpdateHeaderLength)
	_, err = io.ReadFull(r, m.WillTopic)
	return
}

func (m WillTopicUpdateMessage) String() string {
	return fmt.Sprintf("WILLTOPICUPDATE(WillTopic=%#v, QOS=%d, Retain=%t)",
		m.WillTopic, m.QOS, m.Retain)
}
