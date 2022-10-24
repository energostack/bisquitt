package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestUnsubscribeConstructor(t *testing.T) {
	assert := assert.New(t)

	topicAlias := uint16(12)
	topicFilter := "test-topic"
	aliasType := TAT_PREDEFINED
	pkt := NewUnsubscribe(topicAlias, topicFilter, aliasType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Unsubscribe", reflect.TypeOf(pkt).String(), "Type should be Unsubscribe")
	assert.Equal(topicAlias, pkt.TopicAlias, "Bad TopicAlias value")
	assert.Equal(topicFilter, pkt.TopicFilter, "Bad TopicFilter value")
	assert.Equal(aliasType, pkt.TopicAliasType, "Bad TopicAliasType value")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
}

func TestUnsubscribeMarshal(t *testing.T) {
	// Normal topic alias.
	pkt1 := NewUnsubscribe(1234, "", TAT_NORMAL)
	pkt1.SetPacketID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))

	// Predefined topic alias.
	pkt1 = NewUnsubscribe(1234, "", TAT_PREDEFINED)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))

	// Short topic.
	pkt1 = NewUnsubscribe(pkts.EncodeShortTopic("ab"), "", TAT_SHORT)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))

	// Long topic.
	pkt1 = NewUnsubscribe(0, "test-topic", TAT_LONG)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Unsubscribe))
}

func TestUnsubscribeUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short - normal Topic Alias.
	buff := bytes.NewBuffer([]byte{
		6,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_NORMAL),       // Flags
		0, 1,                   // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}

	// Packet too short - predefined Topic Alias.
	buff = bytes.NewBuffer([]byte{
		6,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_PREDEFINED),   // Flags
		0, 1,                   // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}

	// Packet too short - short Topic Alias.
	buff = bytes.NewBuffer([]byte{
		6,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_SHORT),        // Flags
		0, 1,                   // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}

	// Packet too short - long Topic.
	buff = bytes.NewBuffer([]byte{
		5,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_LONG),         // Flags
		0, 1,                   // Packet ID
		// Missing Topic Filter
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}

	// Packet too long - normal Topic Alias.
	buff = bytes.NewBuffer([]byte{
		8,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_NORMAL),       // Flags
		0, 1,                   // Packet ID
		0, 2, // Topic Alias
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}

	// Packet long short - predefined Topic Alias.
	buff = bytes.NewBuffer([]byte{
		8,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_PREDEFINED),   // Flags
		0, 1,                   // Packet ID
		0, 2, // Topic Alias
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}

	// Packet too short - short Topic Alias.
	buff = bytes.NewBuffer([]byte{
		8,                      // Length
		byte(pkts.UNSUBSCRIBE), // Packet Type
		byte(TAT_SHORT),        // Flags
		0, 1,                   // Packet ID
		0, 2, // Topic Alias
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad UNSUBSCRIBE2 packet length")
	}
}

func TestUnsubscribeStringer(t *testing.T) {
	assert := assert.New(t)

	// Normal topic alias.
	pkt := NewUnsubscribe(1234, "", TAT_NORMAL)
	pkt.SetPacketID(2345)
	assert.Equal(`UNSUBSCRIBE2(TopicAlias(normal)=1234, PacketID=2345)`,
		pkt.String())

	// Predefined topic alias.
	pkt = NewUnsubscribe(1234, "", TAT_PREDEFINED)
	pkt.SetPacketID(2345)
	assert.Equal(`UNSUBSCRIBE2(TopicAlias(predefined)=1234, PacketID=2345)`,
		pkt.String())

	// Short topic.
	pkt = NewUnsubscribe(pkts.EncodeShortTopic("ab"), "", TAT_SHORT)
	pkt.SetPacketID(2345)
	assert.Equal(`UNSUBSCRIBE2(Topic(short)="ab", PacketID=2345)`,
		pkt.String())

	// Long topic.
	pkt = NewUnsubscribe(0, "test-topic", TAT_LONG)
	pkt.SetPacketID(2345)
	assert.Equal(`UNSUBSCRIBE2(Topic(long)="test-topic", PacketID=2345)`,
		pkt.String())

	// Illegal topic alias type.
	pkt = NewUnsubscribe(1234, "test-topic", TAT_LONG)
	pkt.TopicAliasType = 5
	pkt.SetPacketID(2345)
	assert.Equal(`UNSUBSCRIBE2(Topic="test-topic", TopicAlias=1234, TopicAliasType=illegal(5), PacketID=2345)`,
		pkt.String())
}

func TestUnsubscribeInvalidAliasType(t *testing.T) {
	assert := assert.New(t)

	// Topic Alias Type checked in constructor.
	assert.PanicsWithValue("invalid TopicAliasType value: 5", func() {
		NewUnsubscribe(1234, "test-topic", 5)
	})

	// Topic Alias Type checked in Pack().
	assert.PanicsWithValue("invalid TopicAliasType value: 5", func() {
		pkt := NewUnsubscribe(1234, "test-topic", TAT_NORMAL)
		pkt.TopicAliasType = 5
		pkt.Pack()
	})
}
