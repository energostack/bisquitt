package packets1

import (
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicReqVarPartLength uint16 = 0

type WillTopicReq struct {
	pkts.Header
}

func NewWillTopicReq() *WillTopicReq {
	return &WillTopicReq{
		Header: *pkts.NewHeader(pkts.WILLTOPICREQ, willTopicReqVarPartLength),
	}
}

func (p *WillTopicReq) Write(w io.Writer) error {
	buf := p.Header.Pack()

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillTopicReq) Unpack(r io.Reader) error {
	return nil
}

func (p WillTopicReq) String() string {
	return "WILLTOPICREQ"
}
