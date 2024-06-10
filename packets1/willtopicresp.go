package packets1

import (
	"fmt"

	pkts "github.com/energostack/bisquitt/packets"
)

const willTopicRespVarPartLength uint16 = 1

type WillTopicResp struct {
	pkts.Header
	// Fields
	ReturnCode ReturnCode
}

func NewWillTopicResp(returnCode ReturnCode) *WillTopicResp {
	return &WillTopicResp{
		Header:     *pkts.NewHeader(pkts.WILLTOPICRESP, willTopicRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *WillTopicResp) Pack() ([]byte, error) {
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReturnCode))

	return buf.Bytes(), nil
}

func (p *WillTopicResp) Unpack(buf []byte) error {
	if len(buf) != int(willTopicRespVarPartLength) {
		return fmt.Errorf("bad WILLTOPICRESP packet length: expected %d, got %d",
			willTopicRespVarPartLength, len(buf))
	}

	p.ReturnCode = ReturnCode(buf[0])

	return nil
}

func (p WillTopicResp) String() string {
	return fmt.Sprintf("WILLTOPICRESP(ReturnCode=%d)", p.ReturnCode)
}
