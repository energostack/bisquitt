package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicReqVarPartLength uint16 = 0

type WillTopicReq struct {
	pkts.Header
	PacketV2
}

func NewWillTopicReq() *WillTopicReq {
	return &WillTopicReq{
		Header: *pkts.NewHeader(pkts.WILLTOPICREQ, willTopicReqVarPartLength),
	}
}

func (p *WillTopicReq) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()
	return buf.Bytes(), nil
}

func (p *WillTopicReq) Unpack(buf []byte) error {
	if len(buf) != int(willTopicReqVarPartLength) {
		return fmt.Errorf("bad WILLTOPICREQ2 packet length: Expected %d, got %d",
			willTopicReqVarPartLength, len(buf))
	}
	return nil
}

func (p WillTopicReq) String() string {
	return "WILLTOPICREQ2"
}
