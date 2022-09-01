package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/stretchr/testify/assert"

	snPkts "github.com/energomonitor/bisquitt/packets"
	snPkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/topics"
	"github.com/energomonitor/bisquitt/util"
)

const (
	// How long to wait to confirm no other packet arrived on the connection.
	// Must be >connTimeout.
	connEmptyTimeout = 500 * time.Millisecond
	// How long to wait for handler to quit.
	handlerQuitTimeout = 2 * connEmptyTimeout
	maxTestPktLength   = 512
)

// New CONNECT transaction must cancel a pending one, if any.  This is because
// some packets of the CONNECT transaction can get lost and the client will
// repeat the CONNECT transaction from the very beginning.
// We want to be sure that the second transaction is a "fresh" one.
func TestRepeatedConnect(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	topicID := uint16(123)
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{
		string(clientID): map[uint16]string{
			topicID: topic,
		},
	})
	defer stp.cancel()

	// client --CONNECT--> GW
	snConnect := snPkts1.NewConnect(clientID, true, false, 1)
	stp.snSend(snConnect, false)

	// GW --CONNECT--> MQTT broker
	mqttConnect := stp.mqttRecv().(*mqPkts.ConnectPacket)
	assert.Equal(string(snConnect.ClientID), mqttConnect.ClientIdentifier)
	assert.Equal(snConnect.CleanSession, mqttConnect.CleanSession)
	assert.Equal(snConnect.Duration, mqttConnect.Keepalive)
	assert.Equal(byte(4), mqttConnect.ProtocolVersion)
	assert.Equal("MQTT", mqttConnect.ProtocolName)

	transaction1, ok := stp.handler.transactions.GetByType(snPkts.CONNECT)
	assert.True(ok)

	// Test transaction1 will be cancelled
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-transaction1.Done():
			return
		case <-time.After(time.Second):
			t.Error("Old transaction cancel timeout!")
		}
	}()

	// client --CONNECT--> GW
	stp.snSend(snConnect, false)

	// GW --CONNECT--> MQTT broker
	mqttConnect = stp.mqttRecv().(*mqPkts.ConnectPacket)
	assert.Equal(string(snConnect.ClientID), mqttConnect.ClientIdentifier)
	assert.Equal(snConnect.CleanSession, mqttConnect.CleanSession)
	assert.Equal(snConnect.Duration, mqttConnect.Keepalive)
	assert.Equal(byte(4), mqttConnect.ProtocolVersion)
	assert.Equal("MQTT", mqttConnect.ProtocolName)

	transaction2, ok := stp.handler.transactions.GetByType(snPkts.CONNECT)
	assert.True(ok)
	assert.NotEqual(transaction1, transaction2)

	// Wait until transaction1 is cancelled.
	wg.Wait()

	// GW <--CONNACK-- MQTT broker
	mqttConnack := mqPkts.NewControlPacket(mqPkts.Connack).(*mqPkts.ConnackPacket)
	mqttConnack.ReturnCode = mqPkts.Accepted
	stp.mqttSend(mqttConnack, false)

	// client <--CONNACK-- GW
	snConnack := stp.snRecv().(*snPkts1.Connack)
	assert.Equal(snPkts1.RC_ACCEPTED, snConnack.ReturnCode)

	assert.Equal(util.StateActive, stp.handler.state.Get())

	// DISCONNECT
	stp.disconnect()
}

// Tests PUBLISH and SUBSCRIBE with predefined topic and QOS 0.
func TestPubSubPredefined(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	topicID := uint16(123)
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{
		string(clientID): map[uint16]string{
			topicID: topic,
		},
	})
	defer stp.cancel()

	// CONNECT
	stp.connect()

	// SUBSCRIBE, PREDEFINED TOPIC

	// client --SUBSCRIBE--> GW
	snSubscribe := snPkts1.NewSubscribe(topicID, snPkts1.TIT_PREDEFINED, nil, 0, false)
	stp.snSend(snSubscribe, true)

	// GW --SUBSCRIBE--> MQTT broker
	mqttSubscribe := stp.mqttRecv().(*mqPkts.SubscribePacket)
	assert.Len(mqttSubscribe.Qoss, 1)
	assert.Equal(snSubscribe.QOS, mqttSubscribe.Qoss[0])
	assert.Len(mqttSubscribe.Topics, 1)
	assert.Equal(topic, mqttSubscribe.Topics[0])

	// GW <--SUBACK-- MQTT broker
	mqttSuback := mqPkts.NewControlPacket(mqPkts.Suback).(*mqPkts.SubackPacket)
	mqttSuback.MessageID = mqttSubscribe.MessageID
	mqttSuback.ReturnCodes = []byte{snSubscribe.QOS}
	stp.mqttSend(mqttSuback, false)

	// client <--SUBACK-- GW
	snSuback := stp.snRecv().(*snPkts1.Suback)
	assert.Equal(snSubscribe.MessageID(), snSuback.MessageID())
	assert.Equal(snPkts1.RC_ACCEPTED, snSuback.ReturnCode)

	// PUBLISH QOS 0, PREDEFINED TOPIC

	payload := []byte("test-msg-1")

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_PREDEFINED, payload, 0, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.Equal(snPublish.QOS, mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(snPublish.Data, mqttPublish.Payload)

	// SUBSCRIPTION PUBLISH

	// GW <--PUBLISH-- MQTT broker
	stp.mqttSend(mqttPublish, true)

	// client <--PUBLISH-- GW
	snPublish = stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_PREDEFINED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)

	// DISCONNECT
	stp.disconnect()
}

