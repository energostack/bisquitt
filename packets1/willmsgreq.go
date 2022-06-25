package packets1

import (
	"io"
)

const willMsgReqVarPartLength uint16 = 0

type WillMsgReq struct {
	Header
}

func NewWillMsgReq() *WillMsgReq {
	return &WillMsgReq{
		Header: *NewHeader(WILLMSGREQ, willMsgReqVarPartLength),
	}
}

func (p *WillMsgReq) Write(w io.Writer) error {
	buf := p.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsgReq) Unpack(r io.Reader) error {
	return nil
}

func (p WillMsgReq) String() string {
	return "WILLMSGREQ"
}
