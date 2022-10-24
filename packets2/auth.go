package packets2

import (
	"bytes"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
)

// Auth method constants.
const (
	AUTH_PLAIN = "PLAIN"
)

const authHeaderLength uint16 = 2

type Auth struct {
	pkts.Header
	PacketV2
	// Fields
	ReasonCode ReasonCode
	Method     string
	Data       []byte
}

// NewAuthPlain creates a new Auth with "PLAIN" method encoded
// authentication data.
// SASL PLAIN method specification: https://datatracker.ietf.org/doc/html/rfc4616
func NewAuthPlain(user string, password []byte) *Auth {
	p := &Auth{Header: *pkts.NewHeader(pkts.AUTH, 0)}
	p.Method = "PLAIN"
	var b bytes.Buffer
	_ = b.WriteByte(0)
	_, _ = b.Write([]byte(user))
	_ = b.WriteByte(0)
	_, _ = b.Write(password)
	p.Data = b.Bytes()
	p.computeLength()
	return p
}

// DecodePlain decodes username and password from AUTH package data encoded
// using "PLAIN" method.
func (p *Auth) DecodePlain() (string, []byte, error) {
	dataParts := bytes.Split(p.Data, []byte{0})
	if len(dataParts) != 3 {
		return "", nil, fmt.Errorf("invalid PLAIN auth data format: %v.", p.Data)
	}
	// NOTE: PLAIN first part (authorization identity) not used.
	return string(dataParts[1]), dataParts[2], nil
}

func (p *Auth) Pack() ([]byte, error) {
	p.computeLength()
	buf := p.Header.PackToBuffer()

	_ = buf.WriteByte(byte(p.ReasonCode))
	_ = buf.WriteByte(byte(len(p.Method)))
	_, _ = buf.Write([]byte(p.Method))
	_, _ = buf.Write([]byte(p.Data))

	return buf.Bytes(), nil
}

func (p *Auth) Unpack(buf []byte) error {
	if len(buf) < int(authHeaderLength) {
		return fmt.Errorf("bad AUTH2 packet length: expected >=%d, got %d",
			authHeaderLength, len(buf))
	}

	p.ReasonCode = ReasonCode(buf[0])
	methodLen := uint16(buf[1])

	if len(buf) < int(authHeaderLength+methodLen) {
		return fmt.Errorf("bad AUTH2 packet length: expected >=%d, got %d",
			authHeaderLength+methodLen, len(buf))
	}

	p.Method = string(buf[authHeaderLength : authHeaderLength+methodLen])
	p.Data = buf[authHeaderLength+methodLen:]

	return nil
}

func (p Auth) String() string {
	// We intentionally do not print Data because it contains sensitive data.
	return fmt.Sprintf("AUTH2(ReasonCode=%s, Method=%#v)", p.ReasonCode, p.Method)
}

func (p *Auth) computeLength() {
	methodLen := uint16(len(p.Method))
	dataLen := uint16(len(p.Data))
	p.Header.SetVarPartLength(authHeaderLength + methodLen + dataLen)
}