// Tests PUBLISH and SUBSCRIBE with predefined topic, QOS 0 and long packet.
func TestPubSubPredefinedLong(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	topicID := uint16(123)
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{
		string(clientID): map[uint16]string{
			topicID: topic,
		},
	})
	defer stp.cancel()

	// CONNECT
	stp.connect()

	// SUBSCRIBE, PREDEFINED TOPIC

	// client --SUBSCRIBE--> GW
	snSubscribe := snPkts1.NewSubscribe(topicID, snPkts1.TIT_PREDEFINED, nil, 0, false)
	stp.snSend(snSubscribe, true)

	// GW --SUBSCRIBE--> MQTT broker
	mqttSubscribe := stp.mqttRecv().(*mqPkts.SubscribePacket)
	assert.Len(mqttSubscribe.Qoss, 1)
	assert.Equal(snSubscribe.QOS, mqttSubscribe.Qoss[0])
	assert.Len(mqttSubscribe.Topics, 1)
	assert.Equal(topic, mqttSubscribe.Topics[0])

	// GW <--SUBACK-- MQTT broker
	mqttSuback := mqPkts.NewControlPacket(mqPkts.Suback).(*mqPkts.SubackPacket)
	mqttSuback.MessageID = mqttSubscribe.MessageID
	mqttSuback.ReturnCodes = []byte{snSubscribe.QOS}
	stp.mqttSend(mqttSuback, false)

	// client <--SUBACK-- GW
	snSuback := stp.snRecv().(*snPkts1.Suback)
	assert.Equal(snSubscribe.MessageID(), snSuback.MessageID())
	assert.Equal(snPkts1.RC_ACCEPTED, snSuback.ReturnCode)

	// PUBLISH QOS 0, PREDEFINED TOPIC

	// We are mocking a network connection with a unix socket so the whole packet
	// length must be <= 512B.
	payloadSize := 384
	payload := make([]byte, payloadSize)
	for i := 0; i < payloadSize; i++ {
		payload[i] = byte(i)
	}

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_PREDEFINED, payload, 0, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.Equal(snPublish.QOS, mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(payload, mqttPublish.Payload)

	// SUBSCRIPTION PUBLISH

	// GW <--PUBLISH-- MQTT broker
	stp.mqttSend(mqttPublish, true)

	// client <--PUBLISH-- GW
	snPublish = stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_PREDEFINED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)

	// DISCONNECT
	stp.disconnect()
}

// REGISTER packet without previous CONNECT is illegal.
// The gateway should close the connection immediately.
func TestDisconnectedRegister(t *testing.T) {
	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	stp.snSend(
		snPkts1.NewRegister(0, "test-topic-0"),
		true,
	)

	stp.assertHandlerDone()
}

// SUBSCRIBE packet without previous CONNECT is illegal.
// The gateway should close the connection immediately.
func TestDisconnectedSubscribe(t *testing.T) {
	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	stp.snSend(
		snPkts1.NewSubscribe(0, snPkts1.TIT_STRING,
			[]byte("test-topic-0"), 0, false),
		true,
	)

	stp.assertHandlerDone()
}

// PUBLISH(QOS 0) packet without previous CONNECT is illegal.
// The gateway should close the connection immediately.
func TestDisconnectedPublishQOS0(t *testing.T) {
	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	snPublish := snPkts1.NewPublish(
		snPkts1.EncodeShortTopic("ab"),
		snPkts1.TIT_SHORT, []byte("test-payload"), 0, false, false,
	)
	stp.snSend(snPublish, true)

	stp.assertHandlerDone()
}

// PUBLISH(QOS -1, registered topic) packet without previous CONNECT is illegal.
// The gateway should close the connection immediately.
func TestDisconnectedPublishQOS3Registered(t *testing.T) {
	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	snPublish := snPkts1.NewPublish(
		123, snPkts1.TIT_REGISTERED, []byte("test-payload"), 3, false, false)
	stp.snSend(snPublish, true)

	stp.assertHandlerDone()
}

// PUBLISH(QOS -1, short topic) is illegal if the authentication is enabled.
// The gateway should close the connection immediately.
func TestDisconnectedAuthPublishQOS3Registered(t *testing.T) {
	stp := newTestSetup(t, true, topics.PredefinedTopics{})
	defer stp.cancel()

	topic := "ab"
	topicID := snPkts1.EncodeShortTopic(topic)
	payload := []byte("test-msg-0")
	qos := uint8(3)

	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_SHORT,
		payload, qos, false, false)
	stp.snSend(snPublish, true)

	stp.assertHandlerDone()
}

