package packets1

import (
	"bytes"
	"fmt"
	"io"
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
	Header
	Reason uint8
	Method string
	Data   []byte
}

// NewAuthPlain creates a new Auth with "PLAIN" method encoded
// authentication data.
func NewAuthPlain(user string, password []byte) *Auth {
	auth := &Auth{Header: *NewHeader(AUTH, 0)}
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

// DecodePlain decodes username and password from AUTH message data encoded
// using "PLAIN" method.
func DecodePlain(auth *Auth) (string, []byte, error) {
	dataParts := bytes.Split(auth.Data, []byte{0})
	if len(dataParts) != 3 {
		return "", nil, fmt.Errorf("Invalid PLAIN auth data format: %v.", auth.Data)
	}
	// NOTE: PLAIN first part (authorization identity) not used.
	return string(dataParts[1]), dataParts[2], nil
}

func (m *Auth) Write(w io.Writer) error {
	buf := m.Header.pack()
	buf.WriteByte(m.Reason)
	buf.WriteByte(byte(len(m.Method)))
	buf.Write([]byte(m.Method))
	buf.Write([]byte(m.Data))

	_, err := buf.WriteTo(w)
	return err
}

func (m *Auth) Unpack(r io.Reader) (err error) {
	if m.Reason, err = readByte(r); err != nil {
		return
	}

	var methodLen uint8
	if methodLen, err = readByte(r); err != nil {
		return
	}
	method := make([]byte, methodLen)
	if _, err = io.ReadFull(r, method); err != nil {
		return
	}
	m.Method = string(method)

	m.Data = make([]byte, m.VarPartLength()-2-uint16(methodLen))
	_, err = io.ReadFull(r, m.Data)
	return
}

func (m Auth) String() string {
	// We intentionally do not print Data because it contains sensitive data.
	return fmt.Sprintf("AUTH(Reason=%d, Method=%#v)", m.Reason, m.Method)
}
