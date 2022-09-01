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

func (p *WillTopicResp) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = pkts.ReadByte(r)
	p.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (p WillTopicResp) String() string {
	return fmt.Sprintf("WILLTOPICRESP(ReturnCode=%d)", p.ReturnCode)
}
