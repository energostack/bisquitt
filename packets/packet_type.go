package packets

import "fmt"

// Packet type constants.
type PacketType uint8

const (
	ADVERTISE     PacketType = 0x00
	SEARCHGW      PacketType = 0x01
	GWINFO        PacketType = 0x02
	AUTH          PacketType = 0x03
	CONNECT       PacketType = 0x04
	CONNECT2      PacketType = 0x04
	CONNACK       PacketType = 0x05
	WILLTOPICREQ  PacketType = 0x06
	WILLTOPIC     PacketType = 0x07
	WILLMSGREQ    PacketType = 0x08
	WILLMSG       PacketType = 0x09
	REGISTER      PacketType = 0x0A
	REGACK        PacketType = 0x0B
	PUBLISH       PacketType = 0x0C
	PUBACK        PacketType = 0x0D
	PUBCOMP       PacketType = 0x0E
	PUBREC        PacketType = 0x0F
	PUBREL        PacketType = 0x10
	SUBSCRIBE     PacketType = 0x12
	SUBACK        PacketType = 0x13
	UNSUBSCRIBE   PacketType = 0x14
	UNSUBACK      PacketType = 0x15
	PINGREQ       PacketType = 0x16
	PINGRESP      PacketType = 0x17
	DISCONNECT    PacketType = 0x18
	WILLTOPICUPD  PacketType = 0x1A
	WILLTOPICRESP PacketType = 0x1B
	WILLMSGUPD    PacketType = 0x1C
	WILLMSGRESP   PacketType = 0x1D
	// 0x03 is reserved
	// 0x11 is reserved
	// 0x19 is reserved
	// 0x1E - 0xFD is reserved
	// 0xFE - Encapsulated message
	// 0xFF is reserved
)

func (t PacketType) String() string {
	switch t {
	case ADVERTISE:
		return "ADVERTISE"
	case SEARCHGW:
		return "SEARCHGW"
	case GWINFO:
		return "GWINFO"
	case AUTH:
		return "AUTH"
	case CONNECT:
		return "CONNECT"
	case CONNACK:
		return "CONNACK"
	case WILLTOPICREQ:
		return "WILLTOPICREQ"
	case WILLTOPIC:
		return "WILLTOPIC"
	case WILLMSGREQ:
		return "WILLMSGREQ"
	case WILLMSG:
		return "WILLMSG"
	case REGISTER:
		return "REGISTER"
	case REGACK:
		return "REGACK"
	case PUBLISH:
		return "PUBLISH"
	case PUBACK:
		return "PUBACK"
	case PUBCOMP:
		return "PUBCOMP"
	case PUBREC:
		return "PUBREC"
	case PUBREL:
		return "PUBREL"
	case SUBSCRIBE:
		return "SUBSCRIBE"
	case SUBACK:
		return "SUBACK"
	case UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case UNSUBACK:
		return "UNSUBACK"
	case PINGREQ:
		return "PINGREQ"
	case PINGRESP:
		return "PINGRESP"
	case DISCONNECT:
		return "DISCONNECT"
	case WILLTOPICUPD:
		return "WILLTOPICUPD"
	case WILLTOPICRESP:
		return "WILLTOPICRESP"
	case WILLMSGUPD:
		return "WILLMSGUPD"
	case WILLMSGRESP:
		return "WILLMSGRESP"
	default:
		return fmt.Sprintf("unknown (%d)", t)
	}
}
