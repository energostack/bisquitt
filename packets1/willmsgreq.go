package packets1

import (
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
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

func (p *WillMsgReq) Write(w io.Writer) error {
	buf := p.Header.Pack()

	_, err := buf.WriteTo(w)
	return err
}

func (p *WillMsgReq) Unpack(r io.Reader) error {
	return nil
}

func (p WillMsgReq) String() string {
	return "WILLMSGREQ"
}
