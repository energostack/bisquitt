// Package packets1 implements MQTT-SN version 1.2 packets structs.
package packets1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// MQTT-SN specification version 1.2, section 5.2.1 defines maximal packet size
// to be 65535B but pion/udp and pion/dtls use maximal packet size of 8192B.
// See e.g.:
// - https://github.com/pion/udp/blob/b66c29020370bbb21647c27cf0b5ac50a18677f7/conn.go#L17
// - https://github.com/pion/dtls/blob/3dc563b9aede91561ece5ae14b6ec6edf6fc5eb9/conn.go#L30
// An effective MQTT-SN maximal packet size is even a few bytes smaller:
// for UDP transport: PUBLISH with 8183B-long payload = 8192B total packet length
// for DTLS transport: PUBLISH with 8146B-long payload = 8155B total packet length
// (I'm not sure if the DTLS maximal length is affected by the cipher used or not)
//
// The MQTT-SN specification presuppose such packet length limit imposed by the
// network layer:
//
// Note that because MQTT-SN does not support message fragmentation and
// reassembly, the maximum message length that could be used in a network is
// governed by the maximum packet size that is supported by that network, and
// not by the maximum length that could be encoded by MQTT-SN.
// [MQTT-SN specification v. 1.2, chapter 5.2.1 Length]
const MaxPacketLen = 8192

// Because I'm not sure about maximal DTLS header length, we have decided to use
// this arbitrary "small enough to be safe" maximal payload length.
const MaxPayloadLength = 7168

// The header of packets longer than 255B starts with 0x01.
//
// See MQTT-SN specification v. 1.2, chapter 5.2.1 Length.
const longPacketFlag = byte(1)

// Header length in bytes for <=255B and >255B long packets.
//
// See MQTT-SN specification v. 1.2, chapter 5.2.1 Length.
const shortHeaderLength = 2
const longHeaderLength = 4

type Packet interface {
	SetVarPartLength(uint16)
	Write(io.Writer) error
	Unpack(io.Reader) error
	String() string
}

// TopicID type constants.
const (
	TIT_REGISTERED uint8 = iota
	TIT_PREDEFINED
	TIT_SHORT
)

// Whole topic string included in the packet (SUBSCRIBE packet only).
const TIT_STRING = uint8(0)

// Return code constants.
type ReturnCode uint8

const (
	RC_ACCEPTED ReturnCode = iota
	RC_CONGESTION
	RC_INVALID_TOPIC_ID
	RC_NOT_SUPPORTED
)

func (c ReturnCode) String() string {
	switch c {
	case RC_ACCEPTED:
		return "accepted"
	case RC_CONGESTION:
		return "congestion"
	case RC_INVALID_TOPIC_ID:
		return "invalid topic ID"
	case RC_NOT_SUPPORTED:
		return "not supported"
	default:
		return fmt.Sprintf("unknown (%d)", c)
	}
}

// Message ID range.
// We intentionally do not use msgID=0. The MQTT-SN specification does
// not forbid it but uses 0 as an "empty, not used" value.
// I suppose, it's better to not use it to be very explicit about that
// the value really _is_ important if it's non-zero.
const (
	MinMessageID uint16 = 1
	MaxMessageID uint16 = 0xFFFF
)

// Topic ID range.
// The values `0x0000` and `0xFFFF` are reserved and therefore should not be used.
//
// See MQTT-SN specification v. 1.2, chapter 5.3.11.
const (
	MinTopicID uint16 = 1
	MaxTopicID uint16 = 0xFFFF - 1
)

type Header struct {
	// Whole packet length (fixed header + variable part).
	pktLength uint16
	msgType   MessageType
}

func NewHeader(msgType MessageType, varPartLength uint16) *Header {
	h := &Header{
		msgType: msgType,
	}
	h.SetVarPartLength(varPartLength)
	return h
}

// SetVarPartLength sets the length of the packet variable part.
//
// See MQTT-SN specification v. 1.2, chapter 5.2 General Message Format.
func (h *Header) SetVarPartLength(length uint16) {
	if length+shortHeaderLength <= 255 {
		h.pktLength = length + shortHeaderLength
	} else {
		h.pktLength = length + longHeaderLength
	}
}

// VarPartLength returns the length of the packet variable part.
//
// See MQTT-SN specification v. 1.2, chapter 5.2 General Message Format.
func (h *Header) VarPartLength() uint16 {
	return h.pktLength - h.HeaderLength()
}

// PacketLength returns the whole packet length (including header).
//
// See MQTT-SN specification v. 1.2, chapter 5.2 General Message Format.
func (h *Header) PacketLength() uint16 {
	return h.pktLength
}

// HeaderLength returns packet header length.
//
// See MQTT-SN specification v. 1.2, chapter 5.2 General Message Format.
func (h *Header) HeaderLength() uint16 {
	if h.pktLength <= 255 {
		return shortHeaderLength
	} else {
		return longHeaderLength
	}
}

// Unpack reads a packet header from the given io.Reader.
func (h *Header) Unpack(b io.Reader) error {
	lengthByte, err := readByte(b)
	if err != nil {
		return err
	}

	if lengthByte == longPacketFlag {
		// Long packet (>255B)
		if h.pktLength, err = readUint16(b); err != nil {
			return err
		}
	} else {
		// Short packet (<=255B)
		h.pktLength = uint16(lengthByte)
	}

	var msgTypeByte uint8
	msgTypeByte, err = readByte(b)
	h.msgType = MessageType(msgTypeByte)
	return err
}

