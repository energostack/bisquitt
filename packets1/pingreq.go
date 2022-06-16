package packets1

import (
	"fmt"
	"io"
)

type Pingreq struct {
	Header
	ClientID []byte
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewPingreq(clientID []byte) *Pingreq {
	m := &Pingreq{
		Header:   *NewHeader(PINGREQ, 0),
		ClientID: clientID,
	}
	m.computeLength()
	return m
}

func (m *Pingreq) computeLength() {
	length := len(m.ClientID)
	m.Header.SetVarPartLength(uint16(length))
}

func (m *Pingreq) Write(w io.Writer) error {
	m.computeLength()

	buf := m.Header.pack()
	if len(m.ClientID) > 0 {
		buf.Write(m.ClientID)
	}

	_, err := buf.WriteTo(w)
	return err
}

func (m *Pingreq) Unpack(r io.Reader) (err error) {
	if m.VarPartLength() > 0 {
		m.ClientID = make([]byte, m.VarPartLength())
		_, err = io.ReadFull(r, m.ClientID)
	} else {
		m.ClientID = nil
	}
	return
}

func (m Pingreq) String() string {
	return fmt.Sprintf("PINGREQ(ClientID=%#v)", string(m.ClientID))
}
