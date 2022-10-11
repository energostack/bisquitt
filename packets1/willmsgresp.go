package packets1

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willMsgRespVarPartLength uint16 = 1

type WillMsgResp struct {
	pkts.Header
	// Fields
	ReturnCode ReturnCode
}

func NewWillMsgResp(returnCode ReturnCode) *WillMsgResp {
	return &WillMsgResp{
		Header:     *pkts.NewHeader(pkts.WILLMSGRESP, willMsgRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *WillMsgResp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReturnCode))

	return buf.Bytes(), nil
}

func (p *WillMsgResp) Unpack(buf []byte) error {
	if len(buf) != int(willMsgRespVarPartLength) {
		return fmt.Errorf("bad WILLMSGRESP packet length: expected %d, got %d",
			willMsgRespVarPartLength, len(buf))
	}

	p.ReturnCode = ReturnCode(buf[0])

	return nil
}

func (p WillMsgResp) String() string {
	return fmt.Sprintf("WILLMSGRESP(ReturnCode=%d)", p.ReturnCode)
}
