package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willMsgRespVarPartLength uint16 = 1

type WillMsgResp struct {
	pkts.Header
	PacketV2
	// Fields
	ReasonCode ReasonCode
}

func NewWillMsgResp(reasonCode ReasonCode) *WillMsgResp {
	return &WillMsgResp{
		Header:     *pkts.NewHeader(pkts.WILLMSGRESP, willMsgRespVarPartLength),
		ReasonCode: reasonCode,
	}
}

func (p *WillMsgResp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReasonCode))

	return buf.Bytes(), nil
}

func (p *WillMsgResp) Unpack(buf []byte) error {
	if len(buf) != int(willMsgRespVarPartLength) {
		return fmt.Errorf("bad WILLMSGRESP2 packet length: expected %d, got %d",
			willMsgRespVarPartLength, len(buf))
	}

	p.ReasonCode = ReasonCode(buf[0])

	return nil
}

func (p WillMsgResp) String() string {
	return fmt.Sprintf("WILLMSGRESP2(ReasonCode=%s)", p.ReasonCode)
}
