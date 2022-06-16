package packets1

import (
	"io"
)

const willTopicReqVarPartLength uint16 = 0

type WillTopicReqMessage struct {
	Header
}

func NewWillTopicReqMessage() *WillTopicReqMessage {
	return &WillTopicReqMessage{
		Header: *NewHeader(WILLTOPICREQ, willTopicReqVarPartLength),
	}
}

func (m *WillTopicReqMessage) Write(w io.Writer) error {
	buf := m.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillTopicReqMessage) Unpack(r io.Reader) error {
	return nil
}

func (m WillTopicReqMessage) String() string {
	return "WILLTOPICREQ"
}
