package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willMsgReqVarPartLength uint16 = 0

type WillMsgReq struct {
	pkts.Header
	PacketV2
}

func NewWillMsgReq() *WillMsgReq {
	return &WillMsgReq{
		Header: *pkts.NewHeader(pkts.WILLMSGREQ, willMsgReqVarPartLength),
	}
}

func (p *WillMsgReq) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()
	return buf.Bytes(), nil
}

func (p *WillMsgReq) Unpack(buf []byte) error {
	if len(buf) != int(willMsgReqVarPartLength) {
		return fmt.Errorf("bad WILLMSGREQ2 packet length: expected %d, got %d",
			willMsgReqVarPartLength, len(buf))
	}

	return nil
}

func (p WillMsgReq) String() string {
	return "WILLMSGREQ2"
}
