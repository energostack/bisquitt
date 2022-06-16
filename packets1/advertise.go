package packets1

import (
	"fmt"
	"io"
)

const advertiseVarPartLength uint16 = 3

type Advertise struct {
	Header
	GatewayID uint8
	Duration  uint16
}

func NewAdvertise(gatewayID uint8, duration uint16) *Advertise {
	return &Advertise{
		Header:    *NewHeader(ADVERTISE, advertiseVarPartLength),
		GatewayID: gatewayID,
		Duration:  duration,
	}
}

func (m *Advertise) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(m.GatewayID)
	buf.Write(encodeUint16(m.Duration))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Advertise) Unpack(r io.Reader) (err error) {
	if m.GatewayID, err = readByte(r); err != nil {
		return
	}

	m.Duration, err = readUint16(r)
	return
}

func (m Advertise) String() string {
	return fmt.Sprintf("ADVERTISE(GatewayID=%d,Duration=%d)",
		m.GatewayID, m.Duration)
}
