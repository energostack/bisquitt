// Package client implements a MQTT-SN version 1.2 client with optional DTLS
// encryption.
//
// Use ClientConfig struct to set various client's options and features.
//
// Example:
//	import (
//		"fmt"
//		"time"
//
//		"github.com/energomonitor/bisquitt/client"
//		"github.com/energomonitor/bisquitt/util"
//	)
//
//	func main() {
//		brokerAddress := "localhost:1883"
//		topic := "dev/test"
//		payload := "test message"
//		qos := 1
//		retain := false
//
//		clientCfg := &client.ClientConfig{
//			ClientID:       "test-client",
//			RetryDelay:     10 * time.Second,
//			RetryCount:     4,
//			ConnectTimeout: 20 * time.Second,
//			KeepAlive:      60 * time.Second,
//			CleanSession:   true,
//		}
//		logger := util.NewDebugLogger("test")
//		c := client.NewClient(logger, clientCfg)
//
//		fmt.Printf("Connecting to a MQTT-SN broker %#v\n", brokerAddress)
//		if err := c.Dial(brokerAddress); err != nil {
//			panic(err)
//		}
//		if err := c.Connect(); err != nil {
//			panic(err)
//		}
//		defer c.Disconnect()
//
//		fmt.Printf("Registering topic %#v\n", topic)
//		if err := c.Register(topic); err != nil {
//			panic(err)
//		}
//
//		fmt.Printf("Publishing: %s <- %s\n", topic, string(payload))
//		if err := c.Publish(topic, uint8(qos), retain, []byte(payload)); err != nil {
//			panic(err)
//		}
//
//		fmt.Println("Everything OK")
//	}
package client

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/pkg/crypto/selfsign"
	"golang.org/x/sync/errgroup"

	pkts "github.com/energomonitor/bisquitt/packets"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/topics"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

type ClientConfig struct {
	// UseDTLS controls whether DTLS should be used to secure the connection
	// to the MQTT-SN gateway.
	UseDTLS        bool
	Certificate    *tls.Certificate
	PrivateKey     crypto.PrivateKey
	CACertificates []*x509.Certificate
	// SelfSigned controls whether the client should use a self-signed
	// certificate and key.  If SelfSigned is false and UseDTLS is true, you
	// must provide CertFile and KeyFile.
	SelfSigned bool
	// Insecure controls whether the client verifies server's certificate.
	// If Insecure is true, the client accepts any certificate and is
	// susceptible to a man-in-the-middle attack. Should be used for testing
	// only.
	Insecure bool
	ClientID string
	// User is used to authenticate with the MQTT-SN gateway.
	User string
	// Password is used to authenticate with the MQTT-SN gateway.
	Password         []byte
	CleanSession     bool
	WillTopic        string
	WillPayload      []byte
	WillQOS          uint8
	WillRetained     bool
	KeepAlive        time.Duration
	ConnectTimeout   time.Duration
	PredefinedTopics topics.PredefinedTopics
	// TRetry in MQTT-SN specification
	RetryDelay time.Duration
	// NRetry in MQTT-SN specification
	RetryCount uint
}

type Client struct {
	cfg                  *ClientConfig
	registeredTopics     map[string]uint16
	registeredTopicsLock sync.RWMutex
	messageHandlers      *messageHandlers
	transactions         *transactions.TransactionStore
	msgID                *util.IDSequence
	conn                 net.Conn
	state                *util.ClientState
	stateChangeCh        chan util.ClientState
	group                *errgroup.Group
	groupCtx             context.Context
	cancel               func()
	log                  util.Logger
	// for testing
	mockupDialFunc func() (net.Conn, error)
}

// NewClient sets up a new client according to the provided configuration.
func NewClient(log util.Logger, cfg *ClientConfig) *Client {
	state := util.StateDisconnected
	return &Client{
		cfg:              cfg,
		registeredTopics: make(map[string]uint16),
		messageHandlers:  &messageHandlers{},
		transactions:     transactions.NewTransactionStore(),
		state:            &state,
		stateChangeCh:    make(chan util.ClientState, 1),
		log:              log,
		msgID:            util.NewIDSequence(pkts.MinPacketID, pkts.MaxPacketID),
	}
}

