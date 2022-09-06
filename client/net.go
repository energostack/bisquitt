package client

import (
	"context"
	"fmt"
	"net"
	"time"

	dtlsProtocol "github.com/pion/dtls/v2/pkg/protocol"

	pkts "github.com/energomonitor/bisquitt/packets"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/util"
)

func (c *Client) send(pkt pkts.Packet) error {
	c.log.Debug("<- %v", pkt)
	buf, err := pkt.Pack()
	if err != nil {
		return err
	}
	_, err = c.conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) keepaliveLoop(ctx context.Context) error {
	c.log.Debug("Keepalive loop starts")
	defer c.log.Debug("Keepalive loop quits")

	// Create and stop a new ticker.
	// It initializes the ticker but prevents it from ticking.
	// The ticker is subsequently reinitialized once the state change is
	// received.
	ticker := time.NewTicker(c.cfg.KeepAlive)
	ticker.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.Ping(); err != nil {
				return err
			}

		case state := <-c.stateChangeCh:
			ticker.Stop()
			if state != util.StateActive {
				continue
			}
			ticker.Reset(c.cfg.KeepAlive)

		case <-ctx.Done():
			return nil
		}
	}
}

func (c *Client) receiveLoop(ctx context.Context) error {
	// Timeout for broker connection read.
	// => ctx.Done() will be checked at least once per this time.
	// => Client will be completely destroyed after at most this time after
	//    ctx is cancelled.
	const readTimeout = 1000 * time.Millisecond

	c.log.Debug("Receive loop starts")
	defer c.log.Debug("Receive loop quits")

	for {
	AGAIN:
		select {
		case <-ctx.Done():
			return nil
		default:
			// continue
		}

		err := c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		if err != nil {
			return err
		}
		pkt, err := pkts1.ReadPacket(c.conn)
		if err != nil {
			switch e := err.(type) {
			case net.Error:
				if e.Temporary() && e.Timeout() {
					goto AGAIN
				}
			case *dtlsProtocol.TimeoutError:
				goto AGAIN
			}
			return err
		}
		c.log.Debug("-> %v", pkt)
		if err := c.handlePacket(pkt); err != nil {
			return err
		}
	}
}

func (c *Client) topicForPublish(pkt *pkts1.Publish) (string, error) {
	var topic string
	switch pkt.TopicIDType {
	case pkts1.TIT_REGISTERED:
		var ok bool
		c.registeredTopicsLock.RLock()
		topic, ok = findTopic(pkt.TopicID, c.registeredTopics)
		c.registeredTopicsLock.RUnlock()
		if !ok {
			return "", fmt.Errorf("invalid topic ID: %d", pkt.TopicID)
		}
	case pkts1.TIT_PREDEFINED:
		var ok bool
		topic, ok = c.cfg.PredefinedTopics.GetTopicName(c.cfg.ClientID, pkt.TopicID)
		if !ok {
			return "", fmt.Errorf("invalid predefined topic ID: %d", pkt.TopicID)
		}
	case pkts1.TIT_SHORT:
		topic = pkts.DecodeShortTopic(pkt.TopicID)

	default:
		return "", fmt.Errorf("invalid Topic ID Type: %d", pkt.TopicIDType)
	}

	return topic, nil
}

