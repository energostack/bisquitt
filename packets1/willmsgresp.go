package packets1

import (
	"fmt"
	"io"
)

const willMsgRespVarPartLength uint16 = 1

type WillMsgResp struct {
	Header
	ReturnCode ReturnCode
}

func NewWillMsgResp(returnCode ReturnCode) *WillMsgResp {
	return &WillMsgResp{
		Header:     *NewHeader(WILLMSGRESP, willMsgRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (p *WillMsgResp) Write(w io.Writer) error {
	buf := p.Header.pack()
	buf.WriteByte(byte(p.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsgResp) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	p.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (p WillMsgResp) String() string {
	return fmt.Sprintf("WILLMSGRESP(ReturnCode=%d)", p.ReturnCode)
}
