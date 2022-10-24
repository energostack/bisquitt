package packets2

import (
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicRespVarPartLength uint16 = 1

type WillTopicResp struct {
	pkts.Header
	PacketV2
	// Fields
	ReasonCode ReasonCode
}

func NewWillTopicResp(reasonCode ReasonCode) *WillTopicResp {
	return &WillTopicResp{
		Header:     *pkts.NewHeader(pkts.WILLTOPICRESP, willTopicRespVarPartLength),
		ReasonCode: reasonCode,
	}
}

func (p *WillTopicResp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReasonCode))

	return buf.Bytes(), nil
}

func (p *WillTopicResp) Unpack(buf []byte) error {
	if len(buf) != int(willTopicRespVarPartLength) {
		return fmt.Errorf("bad WILLTOPICRESP2 packet length: expected %d, got %d",
			willTopicRespVarPartLength, len(buf))
	}

	p.ReasonCode = ReasonCode(buf[0])

	return nil
}

func (p WillTopicResp) String() string {
	return fmt.Sprintf("WILLTOPICRESP2(ReasonCode=%s)", p.ReasonCode)
}
