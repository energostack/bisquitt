package packets

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Packet interface {
	fmt.Stringer

	Write(io.Writer) error
	Unpack([]byte) error
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

// Topic Alias range.
// The values `0x0000` and `0xFFFF` are reserved and therefore should not be used.
//
// See MQTT-SN specification v. 1.2, chapter 5.3.11.
const (
	MinTopicAlias uint16 = 1
	MaxTopicAlias uint16 = 0xFFFF - 1
)

// IsShortTopic determines if the given topic is a short topic.
//
// See MQTT-SN specification v. 1.2, chapter 3 MQTT-SN vs MQTT.
func IsShortTopic(topic string) bool {
	return len(topic) == 2
}

// EncodeShortTopic encodes a short string topic into TopicID (uint16).
//
// See MQTT-SN specification v. 1.2, chapter 3 MQTT-SN vs MQTT.
func EncodeShortTopic(topic string) uint16 {
	var result uint16

	bytes := []byte(topic)
	if len(bytes) > 0 {
		result |= (uint16(bytes[0]) << 8)
	}
	if len(bytes) > 1 {
		result |= uint16(bytes[1])
	}

	return result
}

// DecodeShortTopic decodes a short string topic from TopicID (uint16).
//
// See MQTT-SN specification v. 1.2, chapter 3 MQTT-SN vs MQTT.
func DecodeShortTopic(topicAlias uint16) string {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, topicAlias)
	return string(bytes)
}
