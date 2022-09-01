package packets

import (
	"fmt"
	"io"
)

type Packet interface {
	fmt.Stringer

	Write(io.Writer) error
	Unpack(io.Reader) error

	// NOTE: This method is not used anywhere but we must temporary keep it
	// here because if we remove it, MQTT packets would accidently implement
	// this interface and we would not be able to type-switch between MQTT
	// and MQTT-SN packets.
	// TODO: Remove as soon as this interface is changed and this problem
	//  disappears.
	SetVarPartLength(uint16)
}

// Packet ID range.
// We intentionally do not use pktID=0. The MQTT-SN specification does
// not forbid it but uses 0 as an "empty, not used" value.
// I suppose, it's better to not use it to be very explicit about that
// the value really _is_ important if it's non-zero.
const (
	MinPacketID uint16 = 1
	MaxPacketID uint16 = 0xFFFF
)
