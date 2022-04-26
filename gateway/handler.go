// Handler serves one MQTT-SN client. It is run by the Gateway and receives
// an already-established connection to MQTT-SN client from it.
// After receiving a CONNECT message from the client, it connects to the MQTT broker
// and translates MQTT-SN messages to MQTT ones and vice versa.
//
// Goroutines orchestration is implemented using `errgroup` according to the following rules:
// - All goroutines are spawned using `Handler.group.Go()`.
// - If the goroutine wants to (cleanly) cancel the whole Handler, it returns `Shutdown`.
// - If any of the goroutines returns any error other than `Shutdown`, the Handler
//   is canceled and the error is returned to the Gateway.
// - `Handler.Run()` returns after all goroutines exit.
// - Open connections should be closed at the same level they were opened, i.e.
//   the MQTT-SN connection must be closed by Gateway and the MQTT connection is
//   closed by the Handler.
// - Connections `Read` interruption with context is implemented using `ConnWithContext`.

package gateway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	snMsgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/topics"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"

	mqttPackets "github.com/eclipse/paho.mqtt.golang/packets"
	"golang.org/x/sync/errgroup"
)

type handler struct {
	cfg              *handlerConfig
	id               string
	log              util.Logger
	state            *util.ClientState
	snConn           *util.ConnWithContext
	snRemoteAddr     net.Addr
	mqttConn         *util.ConnWithContext
	registeredTopics sync.Map // uint16 => string
	predefinedTopics topics.PredefinedTopics
	keepAlive        uint16
	clientID         string
	topicID          *util.IDSequence
	msgBuffer        []snMsgs.Message
	group            *errgroup.Group
	transactions     *transactions.TransactionStore
	// for testing
	mockupDialFunc func() net.Conn
}

const (
	// Timeout for MQTT[-SN] connection read.
	// => ctx.Done() will be checked at least once per this time.
	// => Handler will be completely destroyed after at most this time after
	//    ctx is cancelled.
	connTimeout = 100 * time.Millisecond
	// How long to wait for CONNECT transaction to be finished
	connectTransactionTimeout = 5 * time.Second
)

// This error is used to shut down the handler from a goroutine.
// It does not signalize error.
var Shutdown = errors.New("clean shutdown")
var ErrTopicIDsExhausted = errors.New("no more TopicIDs available")
var ErrMqttConnClosed = errors.New("MQTT broker closed connection")
var ErrIllegalMessageWhenDisconnected = errors.New("illegal message in disconnected state")

func hasWildcard(topic string) bool {
	if strings.Contains(topic, "+") {
		return true
	}
	if strings.Contains(topic, "#") {
		return true
	}
	return false
}

type handlerConfig struct {
	MqttBrokerAddress     *net.TCPAddr
	MqttConnectionTimeout time.Duration
	MqttUser              *string
	MqttPassword          []byte
	AuthEnabled           bool
	// TRetry in MQTT-SN specification
	RetryDelay time.Duration
	// NRetry in MQTT-SN specification
	RetryCount uint
}

func newHandler(cfg *handlerConfig, predefinedTopics topics.PredefinedTopics,
	logger util.Logger) *handler {
	state := util.StateDisconnected
	h := &handler{
		cfg:              cfg,
		log:              logger,
		state:            &state,
		predefinedTopics: predefinedTopics,
		topicID:          util.NewIDSequence(snMsgs.MinTopicID, snMsgs.MaxTopicID),
		transactions:     transactions.NewTransactionStore(),
	}

	return h
}

