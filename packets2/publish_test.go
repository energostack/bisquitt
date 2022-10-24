package packets2

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	pkts "github.com/energomonitor/bisquitt/packets"
)

func TestPublishConstructor(t *testing.T) {
	assert := assert.New(t)

	dup := true
	qos := uint8(1)
	retain := true
	aliasType := TAT_SHORT
	topicAlias := uint16(1234)
	topicName := "test/topic"
	data := []byte("test-data")
	pkt := NewPublish(topicAlias, topicName, data, dup, qos, retain, aliasType)

	if pkt == nil {
		t.Fatal("New packet should not be nil")
	}

	assert.Equal("*packets2.Publish", reflect.TypeOf(pkt).String(), "Type should be Publish")
	assert.Equal(topicAlias, pkt.TopicAlias, "Bad TopicAlias value")
	assert.Equal(topicName, pkt.TopicName, "Bad TopicName value")
	assert.Equal(data, pkt.Data, "Bad Data value")
	assert.Equal(dup, pkt.DUP(), "Bad Dup flag value")
	assert.Equal(qos, pkt.QOS, "Bad QOS value")
	assert.Equal(retain, pkt.Retain, "Bad Retain flag value")
	assert.Equal(aliasType, pkt.TopicAliasType, "Bad TopicAliasType value")
	assert.Equal(uint16(0), pkt.PacketID(), "Default PacketID should be 0")
}

func TestPublishMarshal(t *testing.T) {
	// Normal alias.
	pkt1 := NewPublish(1234, "", []byte("test-data"), true, 1, true, TAT_NORMAL)
	pkt1.SetPacketID(2345)
	pkt2 := testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Publish))

	// Predefined alias.
	pkt1 = NewPublish(1234, "", []byte("test-data"), true, 1, true, TAT_PREDEFINED)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Publish))

	// Short topic.
	pkt1 = NewPublish(pkts.EncodeShortTopic("a"), "", []byte("test-data"), true, 1, true, TAT_SHORT)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Publish))

	// Long topic.
	pkt1 = NewPublish(0, "test-topic", []byte("test-data"), true, 1, true, TAT_LONG)
	pkt1.SetPacketID(2345)
	pkt2 = testPacketMarshal(t, pkt1)
	assert.Equal(t, pkt1, pkt2.(*Publish))
}

func TestPublishUnmarshalInvalid(t *testing.T) {
	assert := assert.New(t)

	// Packet too short.
	buff := bytes.NewBuffer([]byte{
		6,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_NORMAL),   // Flags
		0, 2,               // Topic Length
		0, // Packet ID MSB
		// Packet ID LSB missing
	})
	_, err := ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 packet length")
	}

	// Packet too short - normal alias.
	buff = bytes.NewBuffer([]byte{
		8,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_NORMAL),   // Flags
		0, 2,               // Topic Length
		0, 1, // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 packet length")
	}

	// Packet too short - predefined alias.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.PUBLISH),   // Packet Type
		byte(TAT_PREDEFINED), // Flags
		0, 2,                 // Topic Length
		0, 1, // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 packet length")
	}

	// Packet too short - short topic.
	buff = bytes.NewBuffer([]byte{
		8,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_SHORT),    // Flags
		0, 2,               // Topic Length
		0, 1, // Packet ID
		0, // Topic Alias MSB
		// Topic Alias LSB missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 packet length")
	}

	// Packet too short - long topic.
	buff = bytes.NewBuffer([]byte{
		7,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_LONG),     // Flags
		0, 2,               // Topic Length
		0, 1, // Packet ID
		// Topic missing
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 packet length")
	}

	// Topic Length too small - normal alias.
	buff = bytes.NewBuffer([]byte{
		8,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_NORMAL),   // Flags
		0, 1,               // Topic Length
		0, 2, // Packet ID
		0, // Topic Alias
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

	// Topic Length too big - normal alias.
	buff = bytes.NewBuffer([]byte{
		10,                 // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_NORMAL),   // Flags
		0, 3,               // Topic Length
		0, 2, // Packet ID
		0, 3, 4, // Topic Alias
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

	// Topic Length too small - predefined alias.
	buff = bytes.NewBuffer([]byte{
		8,                    // Length
		byte(pkts.PUBLISH),   // Packet Type
		byte(TAT_PREDEFINED), // Flags
		0, 1,                 // Topic Length
		0, 2, // Packet ID
		0, // Topic Alias
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

	// Topic Length too big - predefined alias.
	buff = bytes.NewBuffer([]byte{
		10,                   // Length
		byte(pkts.PUBLISH),   // Packet Type
		byte(TAT_PREDEFINED), // Flags
		0, 3,                 // Topic Length
		0, 2, // Packet ID
		0, 3, 4, // Topic Alias
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

	// Topic Length too small - short topic.
	buff = bytes.NewBuffer([]byte{
		8,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_SHORT),    // Flags
		0, 1,               // Topic Length
		0, 2, // Packet ID
		0, // Topic Alias
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

	// Topic Length too big - short topic.
	buff = bytes.NewBuffer([]byte{
		10,                 // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_SHORT),    // Flags
		0, 3,               // Topic Length
		0, 2, // Packet ID
		0, 3, 4, // Topic Alias
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

	// Topic Length too small - long topic.
	buff = bytes.NewBuffer([]byte{
		7,                  // Length
		byte(pkts.PUBLISH), // Packet Type
		byte(TAT_LONG),     // Flags
		0, 0,               // Topic Length
		0, 2, // Packet ID
	})
	_, err = ReadPacket(buff)
	if assert.Error(err) {
		assert.Contains(err.Error(), "bad PUBLISH2 topic length")
	}

}

func TestPublishStringer(t *testing.T) {
	// Normal alias.
	pkt := NewPublish(1234, "", []byte("test-data"), true, 1, true, TAT_NORMAL)
	pkt.SetPacketID(2345)
	assert.Equal(t,
		`PUBLISH2(TopicAlias(normal)=1234, Data="test-data", QOS=1, Retain=true, PacketID=2345, Dup=true)`,
		pkt.String())

	// Predefined alias.
	pkt = NewPublish(1234, "", []byte("test-data"), true, 1, true, TAT_PREDEFINED)
	pkt.SetPacketID(2345)
	assert.Equal(t,
		`PUBLISH2(TopicAlias(predefined)=1234, Data="test-data", QOS=1, Retain=true, PacketID=2345, Dup=true)`,
		pkt.String())

	// Short topic.
	pkt = NewPublish(pkts.EncodeShortTopic("ab"), "", []byte("test-data"), true, 1, true, TAT_SHORT)
	pkt.SetPacketID(2345)
	assert.Equal(t,
		`PUBLISH2(Topic(short)="ab", Data="test-data", QOS=1, Retain=true, PacketID=2345, Dup=true)`,
		pkt.String())

	// Long topic.
	pkt = NewPublish(0, "test-topic", []byte("test-data"), true, 1, true, TAT_LONG)
	pkt.SetPacketID(2345)
	assert.Equal(t,
		`PUBLISH2(Topic(long)="test-topic", Data="test-data", QOS=1, Retain=true, PacketID=2345, Dup=true)`,
		pkt.String())

	// Illegal topic alias type.
	pkt = NewPublish(1234, "test-topic", []byte("test-data"), true, 1, true, 5)
	pkt.SetPacketID(2345)
	assert.Equal(t,
		`PUBLISH2(Topic="test-topic", TopicAlias=1234, TopicAliasType=illegal(5), Data="test-data", QOS=1, Retain=true, PacketID=2345, Dup=true)`,
		pkt.String())
}
