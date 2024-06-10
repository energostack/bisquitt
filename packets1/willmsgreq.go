package packets1

import (
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

const willMsgReqVarPartLength uint16 = 0

type WillMsgReq struct {
	pkts.Header
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
		return fmt.Errorf("bad WILLMSGREQ packet length: expected %d, got %d",
			willMsgReqVarPartLength, len(buf))
	}

	return nil
}

func (p WillMsgReq) String() string {
	return "WILLMSGREQ"
}