func (c *Client) connectDTLS(ctx context.Context, address string) (net.Conn, error) {
	var certificate *tls.Certificate
	var err error

	if c.cfg.SelfSigned {
		var cert tls.Certificate
		cert, err = selfsign.GenerateSelfSigned()
		certificate = &cert
	} else {
		privateKey := c.cfg.PrivateKey
		if privateKey == nil {
			err = errors.New("private key is missing")
		}
		if certificate = c.cfg.Certificate; certificate != nil {
			certificate.PrivateKey = privateKey
		} else {
			err = errors.New("TLS certificate is missing")
		}
	}
	if err != nil {
		return nil, err
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	for _, cert := range c.cfg.CACertificates {
		certPool.AddCert(cert)
	}

	hostAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	// Prepare the configuration of the DTLS connection
	config := &dtls.Config{
		Certificates:         []tls.Certificate{*certificate},
		InsecureSkipVerify:   c.cfg.Insecure,
		ExtendedMasterSecret: dtls.RequireExtendedMasterSecret,
		RootCAs:              certPool,
	}

	// Connect to a DTLS server
	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	conn, err := dtls.DialWithContext(ctx2, "udp", hostAddress, config)
	if err != nil {
		c.log.Error("Connect error: %v", err)
		return nil, err
	}

	return conn, nil
}

func findTopic(topicID uint16, m map[string]uint16) (string, bool) {
	for topic, topicID2 := range m {
		if topicID2 == topicID {
			return topic, true
		}
	}
	return "", false
}

// Dial connects to a MQTT-SN broker. The address must be in the "host:port" form.
func (c *Client) Dial(address string) error {
	ctx, cancel := context.WithCancel(context.Background())
	group, groupCtx := errgroup.WithContext(ctx)

	c.cancel = cancel
	c.group = group
	c.groupCtx = groupCtx

	var err error
	if c.mockupDialFunc == nil {
		if c.cfg.UseDTLS {
			c.conn, err = c.connectDTLS(ctx, address)
		} else {
			c.conn, err = net.Dial("udp", address)
		}
		if err != nil {
			return err
		}

		if c.cfg.UseDTLS {
			c.log.Debug("DTLS connected")
		} else {
			c.log.Debug("UDP connected")
		}
	} else {
		c.conn, err = c.mockupDialFunc()
		if err != nil {
			return err
		}
	}

	group.Go(func() error {
		return c.receiveLoop(groupCtx)
	})
	if c.cfg.KeepAlive != 0 {
		group.Go(func() error {
			return c.keepaliveLoop(groupCtx)
		})
	} else {
		c.log.Debug("Keep-Alive loop disabled. Set KeepAlive > 0 to enable it.")
	}

	return nil
}

// Wait blocks until the client is terminated.
func (c *Client) Wait() error {
	return c.group.Wait()
}

// Close closes the connection with the MQTT-SN gateway. The client sends
// a DISCONNECT packet before closing the connection.
func (c *Client) Close() error {
	if err := c.Disconnect(); err != nil {
		return err
	}
	c.cancel()
	return c.conn.Close()
}

func (c *Client) setState(new util.ClientState) {
	old := c.state.Set(new)
	if new == old {
		return
	}
	c.log.Debug("State changed to %q.", new)
	c.notifyStateChange(new)
}

// notifyStateChange sends last state change to a channel. The channel is read
// by keep-alive goroutine which uses the new state to schedule the next issue
// of keep-alive packet.
func (c *Client) notifyStateChange(s util.ClientState) {
	if c.cfg.KeepAlive == 0 {
		return
	}
	select {
	case c.stateChangeCh <- s:
	case <-c.groupCtx.Done():
		return
	}
}

// Connect sends a CONNECT packet to the MQTT-SN gateway. According to the
// MQTT-SN specification, this must be the first packet the client sends
// unless it's a PUBLISH packet with QoS = -1.
func (c *Client) Connect() error {
	connect := pkts1.NewConnect(
		uint16(c.cfg.KeepAlive.Seconds()),
		[]byte(c.cfg.ClientID),
		c.cfg.WillTopic != "",
		c.cfg.CleanSession,
	)

	var auth *pkts1.Auth
	if c.cfg.User != "" {
		auth = pkts1.NewAuthPlain(c.cfg.User, c.cfg.Password)
	}

	for i := uint(0); i < c.cfg.RetryCount+1; i++ {
		transaction := newConnectTransaction(c.groupCtx, c)
		c.transactions.StoreByType(pkts.CONNECT, transaction)

		if err := c.send(connect); err != nil {
			return err
		}
		if auth != nil {
			if err := c.send(auth); err != nil {
				return err
			}
		}

		select {
		case <-transaction.Done():
			err := transaction.Err()
			switch err {
			case nil:
				return nil
			case transactions.ErrTimeout:
				continue
			default:
				return err
			}
		case <-c.groupCtx.Done():
			return c.group.Wait()
		}
	}

	return errors.New("connect timeout")
}

// Register sends a REGISTER packet to the MQTT-SN gateway.
func (c *Client) Register(topic string) error {
	msgID, _ := c.msgID.Next()
	transaction := newRegisterTransaction(c, msgID, topic)
	register := pkts1.NewRegister(0, topic)
	register.SetMessageID(msgID)
	c.transactions.Store(msgID, transaction)
	transaction.Proceed(nil, register)
	if err := c.send(register); err != nil {
		transaction.Fail(err)
	}
	select {
	case <-transaction.Done():
		return transaction.Err()
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}

func (c *Client) subscribe(topicName string, topicIDType uint8, topicID uint16, qos uint8, callback MessageHandlerFunc) error {
	msgID, _ := c.msgID.Next()
	transaction := newSubscribeTransaction(c, msgID, callback)
	subscribe := pkts1.NewSubscribe(topicName, topicID, false, qos, topicIDType)
	subscribe.SetMessageID(msgID)
	c.transactions.Store(msgID, transaction)
	transaction.Proceed(nil, subscribe)
	if err := c.send(subscribe); err != nil {
		transaction.Fail(err)
	}
	select {
	case <-transaction.Done():
		return transaction.Err()
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}

// Subscribe subscribes to a topic with the provided QoS. If the topic is 2 characters
// long, it's treated as a short topic. The received packets are passed to the
// provided callback.
func (c *Client) Subscribe(topic string, qos uint8, callback MessageHandlerFunc) error {
	if pkts.IsShortTopic(topic) {
		return c.subscribe("", pkts1.TIT_SHORT, pkts.EncodeShortTopic(topic), qos, callback)
	} else {
		return c.subscribe(topic, pkts1.TIT_STRING, 0, qos, callback)
	}
}

// SubscribePredefined subscribes to a predefined topic with the provided QoS.
// The received packets are passed to the provided callback.
func (c *Client) SubscribePredefined(topicID uint16, qos uint8, callback MessageHandlerFunc) error {
	return c.subscribe("", pkts1.TIT_PREDEFINED, topicID, qos, callback)
}

func (c *Client) unsubscribe(topicName string, topicIDType uint8, topicID uint16) error {
	msgID, _ := c.msgID.Next()
	transaction := newUnsubscribeTransaction(c, msgID)
	unsubscribe := pkts1.NewUnsubscribe(topicName, topicID, topicIDType)
	unsubscribe.SetMessageID(msgID)
	c.transactions.Store(msgID, transaction)
	transaction.Proceed(nil, unsubscribe)
	if err := c.send(unsubscribe); err != nil {
		transaction.Fail(err)
	}
	select {
	case <-transaction.Done():
		return transaction.Err()
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}

// Unsubscribe unsubscribes from a topic. If the topic is 2 characters long,
// it's treated as a short topic.
func (c *Client) Unsubscribe(topic string) error {
	if pkts.IsShortTopic(topic) {
		return c.unsubscribe("", pkts1.TIT_SHORT, pkts.EncodeShortTopic(topic))
	} else {
		return c.unsubscribe(topic, pkts1.TIT_STRING, 0)
	}
}

// UnsubscribePredefined unsubscribes from a predefined topic.
func (c *Client) UnsubscribePredefined(topicID uint16) error {
	return c.unsubscribe("", pkts1.TIT_PREDEFINED, topicID)
}

func (c *Client) publish(topicIDType uint8, topicID uint16, qos uint8, retain bool, payload []byte) error {
	publish := pkts1.NewPublish(topicID, payload, false, qos, retain, topicIDType)
	msgID, _ := c.msgID.Next()
	publish.SetMessageID(msgID)

	var transaction transactions.StatefulTransaction
	switch qos {
	case 0, 3:
		// no transaction
	case 1:
		transaction = newPublishQOS1Transaction(c, msgID)
		transaction.Proceed(awaitingPuback, publish)
	case 2:
		transaction = newPublishQOS2Transaction(c, msgID)
		transaction.Proceed(awaitingPubrec, publish)
	default:
		return fmt.Errorf("invalid qos: %d", qos)
	}
	if transaction == nil {
		return c.send(publish)
	}

	c.transactions.Store(msgID, transaction)
	if err := c.send(publish); err != nil {
		transaction.Fail(err)
	}
	select {
	case <-transaction.Done():
		return transaction.Err()
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}

// Publish publishes a message to the provided topic.
func (c *Client) Publish(topic string, qos uint8, retain bool, payload []byte) error {
	var topicIDType uint8
	var topicID uint16
	if pkts.IsShortTopic(topic) {
		topicIDType = pkts1.TIT_SHORT
		topicID = pkts.EncodeShortTopic(topic)
	} else {
		topicIDType = pkts1.TIT_REGISTERED
		var ok bool
		c.registeredTopicsLock.RLock()
		topicID, ok = c.registeredTopics[topic]
		c.registeredTopicsLock.RUnlock()
		if !ok {
			return fmt.Errorf("topic %#v not registered!", topic)
		}
	}
	return c.publish(topicIDType, topicID, qos, retain, payload)
}

// PublishPredefined publishes a message to the provided predefined topic.
func (c *Client) PublishPredefined(topicID uint16, qos uint8, retain bool, payload []byte) error {
	return c.publish(pkts1.TIT_PREDEFINED, topicID, qos, retain, payload)
}

// Ping sends a PING packet to the MQTT-SN gateway.
func (c *Client) Ping() error {
	transaction := newPingTransaction(c)
	ping := pkts1.NewPingreq(nil)
	c.transactions.StoreByType(pkts.PINGREQ, transaction)
	transaction.Proceed(nil, ping)
	if err := c.send(ping); err != nil {
		transaction.Fail(err)
	}
	select {
	case <-transaction.Done():
		return transaction.Err()
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}

// Sleep informs the MQTT-SN gateway that the client is going to sleep.
func (c *Client) Sleep(duration time.Duration) error {
	transaction := newSleepTransaction(c, duration)
	c.transactions.StoreByType(pkts.DISCONNECT, transaction)
	if err := transaction.Sleep(); err != nil {
		return err
	}
	select {
	case <-transaction.Done():
		return transaction.Err()
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}

// Disconnect sends a DISCONNECT packet to the MQTT-SN gateway.
// According to the "Client's state transition diagram" in the MQTT-SN
// specification v. 1.2, chapter 6.14, the client can send a DISCONNECT
// packet only in an ACTIVE or AWAKE state.
// This function does not return error if the client is in other states
// so it's usable unconditionally in defer.
func (c *Client) Disconnect() error {
	currentState := c.state.Get()
	if currentState != util.StateActive && currentState != util.StateAwake {
		return nil
	}
	transaction := newDisconnectTransaction(c)
	disconnect := pkts1.NewDisconnect(0)
	c.transactions.StoreByType(pkts.DISCONNECT, transaction)
	transaction.Proceed(awaitingDisconnect, disconnect)
	if err := c.send(disconnect); err != nil {
		transaction.Fail(err)
		return err
	}
	c.setState(util.StateDisconnected)
	select {
	case <-transaction.Done():
		err := transaction.Err()
		switch err {
		case nil:
			c.log.Debug("DISCONNECT ACKed, quitting.")
		case transactions.ErrNoMoreRetries:
			c.log.Info("No reply for DISCONNECT packet from broker, quitting anyway.")
		default:
			return err
		}
		c.cancel()
		return nil
	case <-c.groupCtx.Done():
		return c.group.Wait()
	}
}
