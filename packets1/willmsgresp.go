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

func (m *WillMsgResp) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillMsgResp) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m WillMsgResp) String() string {
	return fmt.Sprintf("WILLMSGRESP(ReturnCode=%d)", m.ReturnCode)
}