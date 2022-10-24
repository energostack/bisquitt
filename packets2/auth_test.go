package packets2

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestAuthPlainConstructor(t *testing.T) {
	assert := assert.New(t)

	user := "test-user"
	password := []byte("test-password")
	pkt := NewAuthPlain(user, password)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Auth", reflect.TypeOf(pkt).String(), "Type should be Auth")
	assert.Equal(RC_SUCCESS, pkt.ReasonCode, "Bad ReasonCode value")
	assert.Equal("PLAIN", pkt.Method, "Bad Method value")
	assert.Equal(
		[]byte(fmt.Sprintf("\x00%s\x00%s", user, password)),
		pkt.Data, "Bad Data value")
}

func TestAuthPlainMarshal(t *testing.T) {
	pkt1 := NewAuthPlain("test-user", []byte("test-password"))
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Auth))
}

func TestAuthUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short - Method Length missing.
	buff := bytes.NewBuffer([]byte{
		3,               // Length
		byte(pkts.AUTH), // PktType
		0,               // Reason Code
		// Method Length missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad AUTH2 packet length")
	}

	// Packet too short - Method missing.
	buff = bytes.NewBuffer([]byte{
		4,               // Length
		byte(pkts.AUTH), // PktType
		0,               // Reason Code
		1,               // Method Length
		// Method missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad AUTH2 packet length")
	}
}

func TestAuthDecodePlain(t *testing.T) {
	assert := assert.New(t)

	user := "test-user"
	password := []byte("test-password")

	// Correct PLAIN data.
	buff := bytes.NewBuffer([]byte{
		11 + byte(len(user)+len(password)), // Length
		byte(pkts.AUTH),                    // PktType
		0,                                  // Reason
		5,                                  // Method Length
		'P', 'L', 'A', 'I', 'N',            // Method
	})
	// Data
	buff.WriteByte(0)
	buff.Write([]byte(user))
	buff.WriteByte(0)
	buff.Write(password)

	pkt, err := ReadPacket(buff)
	assert.Nil(err)

	user2, password2, err := pkt.(*Auth).DecodePlain()
	assert.Nil(err)
	assert.Equal(user, user2)
	assert.Equal(password, password2)

	// Invalid PLAIN data.
	buff = bytes.NewBuffer([]byte{
		11,                      // Length
		byte(pkts.AUTH),         // PktType
		0,                       // Reason
		5,                       // Method Length
		'P', 'L', 'A', 'I', 'N', // Method
		0, 1, // Data (do not contain two zero bytes)
	})
	pkt, err = ReadPacket(buff)
	assert.Nil(err)

	_, _, err = pkt.(*Auth).DecodePlain()
	if assert.Error(err) {
		assert.Contains(err.Error(), "invalid PLAIN auth data format")
	}
}

func TestAuthStringer(t *testing.T) {
	pkt := NewAuthPlain("test-user", []byte("test-password"))
	assert.Equal(t, "AUTH2(ReasonCode=success, Method=\"PLAIN\")", pkt.String())
}
