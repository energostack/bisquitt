package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willTopicRespVarPartLength uint16 = 1

type WillTopicResp struct {
	pkts.Header
	ReturnCode ReturnCode
}

func NewWillTopicResp(returnCode ReturnCode) *WillTopicResp {
	return &WillTopicResp{
		Header:     *pkts.NewHeader(pkts.WILLTOPICRESP, willTopicRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *WillTopicResp) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
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