// PUBLISH(QOS -1, short topic) is legal even without previous CONNECT if
// the authentication is disabled.
func TestDisconnectedPublishQOS3Short(t *testing.T) {
	assert := assert.New(t)

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	topic := "ab"
	topicID := snPkts1.EncodeShortTopic(topic)
	payload := []byte("test-msg-0")
	qos := uint8(3)

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_SHORT,
		payload, qos, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.Equal(uint8(0), mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(payload, mqttPublish.Payload)

	// DISCONNECT
	stp.disconnect()
}

// PUBLISH(QOS -1, predefined topic) is legal even without previous CONNECT if
// the authentication is disabled.
func TestDisconnectedPublishQOS3Predefined(t *testing.T) {
	assert := assert.New(t)

	topic := "test-topic-0"
	topicID := uint16(123)
	payload := []byte("test-msg-0")
	qos := uint8(3)

	stp := newTestSetup(t, false, topics.PredefinedTopics{
		"*": map[uint16]string{
			topicID: topic,
		},
	})
	defer stp.cancel()

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_PREDEFINED,
		payload, qos, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.Equal(uint8(0), mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(payload, mqttPublish.Payload)

	// DISCONNECT
	stp.disconnect()
}

// Test PUBLISH with string topic and QOS 0,1,2.
func TestClientPublishQOS0(t *testing.T) {
	assert := assert.New(t)

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	topic := "test-topic-0"
	payload := []byte("test-msg-0")

	stp.connect()
	topicID := stp.register(topic)

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_REGISTERED, payload, 0, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.Equal(uint16(0), mqttPublish.MessageID)
	assert.Equal(snPublish.QOS, mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(snPublish.Data, mqttPublish.Payload)

	// DISCONNECT
	stp.disconnect()
}

func TestClientPublishQOS1(t *testing.T) {
	assert := assert.New(t)

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	topic := "test-topic-1"
	payload := []byte("test-msg-1")

	stp.connect()
	topicID := stp.register(topic)

	mqttNextMsgID := stp.mqttNextMsgID
	payload = []byte("test-msg-1")

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_REGISTERED, payload, 1, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.NotEqual(mqttNextMsgID, mqttPublish.MessageID)
	assert.Equal(snPublish.QOS, mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(snPublish.Data, mqttPublish.Payload)

	// GW <--PUBACK-- MQTT broker
	mqttPuback := mqPkts.NewControlPacket(mqPkts.Puback).(*mqPkts.PubackPacket)
	mqttPuback.MessageID = mqttPublish.MessageID
	stp.mqttSend(mqttPuback, false)

	// client <--PUBACK-- GW
	snPuback := stp.snRecv().(*snPkts1.Puback)
	assert.Equal(snPublish.MessageID(), snPuback.MessageID())

	// DISCONNECT
	stp.disconnect()
}

func TestClientPublishQOS2(t *testing.T) {
	assert := assert.New(t)

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	topic := "test-topic-2"
	payload := []byte("test-msg-2")

	stp.connect()
	topicID := stp.register(topic)

	mqttNextMsgID := stp.mqttNextMsgID
	payload = []byte("test-msg-2")

	// client --PUBLISH--> GW
	snPublish := snPkts1.NewPublish(topicID, snPkts1.TIT_REGISTERED, payload, 2, false, false)
	stp.snSend(snPublish, true)

	// GW --PUBLISH--> MQTT broker
	mqttPublish := stp.mqttRecv().(*mqPkts.PublishPacket)
	assert.NotEqual(mqttNextMsgID, mqttPublish.MessageID)
	assert.Equal(snPublish.QOS, mqttPublish.Qos)
	assert.Equal(topic, mqttPublish.TopicName)
	assert.Equal(snPublish.Data, mqttPublish.Payload)

	// GW <--PUBREC-- MQTT broker
	mqttPubrec := mqPkts.NewControlPacket(mqPkts.Pubrec).(*mqPkts.PubrecPacket)
	mqttPubrec.MessageID = mqttPublish.MessageID
	stp.mqttSend(mqttPubrec, false)

	// client <--PUBREC-- GW
	snPubrec := stp.snRecv().(*snPkts1.Pubrec)
	assert.Equal(snPublish.MessageID(), snPubrec.MessageID())

	// client --PUBREL--> GW
	snPubrel := snPkts1.NewPubrel()
	snPubrel.SetMessageID(mqttPublish.MessageID)
	stp.snSend(snPubrel, false)

	// GW --PUBREL--> MQTT broker
	mqttPubrel := stp.mqttRecv().(*mqPkts.PubrelPacket)
	assert.Equal(snPublish.MessageID(), mqttPubrel.MessageID)

	// DISCONNECT
	stp.disconnect()
}

func TestSubscribeQOS0Wildcard(t *testing.T) {
	assert := assert.New(t)

	wildcard := "test/+"
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	stp.subscribe(wildcard, 0)

	// SUBSCRIPTION PUBLISH FROM MQTT BROKER, LOST MQTT-SN PUBACK

	payload := []byte("test-msg-1")
	qos := uint8(0)

	// GW <--PUBLISH-- MQTT broker
	mqttPublish := mqPkts.NewControlPacket(mqPkts.Publish).(*mqPkts.PublishPacket)
	mqttPublish.Qos = qos
	mqttPublish.TopicName = topic
	mqttPublish.Payload = payload
	stp.mqttSend(mqttPublish, true)

	// client <--REGISTER-- GW
	snRegister := stp.snRecv().(*snPkts1.Register)
	assert.Equal(topic, snRegister.TopicName)
	assert.GreaterOrEqual(snRegister.MessageID(), snPkts1.MinMessageID)
	assert.LessOrEqual(snRegister.MessageID(), snPkts1.MaxMessageID)
	topicID := snRegister.TopicID

	// client --REGACK--> GW
	snRegack := snPkts1.NewRegack(topicID, snPkts1.RC_ACCEPTED)
	snRegack.SetMessageID(snRegister.MessageID())
	stp.snSend(snRegack, false)

	// client <--PUBLISH-- GW
	snPublish := stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)
	assert.Equal(qos, snPublish.QOS)
	assert.Equal(false, snPublish.DUP())

	// DISCONNECT
	stp.disconnect()
}

func TestSubscribeQOS1(t *testing.T) {
	assert := assert.New(t)

	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	topicID := stp.subscribe(topic, 1)

	// SUBSCRIPTION PUBLISH FROM MQTT BROKER, LOST MQTT-SN PUBACK

	payload := []byte("test-msg-1")
	qos := uint8(1)

	// GW <--PUBLISH-- MQTT broker
	mqttPublish := mqPkts.NewControlPacket(mqPkts.Publish).(*mqPkts.PublishPacket)
	mqttPublish.Qos = qos
	mqttPublish.TopicName = topic
	mqttPublish.Payload = payload
	stp.mqttSend(mqttPublish, true)

	// client <--PUBLISH-- GW
	snPublish := stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)
	assert.Equal(qos, snPublish.QOS)
	assert.Equal(false, snPublish.DUP())

	// Two lost PUBACKs => two PUBLISH resends
	for i := 0; i < 2; i++ {
		// (lost: client --PUBACK--> GW)

		// resend: client <--PUBLISH-- GW
		snPublish = stp.snRecv().(*snPkts1.Publish)
		assert.Equal(topicID, snPublish.TopicID)
		assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
		assert.Equal(payload, snPublish.Data)
		assert.Equal(qos, snPublish.QOS)
		assert.Equal(true, snPublish.DUP())
	}

	// client --PUBACK--> GW
	snPuback := snPkts1.NewPuback(snPublish.TopicID, snPkts1.RC_ACCEPTED)
	snPuback.SetMessageID(snPublish.MessageID())
	stp.snSend(snPuback, false)

	// GW --PUBACK--> MQTT broker
	mqttPuback := stp.mqttRecv().(*mqPkts.PubackPacket)
	assert.Equal(snPuback.MessageID(), mqttPuback.MessageID)

	// No more resends expected...
	time.Sleep(stp.handler.cfg.RetryDelay * 2)

	// DISCONNECT
	stp.disconnect()
}

func TestSubscribeQOS1Wildcard(t *testing.T) {
	assert := assert.New(t)

	wildcard := "test/+"
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	stp.subscribe(wildcard, 1)

	// SUBSCRIPTION PUBLISH FROM MQTT BROKER, LOST MQTT-SN PUBACK

	payload := []byte("test-msg-1")
	qos := uint8(1)

	// GW <--PUBLISH-- MQTT broker
	mqttPublish := mqPkts.NewControlPacket(mqPkts.Publish).(*mqPkts.PublishPacket)
	mqttPublish.Qos = qos
	mqttPublish.TopicName = topic
	mqttPublish.Payload = payload
	stp.mqttSend(mqttPublish, true)

	// client <--REGISTER-- GW
	snRegister := stp.snRecv().(*snPkts1.Register)
	assert.Equal(topic, snRegister.TopicName)
	topicID := snRegister.TopicID

	// client --REGACK--> GW
	snRegack := snPkts1.NewRegack(topicID, snPkts1.RC_ACCEPTED)
	snRegack.SetMessageID(snRegister.MessageID())
	stp.snSend(snRegack, false)

	// client <--PUBLISH-- GW
	snPublish := stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)
	assert.Equal(qos, snPublish.QOS)
	assert.Equal(false, snPublish.DUP())

	// Two lost PUBACKs => two PUBLISH resends
	for i := 0; i < 2; i++ {
		// (lost: client --PUBACK--> GW)

		// resend: client <--PUBLISH-- GW
		snPublish = stp.snRecv().(*snPkts1.Publish)
		assert.Equal(topicID, snPublish.TopicID)
		assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
		assert.Equal(payload, snPublish.Data)
		assert.Equal(qos, snPublish.QOS)
		assert.Equal(true, snPublish.DUP())
	}

	// client --PUBACK--> GW
	snPuback := snPkts1.NewPuback(snPublish.TopicID, snPkts1.RC_ACCEPTED)
	snPuback.SetMessageID(snPublish.MessageID())
	stp.snSend(snPuback, false)

	// GW --PUBACK--> MQTT broker
	mqttPuback := stp.mqttRecv().(*mqPkts.PubackPacket)
	assert.Equal(snPuback.MessageID(), mqttPuback.MessageID)

	// No more resends expected...
	time.Sleep(stp.handler.cfg.RetryDelay * 2)

	// DISCONNECT
	stp.disconnect()
}

func TestSubscribeQOS2(t *testing.T) {
	assert := assert.New(t)

	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	topicID := stp.subscribe(topic, 2)

	// SUBSCRIPTION PUBLISH FROM MQTT BROKER

	payload := []byte("test-msg-1")
	qos := uint8(2)

	// GW <--PUBLISH-- MQTT broker
	mqttPublish := mqPkts.NewControlPacket(mqPkts.Publish).(*mqPkts.PublishPacket)
	mqttPublish.Qos = qos
	mqttPublish.TopicName = topic
	mqttPublish.Payload = payload
	stp.mqttSend(mqttPublish, true)
	msgID := mqttPublish.MessageID

	// client <--PUBLISH-- GW
	snPublish := stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)
	assert.Equal(qos, snPublish.QOS)
	assert.Equal(msgID, snPublish.MessageID())

	// Two lost PUBRECs => two PUBLISH resends
	for i := 0; i < 2; i++ {
		// (lost: client --PUBREC--> GW)

		// resend: client <--PUBLISH-- GW
		snPublish := stp.snRecv().(*snPkts1.Publish)
		assert.Equal(topicID, snPublish.TopicID)
		assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
		assert.Equal(payload, snPublish.Data)
		assert.Equal(qos, snPublish.QOS)
		assert.Equal(msgID, snPublish.MessageID())
		assert.Equal(true, snPublish.DUP())
	}

	// client --PUBREC--> GW
	snPubrec := snPkts1.NewPubrec()
	snPubrec.SetMessageID(msgID)
	stp.snSend(snPubrec, false)

	// GW --PUBREC--> MQTT broker
	mqttPubrec := stp.mqttRecv().(*mqPkts.PubrecPacket)
	assert.Equal(msgID, mqttPubrec.MessageID)

	// Lost MQTT PUBREC or PUBREL => MQTT PUBREC resend
	// GW --PUBREC--> MQTT broker
	mqttPubrec = stp.mqttRecv().(*mqPkts.PubrecPacket)
	assert.Equal(msgID, mqttPubrec.MessageID)

	// GW <--PUBREL-- MQTT broker
	mqttPubrel := mqPkts.NewControlPacket(mqPkts.Pubrel).(*mqPkts.PubrelPacket)
	mqttPubrel.MessageID = msgID
	stp.mqttSend(mqttPubrel, false)

	// client <--PUBREL-- GW
	snPubrel := stp.snRecv().(*snPkts1.Pubrel)
	assert.Equal(msgID, snPubrel.MessageID())

	// Two lost PUBCOMPs => two PUBREL resends
	for i := 0; i < 2; i++ {
		// (lost: client --PUBCOMP--> GW)

		// resend: client <--PUBREL-- GW
		snPubrel := stp.snRecv().(*snPkts1.Pubrel)
		assert.Equal(msgID, snPubrel.MessageID())
	}

	// client --PUBCOMP--> GW
	snPubcomp := snPkts1.NewPubcomp()
	snPubcomp.SetMessageID(msgID)
	stp.snSend(snPubcomp, false)

	// GW --PUBCOMP--> MQTT broker
	mqttPubcomp := stp.mqttRecv().(*mqPkts.PubcompPacket)
	assert.Equal(msgID, mqttPubcomp.MessageID)

	// DISCONNECT
	stp.disconnect()
}