func (h *handler) run(ctx context.Context, snConn net.Conn) error {
	h.log.Debug("Handler starts.")
	defer h.log.Debug("Handler quits.")

	var groupCtx context.Context
	h.group, groupCtx = errgroup.WithContext(ctx)

	// We must create a separate MQTT-SN connection context because we want to
	// send DISCONNECT message when the handler is destroyed => we don't want
	// h.snConn to be closed by groupCtx.
	snCtx, snCancel := context.WithCancel(context.Background())
	h.group.Go(func() error {
		<-groupCtx.Done()

		// The specification doesn't mention when or under what
		// circumstances DISCONNECT should be sent to a client.  Common
		// sense compels us to send the message when the client is either
		// active or awake. The rest of the client states, i.e. ASLEEP,
		// will be handled once we have an actual
		// use case or someone else takes issue, whichever happens
		// first.
		switch h.state.Get() {
		case util.StateActive, util.StateAwake:
			snMsg := snMsgs.NewDisconnectMessage(0)
			if err := h.snSend(snMsg); err != nil {
				h.log.Error("Error sending DISCONNECT to a connection: %s", err)
			}
		}

		snCancel()
		return nil
	})
	h.snConn = util.NewConnWithContext(snCtx, snConn, connTimeout)

	var mqttConn net.Conn
	if h.mockupDialFunc != nil {
		// Used in tests.
		mqttConn = h.mockupDialFunc()
	} else {
		h.log.Debug("Connecting to MQTT broker %s", h.cfg.MqttBrokerAddress.String())
		dialer := &net.Dialer{
			Timeout: h.cfg.MqttConnectionTimeout,
		}
		var err error
		mqttConn, err = dialer.DialContext(ctx, "tcp", h.cfg.MqttBrokerAddress.String())
		if err != nil {
			h.log.Error("Error connecting to MQTT broker: %s", err)
			snMsg := snMsgs.NewConnackMessage(snMsgs.RC_CONGESTION)
			if err := h.snSend(snMsg); err != nil {
				h.log.Error("Error sending CONNACK to a connection: %s", err)
			}
			return err
		}
	}
	h.log.Debug("Connected to MQTT broker")
	defer func() {
		h.log.Debug("Closing MQTT connection")
		if err := mqttConn.Close(); err != nil {
			h.log.Error("Error closing MQTT connection: %s", err)
		}
	}()
	h.mqttConn = util.NewConnWithContext(groupCtx, mqttConn, connTimeout)

	h.group.Go(func() error {
		return h.mqttReceiveLoop(groupCtx)
	})

	h.group.Go(func() error {
		return h.snReceiveLoop(snCtx)
	})

	err := h.group.Wait()
	if err == Shutdown {
		return nil
	}
	if err != nil {
		h.log.Error("Handler quits with error: %v", err)
	}
	return err
}

func (h *handler) setState(new util.ClientState) {
	old := h.state.Set(new)
	if new != old {
		h.log.Debug("State changed to %q.", new)
	}
}

func (h *handler) findRegisteredTopicID(topic string) (topicID uint16, found bool) {
	h.registeredTopics.Range(func(key, value interface{}) bool {
		if value.(string) == topic {
			topicID = key.(uint16)
			found = true
			return false
		}
		return true
	})
	return topicID, found
}

func (h *handler) findTopicID(topic string) (uint16, uint8, bool) {
	topicID, ok := h.findRegisteredTopicID(topic)
	if ok {
		return topicID, snMsgs.TIT_REGISTERED, true
	}
	topicID, ok = h.predefinedTopics.GetTopicID(h.clientID, topic)
	if ok {
		return topicID, snMsgs.TIT_PREDEFINED, true
	}

	return 0, 0, false
}

func (h *handler) handleClientPublish(ctx context.Context, snPublish *snMsgs.PublishMessage) error {
	msgID := snPublish.MessageID()

	mqPublish := mqttPackets.NewControlPacket(mqttPackets.Publish).(*mqttPackets.PublishPacket)
	mqPublish.MessageID = msgID
	mqPublish.Dup = snPublish.DUP()
	if snPublish.QOS == 3 {
		mqPublish.Qos = 0
	} else {
		mqPublish.Qos = snPublish.QOS
	}
	mqPublish.Retain = snPublish.Retain
	var topic string
	switch snPublish.TopicIDType {
	case snMsgs.TIT_REGISTERED:
		topicx, ok := h.registeredTopics.Load(snPublish.TopicID)
		if !ok {
			return fmt.Errorf("Unknown topic id %d", snPublish.TopicID)
		}
		topic = topicx.(string)
	case snMsgs.TIT_PREDEFINED:
		var ok bool
		topic, ok = h.predefinedTopics.GetTopicName(h.clientID, snPublish.TopicID)
		if !ok {
			return fmt.Errorf("Unknown topic id %d", snPublish.TopicID)
		}
	case snMsgs.TIT_SHORT:
		topic = snMsgs.DecodeShortTopic(snPublish.TopicID)
	}
	if snPublish.QOS == 1 {
		h.transactions.Store(msgID, newClientPublishQOS1Transaction(ctx, h, msgID, snPublish.TopicID))
	}
	mqPublish.TopicName = topic
	mqPublish.Payload = snPublish.Data

	return h.mqttSend(mqPublish)
}