func (c *Client) handlePacket(pktx pkts.Packet) error {
	switch pkt := pktx.(type) {
	case *pkts1.Connack:
		transactionx, _ := c.transactions.GetByType(pkts.CONNECT)
		transaction, ok := transactionx.(*connectTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Connack(pkt)
		return nil

	case *pkts1.Register:
		c.registeredTopicsLock.Lock()
		// MQTT-SN specification v. 1.2 does not specify what to do if
		// the REGISTER packet contains an already registered TopicID.
		// I suppose the right reaction is to reject the registratin with
		// `Rejected: invalid topic ID`.
		var returnCode pkts1.ReturnCode
		if _, ok := c.registeredTopics[string(pkt.TopicName)]; ok {
			returnCode = pkts1.RC_INVALID_TOPIC_ID
		} else {
			returnCode = pkts1.RC_ACCEPTED
			c.registeredTopics[string(pkt.TopicName)] = pkt.TopicID
		}
		c.registeredTopicsLock.Unlock()

		reply := pkts1.NewRegack(pkt.TopicID, returnCode)
		reply.CopyMessageID(pkt)
		return c.send(reply)

	case *pkts1.Regack:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*registerTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Regack(pkt)
		return nil

	case *pkts1.Suback:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*subscribeTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Suback(pkt)
		return nil

	case *pkts1.Unsuback:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*unsubscribeTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Unsuback(pkt)
		return nil

	// Broker PUBLISH QoS 0,1,2 transaction.
	case *pkts1.Publish:
		switch pkt.QOS {
		case 0:
			// continue
		case 1:
			puback := pkts1.NewPuback(pkt.TopicID, pkts1.RC_ACCEPTED)
			puback.CopyMessageID(pkt)
			if err := c.send(puback); err != nil {
				return err
			}
		case 2:
			var transaction *brokerPublishQOS2Transaction
			transactionx, hasTransaction := c.transactions.Get(pkt.MessageID())
			if hasTransaction {
				// We already have such transaction -> resent PUBLISH.
				var ok bool
				transaction, ok = transactionx.(*brokerPublishQOS2Transaction)
				if !ok {
					c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
					return nil
				}
			} else {
				transaction = newBrokerPublishQOS2Transaction(c, pkt.MessageID())
				c.transactions.Store(pkt.MessageID(), transaction)
			}
			return transaction.Publish(pkt)
		default:
			return fmt.Errorf("invalid QOS in %s", pkt)
		}
		topic, err := c.topicForPublish(pkt)
		if err != nil {
			return err
		}
		c.messageHandlers.handle(c, topic, pkt)
		return nil

	// Broker PUBLISH QoS 2 transaction.
	case *pkts1.Pubrel:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*brokerPublishQOS2Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Pubrel(pkt)
		return nil

	// Client PUBLISH QoS 1 transaction.
	case *pkts1.Puback:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*publishQOS1Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Puback(pkt)
		return nil

	// Client PUBLISH QoS 2 transaction.
	case *pkts1.Pubrec:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*publishQOS2Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		return transaction.Pubrec(pkt)

	// Client PUBLISH QoS 2 transaction.
	case *pkts1.Pubcomp:
		transactionx, _ := c.transactions.Get(pkt.MessageID())
		transaction, ok := transactionx.(*publishQOS2Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Pubcomp(pkt)
		return nil

	case *pkts1.Disconnect:
		transactionx, ok := c.transactions.GetByType(pkts.DISCONNECT)
		if !ok {
			// Unsolicited DISCONNECT from broker.
			c.setState(util.StateDisconnected)
			// TODO: close connection even when broker does not send DISCONNECT?
			c.log.Debug("Received DISCONNECT, quitting")
			c.cancel()
			return nil
		}

		transaction, ok := transactionx.(transactionWithDisconnect)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Disconnect(pkt)
		return nil

	case *pkts1.WillTopicReq:
		willTopic := pkts1.NewWillTopic(c.cfg.WillTopic, c.cfg.WillQOS, c.cfg.WillRetained)
		return c.send(willTopic)

	case *pkts1.WillMsgReq:
		willMsg := pkts1.NewWillMsg(c.cfg.WillPayload)
		return c.send(willMsg)

	case *pkts1.Pingresp:
		transactionx, ok := c.transactions.GetByType(pkts.PINGREQ)
		if !ok {
			// Sleep transaction.
			transactionx, _ = c.transactions.GetByType(pkts.DISCONNECT)
		}
		transaction, ok := transactionx.(transactionWithPingresp)
		if !ok {
			c.log.Error("Unexpected transaction type %T for packet: %v", transactionx, pkt)
			return nil
		}
		transaction.Pingresp(pkt)
		return nil

	default:
		return fmt.Errorf("unhandled MQTT-SN packet: %v", pktx)
	}
}