func TestSubscribeQOS2Wildcard(t *testing.T) {
	assert := assert.New(t)

	wildcard := "test/+"
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	stp.subscribe(wildcard, 2)

	// SUBSCRIPTION PUBLISH FROM MQTT BROKER

	payload := []byte("test-msg-1")
	qos := uint8(2)

	// GW <--PUBLISH-- MQTT broker
	mqttPublish := mqPkts.NewControlPacket(mqPkts.Publish).(*mqPkts.PublishPacket)
	mqttPublish.Qos = qos
	mqttPublish.TopicName = topic
	mqttPublish.Payload = payload
	stp.mqttSend(mqttPublish, true)
	msgID := mqttPublish.MessageID

	// client <--REGISTER-- GW
	snRegister := stp.snRecv().(*snPkts1.Register)
	assert.Equal(topic, snRegister.TopicName)
	topicID := snRegister.TopicID

	// client --REGACK--> GW
	snRegack := snPkts1.NewRegack(topicID, snPkts1.RC_ACCEPTED)
	snRegack.SetMessageID(snRegister.MessageID())
	stp.snSend(snRegack, false)

	// client <--PUBLISH-- GW
	snPublish := stp.snRecv().(*snPkts1.Publish)
	assert.Equal(topicID, snPublish.TopicID)
	assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
	assert.Equal(payload, snPublish.Data)
	assert.Equal(qos, snPublish.QOS)
	assert.Equal(msgID, snPublish.MessageID())

	// Two lost PUBRECs => two PUBLISH resends
	for i := 0; i < 2; i++ {
		// (lost: client --PUBREC--> GW)

		// resend: client <--PUBLISH-- GW
		snPublish := stp.snRecv().(*snPkts1.Publish)
		assert.Equal(topicID, snPublish.TopicID)
		assert.Equal(snPkts1.TIT_REGISTERED, snPublish.TopicIDType)
		assert.Equal(payload, snPublish.Data)
		assert.Equal(qos, snPublish.QOS)
		assert.Equal(msgID, snPublish.MessageID())
		assert.Equal(true, snPublish.DUP())
	}

	// client --PUBREC--> GW
	snPubrec := snPkts1.NewPubrec()
	snPubrec.SetMessageID(msgID)
	stp.snSend(snPubrec, false)

	// GW --PUBREC--> MQTT broker
	mqttPubrec := stp.mqttRecv().(*mqPkts.PubrecPacket)
	assert.Equal(msgID, mqttPubrec.MessageID)

	// Lost MQTT PUBREC or PUBREL => MQTT PUBREC resend
	// GW --PUBREC--> MQTT broker
	mqttPubrec = stp.mqttRecv().(*mqPkts.PubrecPacket)
	assert.Equal(msgID, mqttPubrec.MessageID)

	// GW <--PUBREL-- MQTT broker
	mqttPubrel := mqPkts.NewControlPacket(mqPkts.Pubrel).(*mqPkts.PubrelPacket)
	mqttPubrel.MessageID = msgID
	stp.mqttSend(mqttPubrel, false)

	// client <--PUBREL-- GW
	snPubrel := stp.snRecv().(*snPkts1.Pubrel)
	assert.Equal(msgID, snPubrel.MessageID())

	// Two lost PUBCOMPs => two PUBREL resends
	for i := 0; i < 2; i++ {
		// (lost: client --PUBCOMP--> GW)

		// resend: client <--PUBREL-- GW
		snPubrel := stp.snRecv().(*snPkts1.Pubrel)
		assert.Equal(msgID, snPubrel.MessageID())
	}

	// client --PUBCOMP--> GW
	snPubcomp := snPkts1.NewPubcomp()
	snPubcomp.SetMessageID(msgID)
	stp.snSend(snPubcomp, false)

	// GW --PUBCOMP--> MQTT broker
	mqttPubcomp := stp.mqttRecv().(*mqPkts.PubcompPacket)
	assert.Equal(msgID, mqttPubcomp.MessageID)

	// DISCONNECT
	stp.disconnect()
}