func (h *handler) handleBrokerPublish(ctx context.Context, mqPublish *mqttPackets.PublishPacket) error {
	msgID := mqPublish.MessageID

	// Get TopicID
	var needsRegister bool
	var topicID uint16
	var topicIDType uint8
	if snMsgs.IsShortTopic(mqPublish.TopicName) {
		topicID = snMsgs.EncodeShortTopic(mqPublish.TopicName)
		topicIDType = snMsgs.TIT_SHORT
		needsRegister = false
	} else {
		var ok bool
		topicID, topicIDType, ok = h.findTopicID(mqPublish.TopicName)
		needsRegister = !ok
	}

	snPublish := snMsgs.NewPublishMessage(topicID, topicIDType,
		mqPublish.Payload, mqPublish.Qos, mqPublish.Retain, mqPublish.Dup)
	snPublish.SetMessageID(mqPublish.MessageID)

	if mqPublish.Qos == 0 {
		// QOS 0 publish without topic registration does not need a transaction
		if !needsRegister {
			return h.snSend(snPublish)
		}

		// We are reusing PUBLISH message's MsgID because we
		// really want to keep MQTT-SN MsgIDs and MQTT MsgIDs in
		// sync. If we created a new MsgID, we would need
		// to maintain a MQTT-SN MsgID <-> MQTT MsgID map which
		// would be very painful.
		// Reusing the same MsgID is OK because this MsgID would
		// become "available" as REGACK is received and the
		// MsgID can be used freely for the MQTT-SN PUBLISH.
		// The MQTT-SN specification v. 1.2 says nothing about
		// reusing MsgIDs, but MQTT v 5.0 specification says:
		// 	The Packet Identifier becomes available for reuse
		// 	after the sender has processed the corresponding
		// 	acknowledgement packet
		// [chapter 2.2.1 Packet Identifier]
		// But there's a big problem when PUBLISH is QoS 0, i.e.
		// its MsgID is 0. We use a very dirty hack here to choose
		// an "almost surely available" MsgID :(
		found := false
		for i := snMsgs.MaxMessageID; i >= snMsgs.MinMessageID; i-- {
			if _, ok := h.transactions.Get(i); !ok {
				msgID = i
				found = true
				break
			}
		}
		if !found {
			return errors.New("cannot find available MsgID")
		}
	}

	var transaction brokerPublishTransaction
	switch mqPublish.Qos {
	case 0:
		transaction = newBrokerPublishQOS0Transaction(ctx, h, msgID)
	case 1:
		transaction = newBrokerPublishQOS1Transaction(ctx, h, msgID)
	case 2:
		transaction = newBrokerPublishQOS2Transaction(ctx, h, msgID)
	default:
		return fmt.Errorf("Invalid QoS in %v", mqPublish)
	}

	var snMsg snMsgs.Message
	var nextState transactionState
	if needsRegister {
		topicID, err := h.newTopicID()
		if err != nil {
			return err
		}

		// snPublish will be sent after REGACK is received
		snPublish.TopicID = topicID
		transaction.SetSNPublish(snPublish)

		snRegister := snMsgs.NewRegisterMessage(topicID, mqPublish.TopicName)
		snRegister.SetMessageID(msgID)
		nextState = awaitingRegack
		snMsg = snRegister
	} else {
		snMsg = snPublish
		if mqPublish.Qos == 1 {
			nextState = awaitingPuback
		} else {
			// Qos must be 2.
			nextState = awaitingPubrec
		}
	}

	h.transactions.Store(msgID, transaction)
	return transaction.ProceedSN(nextState, snMsg)
}

