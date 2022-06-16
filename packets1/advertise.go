package messages

import (
	"fmt"
	"io"
)

const advertiseVarPartLength uint16 = 3

type AdvertiseMessage struct {
	Header
	GatewayID uint8
	Duration  uint16
}

func NewAdvertiseMessage(gatewayID uint8, duration uint16) *AdvertiseMessage {
	return &AdvertiseMessage{
		Header:    *NewHeader(ADVERTISE, advertiseVarPartLength),
		GatewayID: gatewayID,
		Duration:  duration,
	}
}

func (m *AdvertiseMessage) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(m.GatewayID)
	buf.Write(encodeUint16(m.Duration))

	_, err := buf.WriteTo(w)
	return err
}

func (m *AdvertiseMessage) Unpack(r io.Reader) (err error) {
	if m.GatewayID, err = readByte(r); err != nil {
		return
	}

	m.Duration, err = readUint16(r)
	return
}

func (m AdvertiseMessage) String() string {
	return fmt.Sprintf("ADVERTISE(GatewayID=%d,Duration=%d)",
		m.GatewayID, m.Duration)
}