func TestUnsubscribeString(t *testing.T) {
	assert := assert.New(t)

	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	stp.subscribe(topic, 0)

	// client --UNSUBSCRIBE--> GW
	snUnsubscribe := snPkts1.NewUnsubscribe(0, snPkts1.TIT_STRING, []byte(topic))
	stp.snSend(snUnsubscribe, true)

	// GW --UNSUBSCRIBE--> MQTT broker
	mqttUnsubscribe := stp.mqttRecv().(*mqPkts.UnsubscribePacket)
	assert.Len(mqttUnsubscribe.Topics, 1)
	assert.Equal(topic, mqttUnsubscribe.Topics[0])

	// GW <--UNSUBACK-- MQTT broker
	mqttUnsuback := mqPkts.NewControlPacket(mqPkts.Unsuback).(*mqPkts.UnsubackPacket)
	mqttUnsuback.MessageID = mqttUnsubscribe.MessageID
	stp.mqttSend(mqttUnsuback, false)

	// client <--UNSUBACK-- GW
	snUnsuback := stp.snRecv().(*snPkts1.Unsuback)
	assert.Equal(snUnsubscribe.MessageID(), snUnsuback.MessageID())

	// DISCONNECT
	stp.disconnect()
}

func TestUnsubscribeShort(t *testing.T) {
	assert := assert.New(t)

	topic := "ab"

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()
	stp.subscribeShort(topic, 0)

	// client --UNSUBSCRIBE--> GW
	snUnsubscribe := snPkts1.NewUnsubscribe(snPkts1.EncodeShortTopic(topic), snPkts1.TIT_SHORT, []byte(""))
	stp.snSend(snUnsubscribe, true)

	// GW --UNSUBSCRIBE--> MQTT broker
	mqttUnsubscribe := stp.mqttRecv().(*mqPkts.UnsubscribePacket)
	assert.Len(mqttUnsubscribe.Topics, 1)
	assert.Equal(topic, mqttUnsubscribe.Topics[0])

	// GW <--UNSUBACK-- MQTT broker
	mqttUnsuback := mqPkts.NewControlPacket(mqPkts.Unsuback).(*mqPkts.UnsubackPacket)
	mqttUnsuback.MessageID = mqttUnsubscribe.MessageID
	stp.mqttSend(mqttUnsuback, false)

	// client <--UNSUBACK-- GW
	snUnsuback := stp.snRecv().(*snPkts1.Unsuback)
	assert.Equal(snUnsubscribe.MessageID(), snUnsuback.MessageID())

	// DISCONNECT
	stp.disconnect()
}

func TestUnsubscribePredefined(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	topicID := uint16(123)
	topic := "test/topic"

	stp := newTestSetup(t, false, topics.PredefinedTopics{
		string(clientID): map[uint16]string{
			topicID: topic,
		},
	})
	defer stp.cancel()

	// CONNECT, SUBSCRIBE
	stp.connect()

	// SUBSCRIBE, PREDEFINED TOPIC

	// client --SUBSCRIBE--> GW
	snSubscribe := snPkts1.NewSubscribe(topicID, snPkts1.TIT_PREDEFINED, nil, 0, false)
	stp.snSend(snSubscribe, true)

	// GW --SUBSCRIBE--> MQTT broker
	mqttSubscribe := stp.mqttRecv().(*mqPkts.SubscribePacket)
	assert.Len(mqttSubscribe.Qoss, 1)
	assert.Equal(snSubscribe.QOS, mqttSubscribe.Qoss[0])
	assert.Len(mqttSubscribe.Topics, 1)
	assert.Equal(topic, mqttSubscribe.Topics[0])

	// GW <--SUBACK-- MQTT broker
	mqttSuback := mqPkts.NewControlPacket(mqPkts.Suback).(*mqPkts.SubackPacket)
	mqttSuback.MessageID = mqttSubscribe.MessageID
	mqttSuback.ReturnCodes = []byte{snSubscribe.QOS}
	stp.mqttSend(mqttSuback, false)

	// client <--SUBACK-- GW
	snSuback := stp.snRecv().(*snPkts1.Suback)
	assert.Equal(snSubscribe.MessageID(), snSuback.MessageID())
	assert.Equal(snPkts1.RC_ACCEPTED, snSuback.ReturnCode)

	// client --UNSUBSCRIBE--> GW
	snUnsubscribe := snPkts1.NewUnsubscribe(topicID, snPkts1.TIT_PREDEFINED, []byte(""))
	stp.snSend(snUnsubscribe, true)

	// GW --UNSUBSCRIBE--> MQTT broker
	mqttUnsubscribe := stp.mqttRecv().(*mqPkts.UnsubscribePacket)
	assert.Len(mqttUnsubscribe.Topics, 1)
	assert.Equal(topic, mqttUnsubscribe.Topics[0])

	// GW <--UNSUBACK-- MQTT broker
	mqttUnsuback := mqPkts.NewControlPacket(mqPkts.Unsuback).(*mqPkts.UnsubackPacket)
	mqttUnsuback.MessageID = mqttUnsubscribe.MessageID
	stp.mqttSend(mqttUnsuback, false)

	// client <--UNSUBACK-- GW
	snUnsuback := stp.snRecv().(*snPkts1.Unsuback)
	assert.Equal(snUnsubscribe.MessageID(), snUnsuback.MessageID())

	// DISCONNECT
	stp.disconnect()
}