func (h *handler) handleMqtt(ctx context.Context, msg mqttPackets.ControlPacket) error {
	h.log.Debug("=> %v", msg)
	switch mqMsg := msg.(type) {

	// Client CONNECT transaction.
	case *mqttPackets.ConnackPacket:
		transactionx, _ := h.transactions.GetByType(snMsgs.CONNECT)
		transaction, ok := transactionx.(*connectTransaction)
		if !ok {
			h.log.Error("Unexpected transaction type %T for message: %v", transactionx, mqMsg)
			return nil
		}
		return transaction.Connack(mqMsg)

	// Client PUBLISH QoS 1 transaction.
	case *mqttPackets.PubackPacket:
		transactionx, _ := h.transactions.Get(mqMsg.MessageID)
		transaction, ok := transactionx.(*clientPublishQOS1Transaction)
		if !ok {
			h.log.Error("Unexpected transaction type %T for message: %v", transactionx, mqMsg)
			return nil
		}
		return transaction.Puback(mqMsg)

	// Client PUBLISH QoS 2 transaction.
	case *mqttPackets.PubrecPacket:
		snPubrec := snMsgs.NewPubrecMessage()
		snPubrec.SetMessageID(mqMsg.MessageID)
		return h.snSend(snPubrec)

	// Client PUBLISH QoS 2 transaction.
	case *mqttPackets.PubcompPacket:
		snPubcomp := snMsgs.NewPubcompMessage()
		snPubcomp.SetMessageID(mqMsg.MessageID)
		return h.snSend(snPubcomp)

	// Client SUBSCRIBE transaction.
	case *mqttPackets.SubackPacket:
		transactionx, _ := h.transactions.Get(mqMsg.MessageID)
		transaction, ok := transactionx.(*subscribeTransaction)
		if !ok {
			h.log.Error("Unexpected transaction type %T for message: %v", transactionx, mqMsg)
			return nil
		}
		return transaction.Suback(mqMsg)

	// Client UNSUBSCRIBE transaction.
	case *mqttPackets.UnsubackPacket:
		snUnsuback := snMsgs.NewUnsubackMessage()
		snUnsuback.SetMessageID(mqMsg.MessageID)
		return h.snSend(snUnsuback)

	// Client PING transaction (keepalive).
	case *mqttPackets.PingrespPacket:
		// Response to sleepPinger pings => do not pass to the sleeping client.
		if h.state.Get() != util.StateActive {
			return nil
		}
		return h.snSend(snMsgs.NewPingrespMessage())

	// MQTT broker PUBLISH QOS 0,1,2 transaction.
	case *mqttPackets.PublishPacket:
		return h.handleBrokerPublish(ctx, mqMsg)

	// MQTT broker PUBLISH QoS 2 transaction.
	case *mqttPackets.PubrelPacket:
		transactionx, _ := h.transactions.Get(mqMsg.MessageID)
		transaction, ok := transactionx.(*brokerPublishQOS2Transaction)
		if !ok {
			h.log.Error("Unexpected transaction type %T for message: %v", transactionx, mqMsg)
			return nil
		}
		return transaction.Pubrel(mqMsg)

	default:
		return fmt.Errorf("Unsupported MQTT message type: %v", msg)
	}
}

func (h *handler) snReceiveLoop(ctx context.Context) error {
	h.log.Debug("MQTT-SN receiver starts.")
	defer h.log.Debug("MQTT-SN receiver quits.")
	for {
		msg, err := h.snReceive()
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			h.log.Error("MQTT-SN receive error: %v", err)
			return err
		}
		err = h.handleMqttSn(ctx, msg)
		if err != nil {
			return err
		}
	}
}

func (h *handler) mqttReceiveLoop(ctx context.Context) error {
	h.log.Debug("MQTT receiver starts.")
	defer h.log.Debug("MQTT receiver quits.")
	for {
		msg, err := mqttPackets.ReadPacket(h.mqttConn)
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			if err == io.EOF {
				// Clean shutdown.
				if h.state.Get() == util.StateDisconnected {
					return Shutdown
				}
				h.log.Error("MQTT broker unexpectedly closed connection")
				return ErrMqttConnClosed
			}
			h.log.Error("MQTT decode error: %v", err)
			return err
		}
		if err := h.handleMqtt(ctx, msg); err != nil {
			return err
		}
	}
}

func (h *handler) newTopicID() (uint16, error) {
	topicID, overflow := h.topicID.Next()
	if overflow {
		return 0, ErrTopicIDsExhausted
	}
	for {
		if _, ok := h.predefinedTopics.GetTopicName(h.clientID, topicID); !ok {
			break
		}
		if topicID, overflow = h.topicID.Next(); overflow {
			return 0, ErrTopicIDsExhausted
		}
	}
	return topicID, nil
}

