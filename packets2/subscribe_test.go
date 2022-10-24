package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestSubscribeConstructor(t *testing.T) {
	assert := assert.New(t)

	topicAlias := uint16(12)
	topicFilter := "test-topic"
	noLocal := true
	qos := uint8(1)
	retainAsPublished := true
	retainHandling := RH_SEND_AT_NEW_SUBSCRIBE
	aliasType := TAT_PREDEFINED
	pkt := NewSubscribe(topicAlias, topicFilter, noLocal, qos,
		retainAsPublished, retainHandling, aliasType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Subscribe", reflect.TypeOf(pkt).String(), "Type should be Subscribe")
	assert.Equal(topicAlias, pkt.TopicAlias, "Bad TopicAlias value")
	assert.Equal(topicFilter, pkt.TopicFilter, "Bad TopicFilter value")
	assert.Equal(noLocal, pkt.NoLocal, "Bad NoLocal value")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(retainAsPublished, pkt.RetainAsPublished, "Bad RetainAsPublished value")
	assert.Equal(retainHandling, pkt.RetainHandling, "Bad RetainHandling value")
	assert.Equal(aliasType, pkt.TopicAliasType, "Bad TopicAliasType value")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
}

func TestSubscribeMarshal(t *testing.T) {
	// Normal topic alias.
	pkt1 := NewSubscribe(1234, "", true, 1, true, RH_SEND_AT_NEW_SUBSCRIBE,
		TAT_NORMAL)
	pkt1.SetPacketID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))

	// Predefined topic alias.
	pkt1 = NewSubscribe(1234, "", true, 1, true, RH_SEND_AT_NEW_SUBSCRIBE,
		TAT_PREDEFINED)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))

	// Short topic.
	pkt1 = NewSubscribe(pkts.EncodeShortTopic("ab"), "", true, 1, true,
		RH_SEND_AT_NEW_SUBSCRIBE, TAT_SHORT)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))

	// Long topic.
	pkt1 = NewSubscribe(0, "test-topic", true, 1, true,
		RH_SEND_AT_NEW_SUBSCRIBE, TAT_LONG)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Subscribe))
}

func TestSubscribeUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short - normal Topic Alias.
	buff := bytes.NewBuffer([]byte{
		6,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_NORMAL),     // Flags
		0, 1,                 // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}

	// Packet too short - predefined Topic Alias.
	buff = bytes.NewBuffer([]byte{
		6,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_PREDEFINED), // Flags
		0, 1,                 // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}

	// Packet too short - short Topic Alias.
	buff = bytes.NewBuffer([]byte{
		6,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_SHORT),      // Flags
		0, 1,                 // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}

	// Packet too short - long Topic.
	buff = bytes.NewBuffer([]byte{
		5,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_LONG),       // Flags
		0, 1,                 // Packet ID
		// Missing Topic Filter
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}

	// Packet too long - normal Topic Alias.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_NORMAL),     // Flags
		0, 1,                 // Packet ID
		0, 2, // Topic Alias
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}

	// Packet long short - predefined Topic Alias.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_PREDEFINED), // Flags
		0, 1,                 // Packet ID
		0, 2, // Topic Alias
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}

	// Packet too short - short Topic Alias.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.SUBSCRIBE), // Packet Type
		byte(TAT_SHORT),      // Flags
		0, 1,                 // Packet ID
		0, 2, // Topic Alias
		0, // junk
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad SUBSCRIBE2 packet length")
	}
}

func TestSubscribeStringer(t *testing.T) {
	assert := assert.New(t)

	// Normal topic alias.
	pkt := NewSubscribe(1234, "", true, 1, true, RH_SEND_AT_NEW_SUBSCRIBE,
		TAT_NORMAL)
	pkt.SetPacketID(2345)
	assert.Equal(`SUBSCRIBE2(TopicAlias(normal)=1234, QOS=1, PacketID=2345, NoLocal=true, RAP=true, RH=send_at_new_subscribe)`,
		pkt.String())

	// Predefined topic alias.
	pkt = NewSubscribe(1234, "", true, 1, true, RH_SEND_AT_NEW_SUBSCRIBE,
		TAT_PREDEFINED)
	pkt.SetPacketID(2345)
	assert.Equal(`SUBSCRIBE2(TopicAlias(predefined)=1234, QOS=1, PacketID=2345, NoLocal=true, RAP=true, RH=send_at_new_subscribe)`,
		pkt.String())

	// Short topic.
	pkt = NewSubscribe(pkts.EncodeShortTopic("ab"), "", true, 1, true,
		RH_SEND_AT_NEW_SUBSCRIBE, TAT_SHORT)
	pkt.SetPacketID(2345)
	assert.Equal(`SUBSCRIBE2(Topic(short)="ab", QOS=1, PacketID=2345, NoLocal=true, RAP=true, RH=send_at_new_subscribe)`,
		pkt.String())

	// Long topic.
	pkt = NewSubscribe(0, "test-topic", true, 1, true,
		RH_SEND_AT_NEW_SUBSCRIBE, TAT_LONG)
	pkt.SetPacketID(2345)
	assert.Equal(`SUBSCRIBE2(Topic(long)="test-topic", QOS=1, PacketID=2345, NoLocal=true, RAP=true, RH=send_at_new_subscribe)`,
		pkt.String())

	// Illegal topic alias type.
	pkt = NewSubscribe(1234, "test-topic", true, 1, true,
		RH_SEND_AT_NEW_SUBSCRIBE, TAT_LONG)
	pkt.TopicAliasType = 5
	pkt.SetPacketID(2345)
	assert.Equal(`SUBSCRIBE2(Topic="test-topic", TopicAlias=1234, TopicAliasType=illegal(5), QOS=1, PacketID=2345, NoLocal=true, RAP=true, RH=send_at_new_subscribe)`,
		pkt.String())
}

func TestSubscribeInvalidAliasType(t *testing.T) {
	assert := assert.New(t)

	// Topic Alias Type checked in constructor.
	assert.PanicsWithValue("invalid TopicAliasType value: 5", func() {
		NewSubscribe(1234, "test-topic", true, 1, true, RH_SEND_AT_NEW_SUBSCRIBE,
			5, // Illegal Topic Alias Type value
		)
	})

	// Topic Alias Type checked in Pack().
	assert.PanicsWithValue("invalid TopicAliasType value: 5", func() {
		pkt := NewSubscribe(1234, "test-topic", true, 1, true, RH_SEND_AT_NEW_SUBSCRIBE,
			TAT_NORMAL)
		pkt.TopicAliasType = 5
		pkt.Pack()
	})
}

func TestRetainHandlingStringer(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("send_at_subscribe", RH_SEND_AT_SUBSCRIBE.String())
	assert.Equal("send_at_new_subscribe", RH_SEND_AT_NEW_SUBSCRIBE.String())
	assert.Equal("dont_send", RH_DONT_SEND.String())
	assert.Equal("illegal(123)", RetainHandling(123).String())
}
