package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/topics"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

const (
	maxTestPktLength = 512
	// How long to wait to confirm no other message arrived on the connection.
	connEmptyTimeout = 250 * time.Millisecond
	// How long to wait for client to quit.
	clientQuitTimeout = 2 * connEmptyTimeout
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		stp.connect(clientID)
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestConnectRejected(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// client --CONNECT--> GW
		connect := stp.recv().(*pkts1.Connect)
		assert.Equal(true, connect.CleanSession)
		assert.Equal([]byte(clientID), connect.ClientID)
		assert.Equal(uint16(0), connect.Duration)
		assert.Equal(uint8(1), connect.ProtocolID)
		assert.Equal(false, connect.Will)

		// client <--CONNACK-- GW
		stp.send(pkts1.NewConnack(pkts1.RC_CONGESTION))
	}()

	err := stp.client.Connect()
	assert.Equal("connection rejected: congestion", err.Error())
	assert.Equal(util.StateDisconnected, stp.client.state.Get())

	wg.Wait()
}

func TestConnectRetry(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --CONNECT--> GW
			connect := stp.recv().(*pkts1.Connect)
			assert.Equal(true, connect.CleanSession)
			assert.Equal([]byte(clientID), connect.ClientID)
			assert.Equal(uint16(0), connect.Duration)
			assert.Equal(uint8(1), connect.ProtocolID)
			assert.Equal(false, connect.Will)
		}

		// client <--CONNACK-- GW
		stp.send(pkts1.NewConnack(pkts1.RC_ACCEPTED))

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestConnectAuthWill(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	willTopic := "test/will"
	willPayload := []byte("will-data")
	willQOS := uint8(1)
	willRetained := true
	user := "test-user"
	password := []byte("test-password")

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// client --CONNECT--> GW
		connect := stp.recv().(*pkts1.Connect)
		assert.Equal(true, connect.CleanSession)
		assert.Equal([]byte(clientID), connect.ClientID)
		assert.Equal(uint16(0), connect.Duration)
		assert.Equal(uint8(1), connect.ProtocolID)
		assert.Equal(true, connect.Will)

		// client --AUTH--> GW
		auth := stp.recv().(*pkts1.Auth)
		assert.Equal(pkts1.AUTH_PLAIN, auth.Method)
		authUser, authPassword, err := pkts1.DecodePlain(auth)
		assert.Equal(user, authUser)
		assert.Equal(password, authPassword)
		assert.Nil(err)

		// client <--WILLTOPICREQ-- GW
		stp.send(pkts1.NewWillTopicReq())

		// client --WILLTOPIC--> GW
		willTopicPkt := stp.recv().(*pkts1.WillTopic)
		assert.Equal(willTopic, willTopicPkt.WillTopic)
		assert.Equal(willQOS, willTopicPkt.QOS)
		assert.Equal(willRetained, willTopicPkt.Retain)

		// client <--WILLMSGREQ-- GW
		willMsgReq := pkts1.NewWillMsgReq()
		stp.send(willMsgReq)

		// client --WILLMSG--> GW
		willMsg := stp.recv().(*pkts1.WillMsg)
		assert.Equal(willPayload, willMsg.WillMsg)

		// client <--CONNACK-- GW
		stp.send(pkts1.NewConnack(pkts1.RC_ACCEPTED))

		stp.disconnect()
	}()

	stp.client.cfg.WillTopic = willTopic
	stp.client.cfg.WillPayload = willPayload
	stp.client.cfg.WillQOS = willQOS
	stp.client.cfg.WillRetained = willRetained
	stp.client.cfg.User = user
	stp.client.cfg.Password = password

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestRegister(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	topicID := uint16(123)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --REGISTER--> GW
		register := stp.recv().(*pkts1.Register)
		assert.Equal(topic, register.TopicName)
		assert.Equal(uint16(0), register.TopicID)

		// client <--REGACK-- GW
		regack := pkts1.NewRegack(topicID, pkts1.RC_ACCEPTED)
		regack.CopyMessageID(register)
		stp.send(regack)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Register(topic); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(topicID, stp.client.registeredTopics[topic])

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS0(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(0)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --REGISTER--> GW
		register := stp.recv().(*pkts1.Register)
		assert.Equal(topic, register.TopicName)
		assert.Equal(uint16(0), register.TopicID)

		// client <--REGACK-- GW
		regack := pkts1.NewRegack(topicID, pkts1.RC_ACCEPTED)
		regack.CopyMessageID(register)
		stp.send(regack)

		// client --PUBLISH--> GW
		publish := stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_REGISTERED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Register(topic); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Publish(topic, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS0Predefined(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(0)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --PUBLISH--> GW
		publish := stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_PREDEFINED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.PublishPredefined(topicID, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS0Short(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topic := "ab"
	qos := uint8(0)
	retain := true
	topicID := pkts1.EncodeShortTopic(topic)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --PUBLISH--> GW
		publish := stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_SHORT, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Publish(topic, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS1(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(1)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --REGISTER--> GW
		register := stp.recv().(*pkts1.Register)
		assert.Equal(topic, register.TopicName)
		assert.Equal(uint16(0), register.TopicID)

		// client <--REGACK-- GW
		regack := pkts1.NewRegack(topicID, pkts1.RC_ACCEPTED)
		regack.CopyMessageID(register)
		stp.send(regack)

		var publish *pkts1.Publish
		// client --PUBLISH--> GW
		publish = stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_REGISTERED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
		msgID := publish.MessageID()

		for i := uint(0); i < stp.client.cfg.RetryCount; i++ {
			// client --PUBLISH--> GW
			publish = stp.recv().(*pkts1.Publish)
			assert.Equal(pkts1.TIT_REGISTERED, publish.TopicIDType)
			assert.Equal(topicID, publish.TopicID)
			assert.Equal(payload, publish.Data)
			assert.Equal(retain, publish.Retain)
			assert.Equal(qos, publish.QOS)
			assert.Equal(msgID, publish.MessageID())
		}

		// client <--PUBACK-- GW
		puback := pkts1.NewPuback(publish.TopicID, pkts1.RC_ACCEPTED)
		puback.CopyMessageID(publish)
		stp.send(puback)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Register(topic); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Publish(topic, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS1Predefined(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(1)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		var publish *pkts1.Publish
		// client --PUBLISH--> GW
		publish = stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_PREDEFINED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
		msgID := publish.MessageID()

		for i := uint(0); i < stp.client.cfg.RetryCount; i++ {
			// client --PUBLISH--> GW
			publish = stp.recv().(*pkts1.Publish)
			assert.Equal(pkts1.TIT_PREDEFINED, publish.TopicIDType)
			assert.Equal(topicID, publish.TopicID)
			assert.Equal(payload, publish.Data)
			assert.Equal(retain, publish.Retain)
			assert.Equal(qos, publish.QOS)
			assert.Equal(msgID, publish.MessageID())
		}

		// client <--PUBACK-- GW
		puback := pkts1.NewPuback(publish.TopicID, pkts1.RC_ACCEPTED)
		puback.CopyMessageID(publish)
		stp.send(puback)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.PublishPredefined(topicID, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS1Short(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topic := "ab"
	qos := uint8(1)
	retain := true
	topicID := pkts1.EncodeShortTopic(topic)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		var publish *pkts1.Publish
		// client --PUBLISH--> GW
		publish = stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_SHORT, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
		msgID := publish.MessageID()

		for i := uint(0); i < stp.client.cfg.RetryCount; i++ {
			// client --PUBLISH--> GW
			publish = stp.recv().(*pkts1.Publish)
			assert.Equal(pkts1.TIT_SHORT, publish.TopicIDType)
			assert.Equal(topicID, publish.TopicID)
			assert.Equal(payload, publish.Data)
			assert.Equal(retain, publish.Retain)
			assert.Equal(qos, publish.QOS)
			assert.Equal(msgID, publish.MessageID())
		}

		// client <--PUBACK-- GW
		puback := pkts1.NewPuback(publish.TopicID, pkts1.RC_ACCEPTED)
		puback.CopyMessageID(publish)
		stp.send(puback)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Publish(topic, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS2(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(2)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --REGISTER--> GW
		register := stp.recv().(*pkts1.Register)
		assert.Equal(topic, register.TopicName)
		assert.Equal(uint16(0), register.TopicID)

		// client <--REGACK-- GW
		regack := pkts1.NewRegack(topicID, pkts1.RC_ACCEPTED)
		regack.CopyMessageID(register)
		stp.send(regack)

		var publish *pkts1.Publish
		// client --PUBLISH--> GW
		publish = stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_REGISTERED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
		msgID := publish.MessageID()

		for i := uint(0); i < stp.client.cfg.RetryCount; i++ {
			// client --PUBLISH--> GW
			publish = stp.recv().(*pkts1.Publish)
			assert.Equal(pkts1.TIT_REGISTERED, publish.TopicIDType)
			assert.Equal(topicID, publish.TopicID)
			assert.Equal(payload, publish.Data)
			assert.Equal(retain, publish.Retain)
			assert.Equal(qos, publish.QOS)
			assert.Equal(msgID, publish.MessageID())
		}

		pubrec := pkts1.NewPubrec()
		pubrec.SetMessageID(msgID)
		stp.send(pubrec)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --PUBREL--> GW
			pubrel := stp.recv().(*pkts1.Pubrel)
			assert.Equal(msgID, pubrel.MessageID())
		}

		pubcomp := pkts1.NewPubcomp()
		pubcomp.SetMessageID(msgID)
		stp.send(pubcomp)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Register(topic); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Publish(topic, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS2Predefined(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(2)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		var publish *pkts1.Publish
		// client --PUBLISH--> GW
		publish = stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_PREDEFINED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
		msgID := publish.MessageID()

		for i := uint(0); i < stp.client.cfg.RetryCount; i++ {
			// client --PUBLISH--> GW
			publish = stp.recv().(*pkts1.Publish)
			assert.Equal(pkts1.TIT_PREDEFINED, publish.TopicIDType)
			assert.Equal(topicID, publish.TopicID)
			assert.Equal(payload, publish.Data)
			assert.Equal(retain, publish.Retain)
			assert.Equal(qos, publish.QOS)
			assert.Equal(msgID, publish.MessageID())
		}

		pubrec := pkts1.NewPubrec()
		pubrec.SetMessageID(msgID)
		stp.send(pubrec)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --PUBREL--> GW
			pubrel := stp.recv().(*pkts1.Pubrel)
			assert.Equal(msgID, pubrel.MessageID())
		}

		pubcomp := pkts1.NewPubcomp()
		pubcomp.SetMessageID(msgID)
		stp.send(pubcomp)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.PublishPredefined(topicID, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS2Short(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topic := "ab"
	qos := uint8(2)
	retain := true
	topicID := pkts1.EncodeShortTopic(topic)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		var publish *pkts1.Publish
		// client --PUBLISH--> GW
		publish = stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_SHORT, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
		msgID := publish.MessageID()

		for i := uint(0); i < stp.client.cfg.RetryCount; i++ {
			// client --PUBLISH--> GW
			publish = stp.recv().(*pkts1.Publish)
			assert.Equal(pkts1.TIT_SHORT, publish.TopicIDType)
			assert.Equal(topicID, publish.TopicID)
			assert.Equal(payload, publish.Data)
			assert.Equal(retain, publish.Retain)
			assert.Equal(qos, publish.QOS)
			assert.Equal(msgID, publish.MessageID())
		}

		pubrec := pkts1.NewPubrec()
		pubrec.SetMessageID(msgID)
		stp.send(pubrec)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --PUBREL--> GW
			pubrel := stp.recv().(*pkts1.Pubrel)
			assert.Equal(msgID, pubrel.MessageID())
		}

		pubcomp := pkts1.NewPubcomp()
		pubcomp.SetMessageID(msgID)
		stp.send(pubcomp)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Publish(topic, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPublishQOS3(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	payload := []byte("test/data")
	topicID := uint16(1)
	qos := uint8(3)
	retain := true

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// client --PUBLISH--> GW
		publish := stp.recv().(*pkts1.Publish)
		assert.Equal(pkts1.TIT_PREDEFINED, publish.TopicIDType)
		assert.Equal(topicID, publish.TopicID)
		assert.Equal(payload, publish.Data)
		assert.Equal(retain, publish.Retain)
		assert.Equal(qos, publish.QOS)
	}()

	if err := stp.client.PublishPredefined(topicID, qos, retain, payload); err != nil {
		stp.t.Fatal(err)
	}

	wg.Wait()
}

func TestSubscribeQOS0(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	qos := uint8(0)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		subscribe.QOS = qos
		assert.Equal([]byte(topic), subscribe.TopicName)
		assert.Equal(pkts1.TIT_STRING, subscribe.TopicIDType)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(1, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client <--PUBLISH-- GW
		msgID := uint16(123)
		publish := pkts1.NewPublish(suback.TopicID,
			pkts1.TIT_REGISTERED, []byte(""), qos, false, false)
		publish.SetMessageID(msgID)
		stp.send(publish)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	callbackFired := make(chan struct{})

	callback := func(client *Client, topic string, pubPkt *pkts1.Publish) {
		close(callbackFired)
	}
	if err := stp.client.Subscribe(topic, qos, callback); err != nil {
		stp.t.Fatal(err)
	}

	select {
	case <-callbackFired:
		// ok
	case <-time.After(time.Second):
		stp.t.Fatal("subscribe callback not fired")
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestSubscribeQOS1(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	qos := uint8(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		subscribe.QOS = qos
		assert.Equal([]byte(topic), subscribe.TopicName)
		assert.Equal(pkts1.TIT_STRING, subscribe.TopicIDType)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(1, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client <--PUBLISH-- GW
		msgID := uint16(123)
		publish := pkts1.NewPublish(suback.TopicID,
			pkts1.TIT_REGISTERED, []byte(""), qos, false, false)
		publish.SetMessageID(msgID)
		stp.send(publish)

		// client --PUBACK--> GW
		puback := stp.recv().(*pkts1.Puback)
		assert.Equal(msgID, puback.MessageID())

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	callbackFired := make(chan struct{})

	callback := func(client *Client, topic string, pubPkt *pkts1.Publish) {
		close(callbackFired)
	}
	if err := stp.client.Subscribe(topic, qos, callback); err != nil {
		stp.t.Fatal(err)
	}

	select {
	case <-callbackFired:
		// ok
	case <-time.After(time.Second):
		stp.t.Fatal("subscribe callback not fired")
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestSubscribeQOS2(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	qos := uint8(2)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		subscribe.QOS = qos
		assert.Equal([]byte(topic), subscribe.TopicName)
		assert.Equal(pkts1.TIT_STRING, subscribe.TopicIDType)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(1, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		msgID := uint16(123)
		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client <--PUBLISH-- GW
			publish := pkts1.NewPublish(suback.TopicID,
				pkts1.TIT_REGISTERED, []byte(""), qos, false, false)
			publish.SetMessageID(msgID)
			stp.send(publish)

			// client --PUBREC--> GW
			pubrec := stp.recv().(*pkts1.Pubrec)
			assert.Equal(msgID, pubrec.MessageID())
		}

		// client <--PUBREL-- GW
		pubrel := pkts1.NewPubrel()
		pubrel.SetMessageID(msgID)
		stp.send(pubrel)

		// client --PUBCOMP--> GW
		pubcomp := stp.recv().(*pkts1.Pubcomp)
		assert.Equal(msgID, pubcomp.MessageID())

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	callbackFired := make(chan struct{})

	callback := func(client *Client, topic string, pubPkt *pkts1.Publish) {
		close(callbackFired)
	}
	if err := stp.client.Subscribe(topic, qos, callback); err != nil {
		stp.t.Fatal(err)
	}

	select {
	case <-callbackFired:
		// ok
	case <-time.After(time.Second):
		stp.t.Fatal("subscribe callback not fired")
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestSubscribeWildcard(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	wildcard := "test/+"
	qos := uint8(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	published := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		assert.Equal(pkts1.TIT_STRING, subscribe.TopicIDType)
		assert.Equal([]byte(wildcard), subscribe.TopicName)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(0, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client <--REGISTER-- GW
		msgID := uint16(2)
		register := pkts1.NewRegister(1, topic)
		register.SetMessageID(msgID)
		stp.send(register)

		// client --REGACK--> GW
		regack := stp.recv().(*pkts1.Regack)
		assert.Equal(pkts1.RC_ACCEPTED, regack.ReturnCode)
		assert.Equal(msgID, regack.MessageID())

		// client <--PUBLISH-- GW
		publish := pkts1.NewPublish(register.TopicID,
			pkts1.TIT_REGISTERED, []byte(""), 0, false, false)
		stp.send(publish)

		close(published)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	callbackFired := make(chan struct{})

	callback := func(client *Client, topic string, pubPkt *pkts1.Publish) {
		close(callbackFired)
	}
	if err := stp.client.Subscribe(wildcard, qos, callback); err != nil {
		stp.t.Fatal(err)
	}

	select {
	case <-published:
		// continue
	case <-time.After(time.Second):
		t.Fatal(`timeout waiting for "published"`)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()

	select {
	case <-callbackFired:
		// ok
	case <-time.After(time.Second):
		stp.t.Fatal("subscribe callback not fired")
	}
}

func TestSubscribeShort(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "ab"
	qos := uint8(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		encodedTopic := pkts1.EncodeShortTopic(topic)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		assert.Equal(pkts1.TIT_SHORT, subscribe.TopicIDType)
		assert.Equal(encodedTopic, subscribe.TopicID)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(0, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client <--PUBLISH-- GW
		publish := pkts1.NewPublish(encodedTopic,
			pkts1.TIT_SHORT, []byte(""), 0, false, false)
		stp.send(publish)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	callbackFired := make(chan struct{})

	callback := func(client *Client, topic string, pubPkt *pkts1.Publish) {
		close(callbackFired)
	}
	if err := stp.client.Subscribe(topic, qos, callback); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()

	select {
	case <-callbackFired:
		// ok
	case <-time.After(time.Second):
		stp.t.Fatal("subscribe callback not fired")
	}
}

func TestSubscribePredefined(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	topicID := uint16(1)
	qos := uint8(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()
	stp.client.cfg.PredefinedTopics.Add(clientID, topic, topicID)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		assert.Equal(pkts1.TIT_PREDEFINED, subscribe.TopicIDType)
		assert.Equal(topicID, subscribe.TopicID)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(0, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client <--PUBLISH-- GW
		publish := pkts1.NewPublish(topicID,
			pkts1.TIT_PREDEFINED, []byte(""), 0, false, false)
		stp.send(publish)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	callbackFired := make(chan struct{})

	callback := func(client *Client, topic string, pubPkt *pkts1.Publish) {
		close(callbackFired)
	}
	if err := stp.client.SubscribePredefined(topicID, qos, callback); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()

	select {
	case <-callbackFired:
		// ok
	case <-time.After(time.Second):
		stp.t.Fatal("subscribe callback not fired")
	}
}

func TestUnsubscribeString(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	qos := uint8(0)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		subscribe.QOS = qos
		assert.Equal([]byte(topic), subscribe.TopicName)
		assert.Equal(pkts1.TIT_STRING, subscribe.TopicIDType)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(1, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client --UNSUBSCRIBE--> GW
		unsubscribe := stp.recv().(*pkts1.Unsubscribe)
		assert.Equal([]byte(topic), unsubscribe.TopicName)
		assert.Equal(pkts1.TIT_STRING, unsubscribe.TopicIDType)

		// client <--UNSUBACK-- GW
		unsuback := pkts1.NewUnsuback()
		unsuback.CopyMessageID(unsubscribe)
		stp.send(unsuback)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Subscribe(topic, qos, nil); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Unsubscribe(topic); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestUnsubscribeShort(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "ab"
	qos := uint8(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		encodedTopic := pkts1.EncodeShortTopic(topic)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		assert.Equal(pkts1.TIT_SHORT, subscribe.TopicIDType)
		assert.Equal(encodedTopic, subscribe.TopicID)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(0, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client --UNSUBSCRIBE--> GW
		unsubscribe := stp.recv().(*pkts1.Unsubscribe)
		assert.Equal(pkts1.TIT_SHORT, unsubscribe.TopicIDType)
		assert.Equal(encodedTopic, unsubscribe.TopicID)

		// client <--UNSUBACK-- GW
		unsuback := pkts1.NewUnsuback()
		unsuback.CopyMessageID(unsubscribe)
		stp.send(unsuback)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Subscribe(topic, qos, nil); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Unsubscribe(topic); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestUnsubscribePredefined(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	topic := "test/a"
	topicID := uint16(1)
	qos := uint8(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()
	stp.client.cfg.PredefinedTopics.Add(clientID, topic, topicID)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --SUBSCRIBE--> GW
		subscribe := stp.recv().(*pkts1.Subscribe)
		assert.Equal(pkts1.TIT_PREDEFINED, subscribe.TopicIDType)
		assert.Equal(topicID, subscribe.TopicID)

		// client <--SUBACK-- GW
		suback := pkts1.NewSuback(0, 0, pkts1.RC_ACCEPTED)
		suback.CopyMessageID(subscribe)
		stp.send(suback)

		// client --UNSUBSCRIBE--> GW
		unsubscribe := stp.recv().(*pkts1.Unsubscribe)
		assert.Equal(pkts1.TIT_PREDEFINED, unsubscribe.TopicIDType)
		assert.Equal(topicID, unsubscribe.TopicID)

		// client <--UNSUBACK-- GW
		unsuback := pkts1.NewUnsuback()
		unsuback.CopyMessageID(unsubscribe)
		stp.send(unsuback)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.SubscribePredefined(topicID, qos, nil); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.UnsubscribePredefined(topicID); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestPing(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		// client --PINGREQ--> GW
		pkt := stp.recv()
		_, ok := pkt.(*pkts1.Pingreq)
		assert.True(ok)

		// client <--PINGRESP-- GW
		pingresp := pkts1.NewPingresp()
		stp.send(pingresp)

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Ping(); err != nil {
		stp.t.Fatal(err)
	}

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestSleep(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	sleepSecs := uint16(1)
	numSleeps := 2
	// Must be two characters long.
	topic := "ab"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --DISCONNECT(1)--> GW
			disconnect := stp.recv().(*pkts1.Disconnect)
			assert.Equal(sleepSecs, disconnect.Duration)
		}

		// client <--DISCONNECT(0)-- GW
		stp.send(pkts1.NewDisconnect(0))

		for i := 0; i < numSleeps; i++ {
			sleepStart := time.Now()

			// (sleep)

			// client --PINGREQ--> GW
			_ = stp.recv().(*pkts1.Pingreq)
			sleepDuration := time.Since(sleepStart)
			wantedSleepDuration := time.Duration(sleepSecs) * time.Second
			sleepDiff := sleepDuration - wantedSleepDuration
			// Sleep duration precission < +-1s
			assert.Less(math.Abs(float64(sleepDiff)), float64(time.Second))
			assert.Equal(util.StateAwake, stp.client.state.Get())

			// client <--PUBLISH-- GW
			encodedTopic := pkts1.EncodeShortTopic(topic)
			publish := pkts1.NewPublish(encodedTopic,
				pkts1.TIT_SHORT, []byte(""), 0, false, false)
			stp.send(publish)

			// client <--PINGRESP-- GW
			pingresp := pkts1.NewPingresp()
			stp.send(pingresp)
		}

		stp.connect(clientID)
		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	for i := 0; i < numSleeps; i++ {
		if err := stp.client.Sleep(time.Duration(sleepSecs) * time.Second); err != nil {
			stp.t.Fatal(err)
		}
		assert.Equal(util.StateAwake, stp.client.state.Get())
	}

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestSleepTimeout(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"
	sleepSecs := uint16(1)

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --DISCONNECT(1)--> GW
			disconnect := stp.recv().(*pkts1.Disconnect)
			assert.Equal(sleepSecs, disconnect.Duration)
		}

		stp.disconnect()
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	err := stp.client.Sleep(time.Duration(sleepSecs) * time.Second)
	if err == nil {
		t.Error("Timeout did not occur")
	}
	assert.Equal(transactions.ErrNoMoreRetries, err)

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())
	stp.assertClientDone()

	wg.Wait()
}

func TestDisconnectRetry(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --DISCONNECT--> GW
			disconnect := stp.recv().(*pkts1.Disconnect)
			assert.Equal(uint16(0), disconnect.Duration)
		}

		// client <--DISCONNECT-- GW
		stp.send(pkts1.NewDisconnect(0))
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())

	wg.Wait()
}

func TestDisconnectTimeout(t *testing.T) {
	assert := assert.New(t)

	clientID := "test-client"

	stp := newTestSetup(t, clientID)
	defer stp.cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stp.connect(clientID)

		for i := uint(0); i < stp.client.cfg.RetryCount+1; i++ {
			// client --DISCONNECT--> GW
			disconnect := stp.recv().(*pkts1.Disconnect)
			assert.Equal(uint16(0), disconnect.Duration)
		}
	}()

	if err := stp.client.Connect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateActive, stp.client.state.Get())

	if err := stp.client.Disconnect(); err != nil {
		stp.t.Fatal(err)
	}
	assert.Equal(util.StateDisconnected, stp.client.state.Get())

	wg.Wait()
}

//
// testSetup
//

type testSetup struct {
	ID         string
	t          *testing.T
	conn       net.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	client     *Client
	clientDone chan struct{}
}

func newTestSetup(t *testing.T, clientID string) *testSetup {
	ctx, cancel := context.WithCancel(context.Background())
	clientDone := make(chan struct{})
	// Test name without "Test" prefix.
	id := reflect.ValueOf(*t).FieldByName("name").String()[4:]
	stp := &testSetup{
		ID:         id,
		t:          t,
		ctx:        ctx,
		cancel:     cancel,
		clientDone: clientDone,
	}

	// Create new randomness source and log its seed.
	seed := time.Now().UTC().UnixNano()
	t.Logf("seed = %d\n", seed)
	rand := rand.New(rand.NewSource(seed))

	log := util.NewDebugLogger(stp.ID)
	var listener *net.UnixListener
	listener, stp.conn = stp.createSocketPair("unixpacket", rand)

	cfg := &ClientConfig{
		PredefinedTopics: make(topics.PredefinedTopics),
		CleanSession:     true,
		ClientID:         clientID,
		RetryDelay:       time.Second,
		RetryCount:       2,
		ConnectTimeout:   time.Second,
	}
	stp.client = NewClient(log, cfg)
	stp.client.mockupDialFunc = func() (net.Conn, error) {
		conn, err := listener.AcceptUnix()
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	if err := stp.client.Dial(""); err != nil {
		t.Fatal(err)
	}

	return stp
}

func (stp *testSetup) createSocketPair(sockType string, rand *rand.Rand) (*net.UnixListener, *net.UnixConn) {
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

func (stp *testSetup) send(pkt pkts1.Packet) {
	if err := pkt.Write(stp.conn); err != nil {
		stp.t.Fatal(err)
	}
}

func (stp *testSetup) recv() pkts1.Packet {
	buff := make([]byte, maxTestPktLength)
	n, err := stp.conn.Read(buff)
	if err != nil {
		if err != io.EOF {
			stp.t.Fatal(err)
		}
	}

	pktReader := bytes.NewReader(buff[:n])
	header := &pkts1.Header{}
	header.Unpack(pktReader)
	pkt := pkts1.NewPacketWithHeader(*header)
	pkt.Unpack(pktReader)

	return pkt
}

func testRead(conn net.Conn, timeout time.Duration) ([]byte, error) {
	buff := make([]byte, maxTestPktLength)
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, fmt.Errorf("can't set read deadline on MQTT-SN connection: %s", err)
	}

	n, err := conn.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff[:n], nil
}

func (stp *testSetup) assertConnEmpty(timeout time.Duration) {
	data, err := testRead(stp.client.conn, timeout)
	assert.Len(stp.t, data, 0, "No data expected on MQTT-SN connection, got: %v", data)
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Temporary() && e.Timeout() {
				return
			}
		}
		stp.t.Errorf("Unexpected error on MQTT-SN connection: %s", err)
	}
}

func (stp *testSetup) assertClientDone() {
	wg := &sync.WaitGroup{}

	// MQTT-SN connection should be empty and closed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		stp.assertConnEmpty(connEmptyTimeout)
	}()

	select {
	case <-time.After(clientQuitTimeout):
		stp.t.Error("client did not quit")
	case <-stp.client.groupCtx.Done():
		err := stp.client.groupCtx.Err()
		if err != nil && err != context.Canceled {
			stp.t.Error(err)
		}
	}

	wg.Wait()
}

//
// Reusable transactions.
//

// CONNECT transaction.
func (stp *testSetup) connect(clientID string) {
	assert := assert.New(stp.t)

	// client --CONNECT--> GW
	connect := stp.recv().(*pkts1.Connect)
	assert.Equal(true, connect.CleanSession)
	assert.Equal([]byte(clientID), connect.ClientID)
	assert.Equal(uint16(0), connect.Duration)
	assert.Equal(uint8(1), connect.ProtocolID)
	assert.Equal(false, connect.Will)

	// client <--CONNACK-- GW
	stp.send(pkts1.NewConnack(pkts1.RC_ACCEPTED))
}

// DISCONNECT transaction.
func (stp *testSetup) disconnect() {
	assert := assert.New(stp.t)

	// client --DISCONNECT--> GW
	disconnect := stp.recv().(*pkts1.Disconnect)
	assert.Equal(uint16(0), disconnect.Duration)

	// client <--DISCONNECT-- GW
	stp.send(pkts1.NewDisconnect(0))
}
