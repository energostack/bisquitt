package packets1

import (
	"fmt"
	"io"
)

const willTopicRespVarPartLength uint16 = 1

type WillTopicResp struct {
	Header
	ReturnCode ReturnCode
}

func NewWillTopicResp(returnCode ReturnCode) *WillTopicResp {
	return &WillTopicResp{
		Header:     *NewHeader(WILLTOPICRESP, willTopicRespVarPartLength),
		ReturnCode: returnCode,
	}
}

func (m *WillTopicResp) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(byte(m.ReturnCode))

	_, err := buf.WriteTo(w)
	return err
}

func (m *WillTopicResp) Unpack(r io.Reader) (err error) {
	var returnCodeByte uint8
	returnCodeByte, err = readByte(r)
	m.ReturnCode = ReturnCode(returnCodeByte)
	return
}

func (m WillTopicResp) String() string {
	return fmt.Sprintf("WILLTOPICRESP(ReturnCode=%d)", m.ReturnCode)
}
