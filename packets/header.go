package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	pktType   PacketType
}

func NewHeader(pktType PacketType, varPartLength uint16) *Header {
	h := &Header{
		pktType: pktType,
	}
	h.SetVarPartLength(varPartLength)
	return h
}

func (h *Header) PacketType() PacketType {
	return h.pktType
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

// Unpack reads a packet header from the given buffer.
func (h *Header) Unpack(buf []byte) error {
	if len(buf) < 2 {
		return fmt.Errorf("bad packet length: expected >=2, got %d", len(buf))
	}

	lengthByte := buf[0]
	if lengthByte == longPacketFlag {
		// Long packet (>255B)
		h.pktLength = binary.BigEndian.Uint16(buf[1:3])
		h.pktType = PacketType(buf[3])
	} else {
		// Short packet (<=255B)
		h.pktLength = uint16(lengthByte)
		h.pktType = PacketType(buf[1])
	}

	return nil
}

func (h *Header) Pack() bytes.Buffer {
	var buff bytes.Buffer

	if h.pktLength > 255 {
		buff.WriteByte(longPacketFlag)
		buff.Write(EncodeUint16(h.pktLength))
	} else {
		buff.WriteByte(byte(h.pktLength))
	}
	buff.WriteByte(byte(h.pktType))

	return buff
}

func ReadByte(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func ReadUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf), nil
}

func EncodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}
