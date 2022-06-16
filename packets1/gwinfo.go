package packets1

import (
	"fmt"
	"io"
)

const gwInfoHeaderLength uint16 = 1

type GwInfoMessage struct {
	Header
	GatewayID      uint8
	GatewayAddress []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewGwInfoMessage(gatewayID uint8, gatewayAddress []byte) *GwInfoMessage {
	m := &GwInfoMessage{
		Header:         *NewHeader(GWINFO, 0),
		GatewayID:      gatewayID,
		GatewayAddress: gatewayAddress,
	}
	m.computeLength()
	return m
}

func (m *GwInfoMessage) computeLength() {
	addrLength := uint16(len(m.GatewayAddress))
	m.Header.SetVarPartLength(gwInfoHeaderLength + addrLength)
}

func (m *GwInfoMessage) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	buf.WriteByte(m.GatewayID)
	buf.Write(m.GatewayAddress)

	_, err := buf.WriteTo(w)
	return err
}

func (m *GwInfoMessage) Unpack(r io.Reader) (err error) {
	if m.GatewayID, err = readByte(r); err != nil {
		return
	}

	if m.VarPartLength() > gwInfoHeaderLength {
		m.GatewayAddress = make([]byte, m.VarPartLength()-gwInfoHeaderLength)
		_, err = io.ReadFull(r, m.GatewayAddress)
	} else {
		m.GatewayAddress = nil
	}
	return
}

func (m GwInfoMessage) String() string {
	return fmt.Sprintf("GWINFO(GatewayID=%d,GatewayAddress=%#v)",
		m.GatewayID, string(m.GatewayAddress))
}
