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
