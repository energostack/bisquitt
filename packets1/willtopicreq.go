package packets1

import (
	"io"
)

const willTopicReqVarPartLength uint16 = 0

type WillTopicReq struct {
	Header
}

func NewWillTopicReq() *WillTopicReq {
	return &WillTopicReq{
		Header: *NewHeader(WILLTOPICREQ, willTopicReqVarPartLength),
	}
}

func (m *WillTopicReq) Write(w io.Writer) error {
	buf := m.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillTopicReq) Unpack(r io.Reader) error {
	return nil
}

func (m WillTopicReq) String() string {
	return "WILLTOPICREQ"
}