func TestLastWill(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	willTopic := "test/status"
	willPayload := []byte("offline")
	willQos := uint8(1)
	willRetain := true
	keepalive := uint16(1)

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT

	// client --CONNECT--> GW
	snConnect := snPkts1.NewConnect(clientID, true, true, keepalive)
	stp.snSend(snConnect, false)

	// client <--WILLTOPICREQ-- GW
	_, ok := stp.snRecv().(*snPkts1.WillTopicReq)
	assert.True(ok)

	// client --CONNECT--> GW
	snWillTopic := snPkts1.NewWillTopic(willTopic, willQos, willRetain)
	stp.snSend(snWillTopic, false)

	// client <--WILLMSGREQ-- GW
	_, ok = stp.snRecv().(*snPkts1.WillMsgReq)
	assert.True(ok)

	// client --CONNECT--> GW
	snWillMsg := snPkts1.NewWillMsg(willPayload)
	stp.snSend(snWillMsg, false)

	// GW --CONNECT--> MQTT broker
	mqttConnect := stp.mqttRecv().(*mqPkts.ConnectPacket)
	assert.True(mqttConnect.WillFlag)
	assert.Equal(willTopic, mqttConnect.WillTopic)
	assert.Equal(willPayload, mqttConnect.WillMessage)
	assert.Equal(willRetain, mqttConnect.WillRetain)
	assert.Equal(willQos, mqttConnect.WillQos)

	// GW <--CONNACK-- MQTT broker
	mqttConnack := mqPkts.NewControlPacket(mqPkts.Connack).(*mqPkts.ConnackPacket)
	mqttConnack.ReturnCode = mqPkts.Accepted
	stp.mqttSend(mqttConnack, false)

	// client <--CONNACK-- GW
	snConnack := stp.snRecv().(*snPkts1.Connack)
	assert.Equal(snPkts1.RC_ACCEPTED, snConnack.ReturnCode)

	assert.Equal(util.StateActive, stp.handler.state.Get())

	// Now, it is a MQTT broker's responsibility to send the last will packet
	// when the client dies unexpectedly - we can't test it here.
	// The broker will detect a dead client by not receiving PINGREQ for
	// >keepalive and will close the connection.
	// Handler should detect the closed MQTT connection and quit.

	// Simulate that the MQTT broker closed connection.
	stp.mqttConn.Close()

	// NOTE: It is unclean if the gateway should send DISCONNECT here or not.
	// client <--DISCONNECT-- GW
	snDisconnectReply := stp.snRecv().(*snPkts1.Disconnect)
	assert.Equal(uint16(0), snDisconnectReply.Duration)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		stp.assertConnEmpty("MQTT-SN", stp.snConn, connEmptyTimeout)
	}()
	wg.Wait()

	// Handler must quit afterwards.
	select {
	case <-time.After(handlerQuitTimeout):
		stp.t.Error("handler did not quit")
	case <-stp.handlerDone:
		// OK
	}
}

