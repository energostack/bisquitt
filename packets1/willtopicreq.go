package packets1

import (
	"fmt"
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

func (p *WillTopicReq) Unpack(buf []byte) error {
	if len(buf) != int(willTopicReqVarPartLength) {
		return fmt.Errorf("bad WILLTOPICREQ packet length: Expected %d, got %d",
			willTopicReqVarPartLength, len(buf))
	}
	return nil
}

func (p WillTopicReq) String() string {
	return "WILLTOPICREQ"
}
