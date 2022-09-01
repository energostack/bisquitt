package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const willMsgRespVarPartLength uint16 = 1

type WillMsgResp struct {
	pkts.Header
	ReturnCode ReturnCode
}

func NewWillMsgResp(returnCode ReturnCode) *WillMsgResp {
	return &WillMsgResp{
		Header:     *pkts.NewHeader(pkts.WILLMSGRESP, willMsgRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *WillMsgResp) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsgResp) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = pkts.ReadByte(r)
	p.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (p WillMsgResp) String() string {
	return fmt.Sprintf("WILLMSGRESP(ReturnCode=%d)", p.ReturnCode)
}
