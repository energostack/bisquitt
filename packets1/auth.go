package packets1

import (
	"bytes"
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

// Authentication reason constants.
const (
	AUTH_SUCCESS        uint8 = 0
	AUTH_CONTINUE       uint8 = 0x18
	AUTH_REAUTHENTICATE uint8 = 0x19
)

// Auth method constants.
const (
	AUTH_PLAIN = "PLAIN"
)

// Non-standard extension to allow user/password authentication in MQTT-SN.
//
// Implemented according to the OASIS Open MQTT-SN v. 2.0 draft
// https://www.oasis-open.org/committees/download.php/68568/mqtt-sn-v2.0-wd09.docx
//
// SASL PLAIN method specification: https://datatracker.ietf.org/doc/html/rfc4616
type Auth struct {
	pkts.Header
	Reason uint8
	Method string
	Data   []byte
}

// NewAuthPlain creates a new Auth with "PLAIN" method encoded
// authentication data.
func NewAuthPlain(user string, password []byte) *Auth {
	auth := &Auth{Header: *pkts.NewHeader(pkts.AUTH, 0)}
	auth.Method = "PLAIN"
	var b bytes.Buffer
	b.Write([]byte{0})
	b.Write([]byte(user))
	b.Write([]byte{0})
	b.Write(password)
	auth.Data = b.Bytes()
	length := 2 + len(auth.Method) + len(auth.Data)
	auth.SetVarPartLength(uint16(length))
	return auth
}

// DecodePlain decodes username and password from AUTH package data encoded
// using "PLAIN" method.
func DecodePlain(auth *Auth) (string, []byte, error) {
	dataParts := bytes.Split(auth.Data, []byte{0})
	if len(dataParts) != 3 {
		return "", nil, fmt.Errorf("invalid PLAIN auth data format: %v.", auth.Data)
	}
	// NOTE: PLAIN first part (authorization identity) not used.
	return string(dataParts[1]), dataParts[2], nil
}

func (p *Auth) Write(w io.Writer) error {
	buf := p.Header.Pack()
	buf.WriteByte(p.Reason)
	buf.WriteByte(byte(len(p.Method)))
	buf.Write([]byte(p.Method))
	buf.Write([]byte(p.Data))

	_, err := buf.WriteTo(w)
	return err
}

func (p *Auth) Unpack(r io.Reader) (err error) {
	if p.Reason, err = pkts.ReadByte(r); err != nil {
		return
	}

	var methodLen uint8
	if methodLen, err = pkts.ReadByte(r); err != nil {
		return
	}
	method := make([]byte, methodLen)
	if _, err = io.ReadFull(r, method); err != nil {
		return
	}
	p.Method = string(method)

	p.Data = make([]byte, p.VarPartLength()-2-uint16(methodLen))
	_, err = io.ReadFull(r, p.Data)
	return
}

func (p Auth) String() string {
	// We intentionally do not print Data because it contains sensitive data.
	return fmt.Sprintf("AUTH(Reason=%d, Method=%#v)", p.Reason, p.Method)
}
