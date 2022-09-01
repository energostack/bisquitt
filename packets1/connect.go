package packets1

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

const connectHeaderLength uint16 = 4

type Connect struct {
	pkts.Header
	CleanSession bool
	ClientID     []byte
	Duration     uint16
	ProtocolID   uint8
	Will         bool
}

// NOTE: Packet length is initialized in this constructor and recomputed in m.Write().
func NewConnect(clientID []byte, cleanSession bool, will bool, duration uint16) *Connect {
	p := &Connect{
		Header:       *pkts.NewHeader(pkts.CONNECT, 0),
		Will:         will,
		CleanSession: cleanSession,
		ProtocolID:   0x01,
		Duration:     duration,
		ClientID:     clientID,
	}
	p.computeLength()
	return p
}

func (p *Connect) computeLength() {
	clientIDLength := uint16(len(p.ClientID))
	p.Header.SetVarPartLength(connectHeaderLength + clientIDLength)
}

func (p *Connect) decodeFlags(b byte) {
	p.Will = (b & flagsWillBit) == flagsWillBit
	p.CleanSession = (b & flagsCleanSessionBit) == flagsCleanSessionBit
}

func (p *Connect) encodeFlags() byte {
	var b byte
	if p.Will {
		b |= flagsWillBit
	}
	if p.CleanSession {
		b |= flagsCleanSessionBit
	}
	return b
}

func (p *Connect) Write(w io.Writer) error {
	p.computeLength()

	buf := p.Header.Pack()
	buf.WriteByte(p.encodeFlags())
	buf.WriteByte(p.ProtocolID)
	buf.Write(pkts.EncodeUint16(p.Duration))
	buf.Write([]byte(p.ClientID))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Connect) Unpack(r io.Reader) (err error) {
	var flagsByte uint8
	if flagsByte, err = pkts.ReadByte(r); err != nil {
		return
	}
	p.decodeFlags(flagsByte)

	if p.ProtocolID, err = pkts.ReadByte(r); err != nil {
		return
	}

	if p.Duration, err = pkts.ReadUint16(r); err != nil {
		return
	}

	p.ClientID = make([]byte, p.VarPartLength()-connectHeaderLength)
	_, err = io.ReadFull(r, p.ClientID)
	return
}

func (p Connect) String() string {
	return fmt.Sprintf(
		"CONNECT(ClientID=%#v, CleanSession=%t, Will=%t, Duration=%d)",

		string(p.ClientID), p.CleanSession, p.Will, p.Duration,
	)
}