func TestConnectTimeout(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")

	stp := newTestSetup(t, false, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT

	// client --CONNECT--> GW
	snConnect := snPkts1.NewConnect(clientID, true, true, 2)
	stp.snSend(snConnect, false)

	// client <--WILLTOPICREQ-- GW
	_, ok := stp.snRecv().(*snPkts1.WillTopicReq)
	assert.True(ok)

	// A malicious client does not continue the transaction.
	// The handler must be cancelled after at most connectTransactionTimeout.
	time.Sleep(connectTransactionTimeout)

	stp.assertHandlerDone()
}

func TestAuthSuccess(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	user := "test-user"
	password := []byte("test-pwd")

	stp := newTestSetup(t, true, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT

	// client --CONNECT--> GW
	snConnect := snPkts1.NewConnect(clientID, true, false, 1)
	stp.snSend(snConnect, false)

	// client --AUTH--> GW
	snAuth := snPkts1.NewAuthPlain(user, password)
	stp.snSend(snAuth, false)

	// GW --CONNECT--> MQTT broker
	mqttConnect := stp.mqttRecv().(*mqPkts.ConnectPacket)
	assert.Equal(string(snConnect.ClientID), mqttConnect.ClientIdentifier)
	assert.Equal(snConnect.CleanSession, mqttConnect.CleanSession)
	assert.Equal(snConnect.Duration, mqttConnect.Keepalive)
	assert.Equal(byte(4), mqttConnect.ProtocolVersion)
	assert.Equal("MQTT", mqttConnect.ProtocolName)
	assert.Equal(true, mqttConnect.UsernameFlag)
	assert.Equal(user, mqttConnect.Username)
	assert.Equal(true, mqttConnect.PasswordFlag)
	assert.Equal(password, mqttConnect.Password)

	// GW <--CONNACK-- MQTT broker
	mqttConnack := mqPkts.NewControlPacket(mqPkts.Connack).(*mqPkts.ConnackPacket)
	mqttConnack.ReturnCode = mqPkts.Accepted
	stp.mqttSend(mqttConnack, false)

	// client <--CONNACK-- GW
	snConnack := stp.snRecv().(*snPkts1.Connack)
	assert.Equal(snPkts1.RC_ACCEPTED, snConnack.ReturnCode)

	assert.Equal(util.StateActive, stp.handler.state.Get())
}

func TestAuthFail(t *testing.T) {
	assert := assert.New(t)

	clientID := []byte("test-client")
	user := "test-user"
	password := []byte("test-pwd")

	stp := newTestSetup(t, true, topics.PredefinedTopics{})
	defer stp.cancel()

	// CONNECT

	// client --CONNECT--> GW
	snConnect := snPkts1.NewConnect(clientID, true, false, 1)
	stp.snSend(snConnect, false)

	// client --AUTH--> GW
	snAuth := snPkts1.NewAuthPlain(user, password)
	stp.snSend(snAuth, false)

	// GW --CONNECT--> MQTT broker
	mqttConnect := stp.mqttRecv().(*mqPkts.ConnectPacket)
	assert.Equal(string(snConnect.ClientID), mqttConnect.ClientIdentifier)
	assert.Equal(snConnect.CleanSession, mqttConnect.CleanSession)
	assert.Equal(snConnect.Duration, mqttConnect.Keepalive)
	assert.Equal(byte(4), mqttConnect.ProtocolVersion)
	assert.Equal("MQTT", mqttConnect.ProtocolName)
	assert.Equal(true, mqttConnect.UsernameFlag)
	assert.Equal(user, mqttConnect.Username)
	assert.Equal(true, mqttConnect.PasswordFlag)
	assert.Equal(password, mqttConnect.Password)

	// GW <--CONNACK-- MQTT broker
	mqttConnack := mqPkts.NewControlPacket(mqPkts.Connack).(*mqPkts.ConnackPacket)
	mqttConnack.ReturnCode = mqPkts.ErrRefusedNotAuthorised
	stp.mqttSend(mqttConnack, false)

	// client <--CONNACK-- GW
	snConnack := stp.snRecv().(*snPkts1.Connack)
	assert.Equal(snPkts1.RC_CONGESTION, snConnack.ReturnCode)

	assert.Equal(util.StateDisconnected, stp.handler.state.Get())
}

//
// testSetup
//

type testSetup struct {
	ID            string
	t             *testing.T
	mqttConn      net.Conn
	snConn        net.Conn
	snNextMsgID   uint16
	mqttNextMsgID uint16
	ctx           context.Context
	cancel        context.CancelFunc
	handler       *handler1
	handlerDone   chan struct{}
}

func newTestSetup(t *testing.T, auth bool, predefinedTopics topics.PredefinedTopics) *testSetup {
	ctx, cancel := context.WithCancel(context.Background())
	handlerDone := make(chan struct{})
	// Test name without "Test" prefix.
	id := reflect.ValueOf(*t).FieldByName("name").String()[4:]
	stp := &testSetup{
		ID:            id,
		t:             t,
		ctx:           ctx,
		cancel:        cancel,
		handlerDone:   handlerDone,
		snNextMsgID:   1,
		mqttNextMsgID: 1,
	}
	stp.newHandler(auth, predefinedTopics)
	return stp
}

func (stp *testSetup) newHandler(auth bool, predefinedTopics topics.PredefinedTopics) {
	log := util.NewDebugLogger("h-" + stp.ID)

	var snListener *net.UnixListener
	var mqttListener *net.UnixListener
	snListener, stp.snConn = stp.createSocketPair("unixpacket")
	mqttListener, stp.mqttConn = stp.createSocketPair("unix")

	handlerChan := make(chan *handler1)
	go func() {
		defer close(stp.handlerDone)

		snConnGateway, err := snListener.AcceptUnix()
		if err != nil {
			stp.t.Fatal(err)
		}
		mqttConnGateway, err := mqttListener.AcceptUnix()
		if err != nil {
			stp.t.Fatal(err)
		}

		cfg := &handlerConfig{
			AuthEnabled: auth,
			RetryDelay:  time.Second,
			RetryCount:  2,
		}
		handler := newHandler(cfg, predefinedTopics, log)
		handler.mockupDialFunc = func() net.Conn {
			return mqttConnGateway
		}
		select {
		case <-stp.ctx.Done():
			return
		case handlerChan <- handler:
			// continue
		}
		handler.run(stp.ctx, snConnGateway)
	}()

	select {
	case <-stp.ctx.Done():
		return
	case stp.handler = <-handlerChan:
		// continue
	}
}

func (stp *testSetup) createSocketPair(sockType string) (*net.UnixListener, *net.UnixConn) {
	// NOTE: "@" means "unnamed socket"
	socket := fmt.Sprintf("@%d", rand.Uint64())
	addr := &net.UnixAddr{Name: socket, Net: sockType}

	listener, err := net.ListenUnix(sockType, addr)
	if err != nil {
		stp.t.Fatal(err)
	}

	conn, err := net.DialUnix(sockType, nil, addr)
	if err != nil {
		stp.t.Fatal(err)
	}

	return listener, conn
}

func (stp *testSetup) snSend(pkt snPkts1.Packet, setMsgID bool) {
	if setMsgID {
		if pkt2, ok := pkt.(snPkts1.PacketWithID); ok {
			pkt2.SetMessageID(stp.snNextMsgID)
			stp.snNextMsgID++
		}
	}

	err := pkt.Write(stp.snConn)
	if err != nil {
		stp.t.Fatal(err)
	}
}

func (stp *testSetup) snRecv() snPkts1.Packet {
	buff := make([]byte, maxTestPktLength)
	n, err := stp.snConn.Read(buff)
	if err != nil {
		if err != io.EOF {
			stp.t.Fatal(err)
		}
	}

	pktReader := bytes.NewReader(buff[:n])
	header := &snPkts.Header{}
	header.Unpack(pktReader)
	pkt := snPkts1.NewPacketWithHeader(*header)
	pkt.Unpack(pktReader)

	return pkt
}

func (stp *testSetup) mqttSend(pkt mqPkts.ControlPacket, setMsgID bool) {
	if setMsgID {
		switch pkt2 := pkt.(type) {
		case *mqPkts.PublishPacket:
			pkt2.MessageID = stp.mqttNextMsgID
		default:
			stp.t.Fatalf("Cannot set MsgID for %v", pkt)
		}
		stp.mqttNextMsgID++
	}

	err := pkt.Write(stp.mqttConn)
	if err != nil {
		stp.t.Fatal(err)
	}
}

func (stp *testSetup) mqttRecv() mqPkts.ControlPacket {
	pkt, err := mqPkts.ReadPacket(stp.mqttConn)
	if err != nil {
		stp.t.Fatal(err)
	}

	return pkt
}

func testRead(connID string, conn net.Conn, timeout time.Duration) ([]byte, error) {
	buff := make([]byte, maxTestPktLength)
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, fmt.Errorf("can't set read deadline on %s connection: %s", connID, err)
	}

	n, err := conn.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff[:n], nil
}

func (stp *testSetup) assertConnEmpty(connID string, conn net.Conn, timeout time.Duration) {
	data, err := testRead(connID, conn, timeout)
	assert.Len(stp.t, data, 0, "No data expected on %s connection, got: %v", connID, data)
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Temporary() && e.Timeout() {
				return
			}
		}
		stp.t.Errorf("Unexpected error on %s connection: %s", connID, err)
	}
}

func (stp *testSetup) assertConnClosed(connID string, conn net.Conn, timeout time.Duration) {
	assert := assert.New(stp.t)

	data, err := testRead(connID, conn, timeout)
	assert.Len(data, 0, "%s connection: no data expected, got: %v", connID, data)
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Temporary() && e.Timeout() {
				stp.t.Errorf("%s connection: not closed", connID)
				return
			}
		}
	}
	assert.Equal(io.EOF, err, "%s connection: expected EOF, got: %s", connID, err)
}

