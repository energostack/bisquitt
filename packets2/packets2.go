// Package packets2 implements MQTT-SN version 2.0 packets structs.
//
// Implemented according to the WD20 draft of the specification:
// https://www.oasis-open.org/committees/download.php/70377/mqtt-sn-v2.0-wd20.docx
package packets2

import (
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
)

// MQTT-SN specification defines maximal packet size
// to be 65535B but pion/udp and pion/dtls use maximal packet size of 8192B.
// See e.g.:
// - https://github.com/pion/udp/blob/b66c29020370bbb21647c27cf0b5ac50a18677f7/conn.go#L17
// - https://github.com/pion/dtls/blob/3dc563b9aede91561ece5ae14b6ec6edf6fc5eb9/conn.go#L30
// An effective MQTT-SN maximal packet size is even a few bytes smaller:
// for UDP transport: PUBLISH with 8183B-long payload = 8192B total packet length
// for DTLS transport: PUBLISH with 8146B-long payload = 8155B total packet length
// (I'm not sure if the DTLS maximal length is affected by the cipher used or not)
const MaxPacketLen = 8192

type PacketV2 struct{}

func (p PacketV2) MqttSnVersion() uint8 {
	return 2
}

// Reason code constants.
type ReasonCode uint8

const (
	RC_SUCCESS                      ReasonCode = 0x00
	RC_CONGESTION                   ReasonCode = 0x01
	RC_INVALID_TOPIC_ALIAS          ReasonCode = 0x02
	RC_NOT_SUPPORTED                ReasonCode = 0x03
	RC_NO_SESSION                   ReasonCode = 0x05
	RC_UNSUPPORTED_PROTOCOL_VERSION ReasonCode = 0x84
	RC_NOT_AUTHORIZED               ReasonCode = 0x87
	RC_BAD_AUTHENTICATION_METHOD    ReasonCode = 0x8C
	RC_PACKET_TOO_LARGE             ReasonCode = 0x95
	RC_PAYLOAD_FORMAT_INVALID       ReasonCode = 0x99
)

func (rc ReasonCode) String() string {
	switch rc {
	case RC_SUCCESS:
		return "success"
	case RC_CONGESTION:
		return "congestion"
	case RC_INVALID_TOPIC_ALIAS:
		return "invalid topic alias"
	case RC_NOT_SUPPORTED:
		return "not supported"
	case RC_NO_SESSION:
		return "no session"
	case RC_UNSUPPORTED_PROTOCOL_VERSION:
		return "unsupported protocol version"
	case RC_NOT_AUTHORIZED:
		return "not authorized"
	case RC_BAD_AUTHENTICATION_METHOD:
		return "bad authentication error"
	case RC_PACKET_TOO_LARGE:
		return "packet too large"
	case RC_PAYLOAD_FORMAT_INVALID:
		return "payload format invalid"
	default:
		return fmt.Sprintf("invalid (%d)", rc)
	}
}

// Topic Alias Type constants.
type TopicAliasType uint8

const (
	TAT_NORMAL TopicAliasType = iota
	TAT_PREDEFINED
	TAT_SHORT
	TAT_LONG
)

func (tat TopicAliasType) String() string {
	switch tat {
	case TAT_NORMAL:
		return "normal"
	case TAT_PREDEFINED:
		return "predefined"
	case TAT_SHORT:
		return "short"
	case TAT_LONG:
		return "long"
	default:
		return fmt.Sprintf("illegal(%d)", tat)
	}
}

// Flags bits masks.
const flagsNoLocalBit = 0b10000000
const flagsDUPBit = 0b10000000
const flagsQOSBits = 0b01100000
const flagsRetainAsPublishedBit = 0b00010000
const flagsRetainBit = 0b00010000
const flagsRetainHandlingBits = 0b00001100
const flagsTopicAliasTypeBits = 0b00000011

// NewPacketWithHeader returns a particular message struct with a given header.
// The struct type is determined by h.msgType.
func NewPacketWithHeader(h pkts.Header) (pkt pkts.Packet, err error) {
	switch h.PacketType() {
	case pkts.ADVERTISE:
		pkt = &Advertise{Header: h}
	case pkts.SEARCHGW:
		pkt = &SearchGw{Header: h}
	case pkts.GWINFO:
		pkt = &GwInfo{Header: h}
	case pkts.AUTH:
		pkt = &Auth{Header: h}
	case pkts.CONNECT:
		pkt = &Connect{Header: h}
	case pkts.CONNACK:
		pkt = &Connack{Header: h}
	//case pkts.WILLTOPICREQ:
	//	pkt = &WillTopicReq{Header: h}
	//case pkts.WILLTOPIC:
	//	pkt = &WillTopic{Header: h}
	case pkts.WILLMSGREQ:
		pkt = &WillMsgReq{Header: h}
	case pkts.WILLMSG:
		pkt = &WillMsg{Header: h}
	case pkts.REGISTER:
		pkt = &Register{Header: h}
	case pkts.REGACK:
		pkt = &Regack{Header: h}
	case pkts.PUBLISH:
		pkt = &Publish{Header: h}
	case pkts.PUBACK:
		pkt = &Puback{Header: h}
	case pkts.PUBCOMP:
		pkt = &Pubcomp{Header: h}
	case pkts.PUBREC:
		pkt = &Pubrec{Header: h}
	case pkts.PUBREL:
		pkt = &Pubrel{Header: h}
	case pkts.SUBSCRIBE:
		pkt = &Subscribe{Header: h}
	case pkts.SUBACK:
		pkt = &Suback{Header: h}
	case pkts.UNSUBSCRIBE:
		pkt = &Unsubscribe{Header: h}
	case pkts.UNSUBACK:
		pkt = &Unsuback{Header: h}
	case pkts.PINGREQ:
		pkt = &Pingreq{Header: h}
	case pkts.PINGRESP:
		pkt = &Pingresp{Header: h}
	case pkts.DISCONNECT:
		pkt = &Disconnect{Header: h}
	//case pkts.WILLTOPICUPD:
	//	pkt = &WillTopicUpd{Header: h}
	//case pkts.WILLTOPICRESP:
	//	pkt = &WillTopicResp{Header: h}
	//case pkts.WILLMSGUPD:
	//	pkt = &WillMsgUpd{Header: h}
	//case pkts.WILLMSGRESP:
	//	pkt = &WillMsgResp{Header: h}
	default:
		err = fmt.Errorf("invalid MQTT-SN 2.0 packet type: %d", h.PacketType())
	}
	return
}

// ReadPacket reads an MQTT-SN packet from the given io.Reader.
// BEWARE: The reader must be a "packet reader" - i.e. it must return one whole
// packet per every Read() call.
func ReadPacket(r io.Reader) (pkt pkts.Packet, err error) {
	rawPacket := make([]byte, MaxPacketLen)
	n, err := r.Read(rawPacket)
	if err != nil {
		return nil, err
	}
	rawPacket = rawPacket[:n]

	var h pkts.Header
	if err := h.Unpack(rawPacket); err != nil {
		return nil, err
	}
	pkt, err = NewPacketWithHeader(h)
	if err != nil {
		return nil, err
	}
	if err := pkt.Unpack(rawPacket[h.HeaderLength():]); err != nil {
		return nil, err
	}

	return pkt, nil
}
