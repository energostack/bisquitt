// Package packets1 implements MQTT-SN version 1.2 packets structs.
package packets1

import (
	"errors"
	"fmt"
	"io"

	pkts "github.com/energomonitor/bisquitt/packets"
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

// ReadPacket reads an MQTT-SN packet from the given io.Reader.
// BEWARE: The reader must be a "packet reader" - i.e. it must return one whole
// packet per every Read() call.
func ReadPacket(r io.Reader) (pkt pkts.Packet, err error) {
	var h pkts.Header

	packet := make([]byte, MaxPacketLen)
	n, err := r.Read(packet)
	if err != nil {
		return nil, err
	}
	packet = packet[:n]
	h.Unpack(packet)
	pkt = NewPacketWithHeader(h)
	if pkt == nil {
		return nil, errors.New("invalid MQTT-SN packet")
	}
	pkt.Unpack(packet[h.HeaderLength():])

	return pkt, nil
}

// NewPacketWithHeader returns a particular packet struct with a given header.
// The struct type is determined by h.msgType.
func NewPacketWithHeader(h pkts.Header) (pkt pkts.Packet) {
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
	case pkts.WILLTOPICREQ:
		pkt = &WillTopicReq{Header: h}
	case pkts.WILLTOPIC:
		pkt = &WillTopic{Header: h}
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
	case pkts.WILLTOPICUPD:
		pkt = &WillTopicUpdate{Header: h}
	case pkts.WILLTOPICRESP:
		pkt = &WillTopicResp{Header: h}
	case pkts.WILLMSGUPD:
		pkt = &WillMsgUpd{Header: h}
	case pkts.WILLMSGRESP:
		pkt = &WillMsgResp{Header: h}
	}
	return
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
