package packets2

import (
	"bytes"
	"testing"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func testPacketMarshal(t *testing.T, pkt1 pkts.Packet) pkts.Packet {
	buf, err := pkt1.Pack()
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf)
	pkt2, err := ReadPacket(r)
	if err != nil {
		t.Fatal(err)
	}

	return pkt2
}