func (h *handler) registerTopic(topic string) (uint16, error) {
	// If already registered, return existing TopicID.
	if topicID, ok := h.findRegisteredTopicID(topic); ok {
		return topicID, nil
	}
	// New registration.
	topicID, err := h.newTopicID()
	if err != nil {
		return 0, err
	}
	h.registeredTopics.Store(topicID, topic)
	return topicID, nil
}

func (h *handler) handleConnect(ctx context.Context, snConnect *snMsgs.ConnectMessage) error {
	// The ProtocolId [...] is coded 0x01. All other values are reserved.
	// MQTT-SN specification v. 1.2, chapter 5.3.8
	if snConnect.ProtocolID != 0x01 {
		reply := &snMsgs.ConnackMessage{
			ReturnCode: snMsgs.RC_NOT_SUPPORTED,
		}
		return h.snSend(reply)
	}

	if h.state.Get() == util.StateAwake {
		h.setState(util.StateActive)
		reply := &snMsgs.ConnackMessage{
			ReturnCode: snMsgs.RC_ACCEPTED,
		}
		return h.snSend(reply)
	}

	// The MQTT-SN specification does not explicitly forbid zero keepalive
	// (meaning "no keepalive" in MQTT) but without keepalive we
	// would not be able to detect lost clients because UDP does not
	// detect lost connection (unlike TCP in MQTT). This would lead
	// to a potential dead Handlers accumulation and such to
	// exploitable memory leaks.
	// Hence, we simply do not accept zero keepalive.
	if snConnect.Duration == 0 {
		reply := &snMsgs.ConnackMessage{
			ReturnCode: snMsgs.RC_NOT_SUPPORTED,
		}
		return h.snSend(reply)
	}

	h.keepAlive = snConnect.Duration
	h.clientID = string(snConnect.ClientID)

	mqConnect := &mqttPackets.ConnectPacket{
		FixedHeader: mqttPackets.FixedHeader{
			MessageType: mqttPackets.Connect,
		},
		ClientIdentifier: h.clientID,
		CleanSession:     snConnect.CleanSession,
		Keepalive:        h.keepAlive,
		ProtocolVersion:  4,
		ProtocolName:     "MQTT",
		UsernameFlag:     h.cfg.MqttUser != nil,
		PasswordFlag:     h.cfg.MqttPassword != nil,
		Password:         h.cfg.MqttPassword,
		WillFlag:         snConnect.Will,
	}
	if mqConnect.UsernameFlag {
		mqConnect.Username = *h.cfg.MqttUser
	}

	// Cancel previous transaction, if any.
	if oldTransaction, ok := h.transactions.GetByType(snMsgs.CONNECT); ok {
		oldTransaction.Fail(Cancelled)
	}
	transaction := newConnectTransaction(ctx, h, h.cfg.AuthEnabled, mqConnect)
	h.transactions.StoreByType(snMsgs.CONNECT, transaction)
	return transaction.Start(ctx)
}

