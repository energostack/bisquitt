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

func (m *WillMsgReq) Write(w io.Writer) error {
	buf := m.Header.pack()

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillMsgReq) Unpack(r io.Reader) error {
	return nil
}

func (m WillMsgReq) String() string {
	return "WILLMSGREQ"
}