func (h *Header) pack() bytes.Buffer {
	var buff bytes.Buffer

	if h.pktLength > 255 {
		buff.WriteByte(longPacketFlag)
		buff.Write(encodeUint16(h.pktLength))
	} else {
		buff.WriteByte(byte(h.pktLength))
	}
	buff.WriteByte(byte(h.msgType))

	return buff
}

// ReadPacket reads an MQTT-SN packet from the given io.Reader.
func ReadPacket(r io.Reader) (pkt Packet, err error) {
	var h Header
	packet := make([]byte, MaxPacketLen)
	n, err := r.Read(packet)
	if err != nil {
		return nil, err
	}
	packetBuf := bytes.NewBuffer(packet[:n])
	h.Unpack(packetBuf)
	pkt = NewMessageWithHeader(h)
	if pkt == nil {
		return nil, errors.New("invalid MQTT-SN packet")
	}
	pkt.Unpack(packetBuf)
	return pkt, nil
}

// NewMessageWithHeader returns a particular packet struct with a given header.
// The struct type is determined by h.msgType.
func NewMessageWithHeader(h Header) (m Packet) {
	switch h.msgType {
	case ADVERTISE:
		m = &Advertise{Header: h}
	case SEARCHGW:
		m = &SearchGw{Header: h}
	case GWINFO:
		m = &GwInfo{Header: h}
	case AUTH:
		m = &Auth{Header: h}
	case CONNECT:
		m = &Connect{Header: h}
	case CONNACK:
		m = &Connack{Header: h}
	case WILLTOPICREQ:
		m = &WillTopicReq{Header: h}
	case WILLTOPIC:
		m = &WillTopic{Header: h}
	case WILLMSGREQ:
		m = &WillMsgReq{Header: h}
	case WILLMSG:
		m = &WillMsg{Header: h}
	case REGISTER:
		m = &Register{Header: h}
	case REGACK:
		m = &Regack{Header: h}
	case PUBLISH:
		m = &Publish{Header: h}
	case PUBACK:
		m = &Puback{Header: h}
	case PUBCOMP:
		m = &Pubcomp{Header: h}
	case PUBREC:
		m = &Pubrec{Header: h}
	case PUBREL:
		m = &Pubrel{Header: h}
	case SUBSCRIBE:
		m = &Subscribe{Header: h}
	case SUBACK:
		m = &Suback{Header: h}
	case UNSUBSCRIBE:
		m = &Unsubscribe{Header: h}
	case UNSUBACK:
		m = &Unsuback{Header: h}
	case PINGREQ:
		m = &Pingreq{Header: h}
	case PINGRESP:
		m = &Pingresp{Header: h}
	case DISCONNECT:
		m = &Disconnect{Header: h}
	case WILLTOPICUPD:
		m = &WillTopicUpdate{Header: h}
	case WILLTOPICRESP:
		m = &WillTopicResp{Header: h}
	case WILLMSGUPD:
		m = &WillMsgUpdate{Header: h}
	case WILLMSGRESP:
		m = &WillMsgResp{Header: h}
	}
	return
}

func readByte(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func readUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf), nil
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

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
func DecodeShortTopic(topicID uint16) string {
	return string(encodeUint16(topicID))
}

// Flags bit mask constants.
const (
	flagsTopicIDTypeBits = 0x03
	flagsCleanSessionBit = 0x04
	flagsWillBit         = 0x08
	flagsRetainBit       = 0x10
	flagsQOSBits         = 0x60
	flagsDUPBit          = 0x80
)

// MessageType constants.
type MessageType uint8

const (
	ADVERTISE     MessageType = 0x00
	SEARCHGW      MessageType = 0x01
	GWINFO        MessageType = 0x02
	AUTH          MessageType = 0x03
	CONNECT       MessageType = 0x04
	CONNACK       MessageType = 0x05
	WILLTOPICREQ  MessageType = 0x06
	WILLTOPIC     MessageType = 0x07
	WILLMSGREQ    MessageType = 0x08
	WILLMSG       MessageType = 0x09
	REGISTER      MessageType = 0x0A
	REGACK        MessageType = 0x0B
	PUBLISH       MessageType = 0x0C
	PUBACK        MessageType = 0x0D
	PUBCOMP       MessageType = 0x0E
	PUBREC        MessageType = 0x0F
	PUBREL        MessageType = 0x10
	SUBSCRIBE     MessageType = 0x12
	SUBACK        MessageType = 0x13
	UNSUBSCRIBE   MessageType = 0x14
	UNSUBACK      MessageType = 0x15
	PINGREQ       MessageType = 0x16
	PINGRESP      MessageType = 0x17
	DISCONNECT    MessageType = 0x18
	WILLTOPICUPD  MessageType = 0x1A
	WILLTOPICRESP MessageType = 0x1B
	WILLMSGUPD    MessageType = 0x1C
	WILLMSGRESP   MessageType = 0x1D
	// 0x03 is reserved
	// 0x11 is reserved
	// 0x19 is reserved
	// 0x1E - 0xFD is reserved
	// 0xFE - Encapsulated message
	// 0xFF is reserved
)

func (t MessageType) String() string {
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