func (h *handler) handleSubscribe(ctx context.Context, snSubscribe *snMsgs.SubscribeMessage) error {
	var topic string
	// From MQTT-SN specification v. 1.2, chapter 5.4.16 SUBACK:
	// 	TopicID [...] [is] not relevant in case of subscriptions to a short topic name or to a topic name which
	// 	contains wildcard characters
	// We will use topicID=0 in such cases. SubackMessage
	var topicID uint16
	switch snSubscribe.TopicIDType {
	case snMsgs.TIT_STRING:
		topic = string(snSubscribe.TopicName)
		if !hasWildcard(topic) {
			var err error
			topicID, err = h.newTopicID()
			if err != nil {
				snSuback := snMsgs.NewSubackMessage(0, 0, snMsgs.RC_INVALID_TOPIC_ID)
				// We are kind of misusing the "invalid topic ID" return code here.
				// Please see note in `case *snMsgs.RegisterMessage`.
				snSuback.CopyMessageID(snSubscribe)
				return h.snSend(snSuback)
			}
			// We must register the topic here, even when we can get
			// a non-successful SUBACK later because MQTT specification says
			// explicitly:
			// The Server is permitted to start sending PUBLISH packets matching
			// the Subscription before the Server sends the SUBACK Packet.
			// [MQTT v.5.0, chapter 3.8.4 SUBSCRIBE Actions]
			h.registeredTopics.Store(topicID, topic)
		}
		// topicID remains zero if client is subscribing to a wildcard topic.
	case snMsgs.TIT_PREDEFINED:
		var ok bool
		topic, ok = h.predefinedTopics.GetTopicName(h.clientID, snSubscribe.TopicID)
		if !ok {
			return fmt.Errorf("Unknown topic id %d", snSubscribe.TopicID)
		}
		topicID = snSubscribe.TopicID
	case snMsgs.TIT_SHORT:
		topic = snMsgs.DecodeShortTopic(snSubscribe.TopicID)
		// topicID remains zero.
	}

	msgID := snSubscribe.MessageID()
	transaction := newSubscribeTransaction(ctx, h, msgID, topicID)
	h.transactions.Store(msgID, transaction)

	mqSubscribe := mqttPackets.NewControlPacket(mqttPackets.Subscribe).(*mqttPackets.SubscribePacket)
	mqSubscribe.MessageID = snSubscribe.MessageID()
	mqSubscribe.Dup = snSubscribe.DUP()
	mqSubscribe.Qoss = []byte{snSubscribe.QOS}
	mqSubscribe.Topics = []string{topic}
	return h.mqttSend(mqSubscribe)
}

func (h *handler) handleUnsubscribe(_ context.Context, snUnsubscribe *snMsgs.UnsubscribeMessage) error {
	var topic string
	switch snUnsubscribe.TopicIDType {
	case snMsgs.TIT_STRING:
		topic = string(snUnsubscribe.TopicName)
	case snMsgs.TIT_PREDEFINED:
		var ok bool
		topic, ok = h.predefinedTopics.GetTopicName(h.clientID, snUnsubscribe.TopicID)
		if !ok {
			return fmt.Errorf("Unknown topic id %d", snUnsubscribe.TopicID)
		}
	case snMsgs.TIT_SHORT:
		topic = snMsgs.DecodeShortTopic(snUnsubscribe.TopicID)
	}

	mqUnsubscribe := mqttPackets.NewControlPacket(mqttPackets.Unsubscribe).(*mqttPackets.UnsubscribePacket)
	mqUnsubscribe.MessageID = snUnsubscribe.MessageID()
	mqUnsubscribe.Topics = []string{topic}
	return h.mqttSend(mqUnsubscribe)
}

// Check whether the given message is legal in the current Handler's state.
//
// We check only messages received in the "disconnected" state because:
// 1. It is the only state before authentication, hence illegal messages could
//    potentially be used to attack the gateway by an unauthenticated user.
// 2. Messages in other states are correctly handled by their respective
//    transactions. Also unexpected messages can be caused by delayed UDP
//    packets etc. therefore we do not want to close the connection
//    when such a message appears.
func (h *handler) checkMessageLegal(msg snMsgs.Message) error {
	state := h.state.Get()
	if state != util.StateDisconnected {
		return nil
	}

	switch snMsg := msg.(type) {
	case *snMsgs.ConnectMessage:
		return nil
	case *snMsgs.AuthMessage:
		return nil
	case *snMsgs.WillMsgMessage:
		return nil
	case *snMsgs.WillTopicMessage:
		return nil
	// Handler is switched to disconnected state _before_ client
	// responds to DISCONNECT => we must enable DISCONNECT message.
	case *snMsgs.DisconnectMessage:
		return nil
	case *snMsgs.PublishMessage:
		// MQTT-SN specification v. 1.2, chapter 6.8 PUBLISH with QoS Level -1
		// QOS 3 messages with short or predefined topics are allowed
		// without prior CONNECT.
		// We do not allow these messages when authentication is enabled.
		if !h.cfg.AuthEnabled &&
			snMsg.QOS == 3 &&
			(snMsg.TopicIDType == snMsgs.TIT_SHORT ||
				snMsg.TopicIDType == snMsgs.TIT_PREDEFINED) {
			return nil
		}
	}

	h.log.Error("Illegal message in %q state: %v", state, msg)
	return ErrIllegalMessageWhenDisconnected
}

