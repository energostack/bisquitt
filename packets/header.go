package packets

import (
	"bytes"
	"encoding/binary"
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

// Unpack reads a packet header from the given io.Reader.
func (h *Header) Unpack(b io.Reader) error {
	lengthByte, err := ReadByte(b)
	if err != nil {
		return err
	}

	if lengthByte == longPacketFlag {
		// Long packet (>255B)
		if h.pktLength, err = ReadUint16(b); err != nil {
			return err
		}
	} else {
		// Short packet (<=255B)
		h.pktLength = uint16(lengthByte)
	}

	var msgTypeByte uint8
	msgTypeByte, err = ReadByte(b)
	h.pktType = PacketType(msgTypeByte)
	return err
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
