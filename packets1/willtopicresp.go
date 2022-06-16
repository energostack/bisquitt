package packets1

import (
	"fmt"
	"io"
)

const willTopicRespVarPartLength uint16 = 1

type WillTopicRespMessage struct {
	Header
	ReturnCode ReturnCode
}

func NewWillTopicRespMessage(returnCode ReturnCode) *WillTopicRespMessage {
	return &WillTopicRespMessage{
		Header:     *NewHeader(WILLTOPICRESP, willTopicRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (m *WillTopicRespMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillTopicRespMessage) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m WillTopicRespMessage) String() string {
	return fmt.Sprintf("WILLTOPICRESP(ReturnCode=%d)", m.ReturnCode)
}