func (h *handler) handleMqttSn(ctx context.Context, msg snMsgs.Message) error {
	if err := h.checkMessageLegal(msg); err != nil {
		return err
	}

	switch snMsg := msg.(type) {

	// Client CONNECT transaction.
	case *snMsgs.ConnectMessage:
		return h.handleConnect(ctx, snMsg)

	// Client CONNECT transaction.
	case *snMsgs.AuthMessage:
		transactionx, _ := h.transactions.GetByType(snMsgs.CONNECT)
		if transaction, ok := transactionx.(*connectTransaction); ok {
			return transaction.Auth(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	// Client CONNECT transaction.
	case *snMsgs.WillTopicMessage:
		transactionx, _ := h.transactions.GetByType(snMsgs.CONNECT)
		if transaction, ok := transactionx.(*connectTransaction); ok {
			return transaction.WillTopic(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	// Client CONNECT transaction.
	case *snMsgs.WillMsgMessage:
		transactionx, _ := h.transactions.GetByType(snMsgs.CONNECT)
		if transaction, ok := transactionx.(*connectTransaction); ok {
			return transaction.WillMsg(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	// Client REGISTER transaction.
	case *snMsgs.RegisterMessage:
		returnCode := snMsgs.RC_ACCEPTED
		topicID, err := h.registerTopic(string(snMsg.TopicName))
		if err != nil {
			// The only reason registerTopic can return an error is when all
			// the available TopicIDs are already used. The MQTT-SN specification
			// does not define what the MQTT-SN broker should do in such a situation.
			// This should not happen in real life because no client would typically
			// use 65535 different topics... Nevertheless, the client did not
			// break any rules, so we should not just drop the connection.
			// In MQTT-SN spec 1.2, there's no suitable return code to use,
			// so we decided to use "invalid topic ID" as the least of all evils.
			// Hopefully, future specification versions will define what is the right
			// thing to do.
			returnCode = snMsgs.RC_INVALID_TOPIC_ID
		}
		m2 := snMsgs.NewRegackMessage(topicID, returnCode)
		m2.CopyMessageID(snMsg)
		return h.snSend(m2)

	// Client PUBLISH QoS 0,1,2,3 transaction.
	case *snMsgs.PublishMessage:
		return h.handleClientPublish(ctx, snMsg)

	// Client PUBLISH QoS 2 transaction.
	case *snMsgs.PubrelMessage:
		mqPubrel := mqttPackets.NewControlPacket(mqttPackets.Pubrel).(*mqttPackets.PubrelPacket)
		mqPubrel.MessageID = snMsg.MessageID()
		return h.mqttSend(mqPubrel)

	// Client SUBSCRIBE transaction.
	case *snMsgs.SubscribeMessage:
		return h.handleSubscribe(ctx, snMsg)

	// Client UNSUBSCRIBE transaction.
	case *snMsgs.UnsubscribeMessage:
		return h.handleUnsubscribe(ctx, snMsg)

	// Client PING transaction (going AWAKE or just a keepalive).
	case *snMsgs.PingreqMessage:
		if h.state.Get() == util.StateAsleep {
			// Must be set before snSend otherwise the messages will be queued...
			h.setState(util.StateAwake)
			for _, m2 := range h.msgBuffer {
				if err := h.snSend(m2); err != nil {
					return err
				}
			}
			h.msgBuffer = nil
			return h.snSend(snMsgs.NewPingrespMessage())
		} else {
			mqMsg := mqttPackets.NewControlPacket(mqttPackets.Pingreq).(*mqttPackets.PingreqPacket)
			return h.mqttSend(mqMsg)
		}

	// Client DISCONNECT transaction.
	case *snMsgs.DisconnectMessage:
		if snMsg.Duration == 0 {
			mqMsg := mqttPackets.NewControlPacket(mqttPackets.Disconnect).(*mqttPackets.DisconnectPacket)
			h.mqttSend(mqMsg)
			h.setState(util.StateDisconnected)
			m3 := snMsgs.NewDisconnectMessage(0)
			if err := h.snSend(m3); err != nil {
				return err
			}
			return Shutdown
		} else {
			h.log.Debug("Going to sleep for %vs", snMsg.Duration)
			if h.keepAlive != 0 && snMsg.Duration > h.keepAlive {
				// We must ensure MQTT gateway considers client alive during sleep period.
				cancelPinger := h.startSleepPinger(ctx)
				time.AfterFunc(time.Duration(snMsg.Duration)*time.Second, cancelPinger)
			}
			h.msgBuffer = nil
			m2 := snMsgs.NewDisconnectMessage(0)
			if err := h.snSend(m2); err != nil {
				return err
			}
			// Must be set after snSend otherwise the message will be queued...
			h.setState(util.StateAsleep)
			return nil
		}

	// MQTT-SN gateway REGISTER transaction.
	// Reply to REGISTER messages sent from the gateway to a client. The client
	// subscribes for a wildcard topic, a mqtt broker sends a "response" PUBLISH
	// message with an unregistered topic => the gateway initializes
	// registration and the client must acknowledge it.
	case *snMsgs.RegackMessage:
		transactionx, _ := h.transactions.Get(snMsg.MessageID())
		if transaction, ok := transactionx.(transactionWithRegack); ok {
			return transaction.Regack(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	// MQTT broker PUBLISH QoS 1 transaction.
	case *snMsgs.PubackMessage:
		transactionx, _ := h.transactions.Get(snMsg.MessageID())
		if transaction, ok := transactionx.(*brokerPublishQOS1Transaction); ok {
			return transaction.Puback(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	// MQTT broker PUBLISH QoS 2 transaction.
	case *snMsgs.PubrecMessage:
		transactionx, _ := h.transactions.Get(snMsg.MessageID())
		if transaction, ok := transactionx.(*brokerPublishQOS2Transaction); ok {
			return transaction.Pubrec(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	// MQTT broker PUBLISH QoS 2 transaction.
	case *snMsgs.PubcompMessage:
		transactionx, _ := h.transactions.Get(snMsg.MessageID())
		if transaction, ok := transactionx.(*brokerPublishQOS2Transaction); ok {
			return transaction.Pubcomp(snMsg)
		}
		h.log.Error("Unexpected transaction type %T for message: %v", transactionx, snMsg)
		return nil

	default:
		return fmt.Errorf("Unsupported MQTT-SN message type: %v", msg)
	}
}

func (h *handler) startSleepPinger(ctx context.Context) context.CancelFunc {
	ctx2, cancel := context.WithCancel(ctx)
	h.group.Go(func() error {
		h.log.Debug("Sleep pinger starts.")
		defer h.log.Debug("Sleep pinger quits.")
		for {
			select {
			case <-time.After(time.Duration(h.keepAlive) * time.Second):
				m := mqttPackets.NewControlPacket(mqttPackets.Pingreq).(*mqttPackets.PingreqPacket)
				if err := h.mqttSend(m); err != nil {
					return err
				}
			case <-ctx2.Done():
				return nil
			}
		}
	})
	return cancel
}

func (h *handler) snSend(msg snMsgs.Message) error {
	if h.state.Get() == util.StateAsleep {
		h.log.Debug("Queued %v", msg)
		h.msgBuffer = append(h.msgBuffer, msg)
		// TODO: Potentional serialization errors will be delayed!
		return nil
	}
	h.log.Debug("<- %v", msg)
	err := msg.Write(h.snConn)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) snReceive() (snMsgs.Message, error) {
	// TODO: make static...
	buffer := make([]byte, snMsgs.MaxPacketLen)

	// TODO: Here, we rely on the assumption that we always read precissely one
	// whole packet. This is not guaranteed in the pion/dtls API documentation.
	n, err := h.snConn.Read(buffer)
	if err != nil {
		return nil, err
	}

	pkt := buffer[:n]

	if len(pkt) < 2 {
		return nil, errors.New("Illegal packet: too short")
	}

	pktReader := bytes.NewReader(pkt)
	header := &snMsgs.Header{}
	header.Unpack(pktReader)
	msg := snMsgs.NewMessageWithHeader(*header)
	msg.Unpack(pktReader)

	h.log.Debug("-> %v", msg)
	return msg, nil
}

func (h *handler) mqttSend(msg mqttPackets.ControlPacket) error {
	h.log.Debug("<= %v", msg)
	buff := &bytes.Buffer{}
	err := msg.Write(buff)
	if err != nil {
		return err
	}
	_, err = h.mqttConn.Write(buff.Bytes())
	if err != nil {
		return err
	}
	return nil
}
