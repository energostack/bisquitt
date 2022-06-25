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

func (p *WillTopicReq) Write(w io.Writer) error {
	buf := p.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillTopicReq) Unpack(r io.Reader) error {
	return nil
}

func (p WillTopicReq) String() string {
	return "WILLTOPICREQ"
}
