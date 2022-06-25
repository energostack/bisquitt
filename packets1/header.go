package packets1

import (
	"bytes"
	"io"
)

// The header of packets longer than 255B starts with 0x01.
//
// See MQTT-SN specification v. 1.2, chapter 5.2.1 Length.
const longPacketFlag = byte(1)

// Header length in bytes for <=255B and >255B long packets.
//
// See MQTT-SN specification v. 1.2, chapter 5.2.1 Length.
const shortHeaderLength = 2
const longHeaderLength = 4

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