func (stp *testSetup) assertHandlerDone() {
	wg := &sync.WaitGroup{}

	// MQTT-SN connection should be empty but should NOT be closed (closing
	// it is parent's responsibility).
	wg.Add(1)
	go func() {
		defer wg.Done()
		stp.assertConnEmpty("MQTT-SN", stp.snConn, connEmptyTimeout)
	}()

	// MQTT connection should be empty and closed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		stp.assertConnClosed("MQTT", stp.mqttConn, connEmptyTimeout)
	}()

	select {
	case <-time.After(handlerQuitTimeout):
		stp.t.Error("handler did not quit")
	case <-stp.handlerDone:
		// OK
	}

	wg.Wait()
}

//
// Reusable transactions.
//

// Client CONNECT transaction.
func (stp *testSetup) connect() {
	assert := assert.New(stp.t)

	clientID := []byte("test-client")

	// client --CONNECT--> GW
	snConnect := snPkts1.NewConnect(clientID, true, false, 1)
	stp.snSend(snConnect, false)

	// GW --CONNECT--> MQTT broker
	mqttConnect := stp.mqttRecv().(*mqPkts.ConnectPacket)
	assert.Equal(string(snConnect.ClientID), mqttConnect.ClientIdentifier)
	assert.Equal(snConnect.CleanSession, mqttConnect.CleanSession)
	assert.Equal(snConnect.Duration, mqttConnect.Keepalive)
	assert.Equal(byte(4), mqttConnect.ProtocolVersion)
	assert.Equal("MQTT", mqttConnect.ProtocolName)

	// GW <--CONNACK-- MQTT broker
	mqttConnack := mqPkts.NewControlPacket(mqPkts.Connack).(*mqPkts.ConnackPacket)
	mqttConnack.ReturnCode = mqPkts.Accepted
	stp.mqttSend(mqttConnack, false)

	// client <--CONNACK-- GW
	snConnack := stp.snRecv().(*snPkts1.Connack)
	assert.Equal(snPkts1.RC_ACCEPTED, snConnack.ReturnCode)

	assert.Equal(util.StateActive, stp.handler.state.Get())
}

// Client REGISTER transaction.
func (stp *testSetup) register(topic string) uint16 {
	assert := assert.New(stp.t)

	// client --REGISTER--> GW
	snRegister := snPkts1.NewRegister(0, topic)
	snRegister.TopicName = topic
	stp.snSend(snRegister, true)

	// client <--REGACK-- GW
	snRegack := stp.snRecv().(*snPkts1.Regack)
	assert.Equal(snPkts1.RC_ACCEPTED, snRegack.ReturnCode)
	assert.Equal(snRegister.MessageID(), snRegack.MessageID())
	assert.Greater(snRegack.TopicID, uint16(0))

	return snRegack.TopicID
}

// Client SUBSCRIBE transaction.
func (stp *testSetup) subscribe(topic string, qos uint8) uint16 {
	assert := assert.New(stp.t)

	// client --SUBSCRIBE--> GW
	snSubscribe := snPkts1.NewSubscribe(0, snPkts1.TIT_STRING, []byte(topic), qos, false)
	stp.snSend(snSubscribe, true)

	// GW --SUBSCRIBE--> MQTT broker
	mqttSubscribe := stp.mqttRecv().(*mqPkts.SubscribePacket)
	assert.Len(mqttSubscribe.Qoss, 1)
	assert.Equal(snSubscribe.QOS, mqttSubscribe.Qoss[0])
	assert.Len(mqttSubscribe.Topics, 1)
	assert.Equal(topic, mqttSubscribe.Topics[0])

	// GW <--SUBACK-- MQTT broker
	mqttSuback := mqPkts.NewControlPacket(mqPkts.Suback).(*mqPkts.SubackPacket)
	mqttSuback.MessageID = mqttSubscribe.MessageID
	mqttSuback.ReturnCodes = []byte{snSubscribe.QOS}
	stp.mqttSend(mqttSuback, false)

	// client <--SUBACK-- GW
	snSuback := stp.snRecv().(*snPkts1.Suback)
	assert.Equal(snSubscribe.MessageID(), snSuback.MessageID())
	assert.Equal(snPkts1.RC_ACCEPTED, snSuback.ReturnCode)
	if hasWildcard(topic) {
		assert.Equal(uint16(0), snSuback.TopicID)
	} else {
		assert.GreaterOrEqual(snSuback.TopicID, snPkts1.MinTopicID)
		assert.LessOrEqual(snSuback.TopicID, snPkts1.MaxTopicID)
	}

	return snSuback.TopicID
}

func (stp *testSetup) subscribeShort(topic string, qos uint8) {
	assert := assert.New(stp.t)

	assert.True(snPkts1.IsShortTopic(topic))

	// client --SUBSCRIBE--> GW
	topicID := snPkts1.EncodeShortTopic(topic)
	snSubscribe := snPkts1.NewSubscribe(topicID, snPkts1.TIT_SHORT, nil, qos, false)
	stp.snSend(snSubscribe, true)

	// GW --SUBSCRIBE--> MQTT broker
	mqttSubscribe := stp.mqttRecv().(*mqPkts.SubscribePacket)
	assert.Len(mqttSubscribe.Qoss, 1)
	assert.Equal(snSubscribe.QOS, mqttSubscribe.Qoss[0])
	assert.Len(mqttSubscribe.Topics, 1)
	assert.Equal(topic, mqttSubscribe.Topics[0])

	// GW <--SUBACK-- MQTT broker
	mqttSuback := mqPkts.NewControlPacket(mqPkts.Suback).(*mqPkts.SubackPacket)
	mqttSuback.MessageID = mqttSubscribe.MessageID
	mqttSuback.ReturnCodes = []byte{snSubscribe.QOS}
	stp.mqttSend(mqttSuback, false)

	// client <--SUBACK-- GW
	snSuback := stp.snRecv().(*snPkts1.Suback)
	assert.Equal(snSubscribe.MessageID(), snSuback.MessageID())
	assert.Equal(snPkts1.RC_ACCEPTED, snSuback.ReturnCode)
	assert.Equal(snSuback.TopicID, uint16(0))
}

// Client DISCONNECT transaction.
func (stp *testSetup) disconnect() {
	assert := assert.New(stp.t)

	// client --DISCONNECT--> GW
	snDisconnect := snPkts1.NewDisconnect(0)
	stp.snSend(snDisconnect, true)

	// GW --DISCONNECT--> MQTT broker
	mqttDisconnect := stp.mqttRecv().(*mqPkts.DisconnectPacket)
	assert.Equal(uint8(mqPkts.Disconnect), mqttDisconnect.MessageType)

	// client <--DISCONNECT-- GW
	snDisconnectReply := stp.snRecv().(*snPkts1.Disconnect)
	assert.Equal(uint16(0), snDisconnectReply.Duration)

	stp.assertHandlerDone()
}
