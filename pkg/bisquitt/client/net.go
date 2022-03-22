package client

import (
	"context"
	"fmt"
	"net"
	"time"

	msgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/util"

	dtlsProtocol "github.com/pion/dtls/v2/pkg/protocol"
)

func (c *Client) send(msg msgs.Message) error {
	c.log.Debug("<- %v", msg)
	return msg.Write(c.conn)
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
		msg, err := msgs.ReadPacket(c.conn)
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
		c.log.Debug("-> %v", msg)
		if err := c.handlePacket(msg); err != nil {
			return err
		}
	}
}

func (c *Client) topicForPublish(msg *msgs.PublishMessage) (string, error) {
	var topic string
	switch msg.TopicIDType {
	case msgs.TIT_REGISTERED:
		var ok bool
		c.registeredTopicsLock.RLock()
		topic, ok = findTopic(msg.TopicID, c.registeredTopics)
		c.registeredTopicsLock.RUnlock()
		if !ok {
			return "", fmt.Errorf("Invalid topic ID: %d", msg.TopicID)
		}
	case msgs.TIT_PREDEFINED:
		var ok bool
		topic, ok = c.cfg.PredefinedTopics.GetTopicName(c.cfg.ClientID, msg.TopicID)
		if !ok {
			return "", fmt.Errorf("Invalid predefined topic ID: %d", msg.TopicID)
		}
	case msgs.TIT_SHORT:
		topic = msgs.DecodeShortTopic(msg.TopicID)

	default:
		return "", fmt.Errorf("Invalid Topic ID Type: %d", msg.TopicIDType)
	}

	return topic, nil
}

func (c *Client) handlePacket(msgx msgs.Message) error {
	switch msg := msgx.(type) {
	case *msgs.ConnackMessage:
		transactionx, _ := c.transactions.GetByType(msgs.CONNECT)
		transaction, ok := transactionx.(*connectTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Connack(msg)
		return nil

	case *msgs.RegisterMessage:
		c.registeredTopicsLock.Lock()
		// MQTT-SN specification v. 1.2 does not specify what to do if
		// the REGISTER message contains an already registered TopicID.
		// I suppose the right reaction is to reject the registratin with
		// `Rejected: invalid topic ID`.
		var returnCode msgs.ReturnCode
		if _, ok := c.registeredTopics[string(msg.TopicName)]; ok {
			returnCode = msgs.RC_INVALID_TOPIC_ID
		} else {
			returnCode = msgs.RC_ACCEPTED
			c.registeredTopics[string(msg.TopicName)] = msg.TopicID
		}
		c.registeredTopicsLock.Unlock()

		reply := msgs.NewRegackMessage(msg.TopicID, returnCode)
		reply.CopyMessageID(msg)
		return c.send(reply)

	case *msgs.RegackMessage:
		transactionx, _ := c.transactions.Get(msg.MessageID())
		transaction, ok := transactionx.(*registerTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Regack(msg)
		return nil

	case *msgs.SubackMessage:
		transactionx, _ := c.transactions.Get(msg.MessageID())
		transaction, ok := transactionx.(*subscribeTransaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Suback(msg)
		return nil

	// Broker PUBLISH QoS 0,1,2 transaction.
	case *msgs.PublishMessage:
		switch msg.QOS {
		case 0:
			// continue
		case 1:
			puback := msgs.NewPubackMessage(msg.TopicID, msgs.RC_ACCEPTED)
			puback.CopyMessageID(msg)
			if err := c.send(puback); err != nil {
				return err
			}
		case 2:
			var transaction *brokerPublishQOS2Transaction
			transactionx, hasTransaction := c.transactions.Get(msg.MessageID())
			if hasTransaction {
				// We already have such transaction -> resent PUBLISH.
				var ok bool
				transaction, ok = transactionx.(*brokerPublishQOS2Transaction)
				if !ok {
					c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
					return nil
				}
			} else {
				transaction = newBrokerPublishQOS2Transaction(c, msg.MessageID())
				c.transactions.Store(msg.MessageID(), transaction)
			}
			return transaction.Publish(msg)
		default:
			return fmt.Errorf("invalid QOS in %s", msg)
		}
		topic, err := c.topicForPublish(msg)
		if err != nil {
			return err
		}
		c.messageHandlers.handle(c, topic, msg)
		return nil

	// Broker PUBLISH QoS 2 transaction.
	case *msgs.PubrelMessage:
		transactionx, _ := c.transactions.Get(msg.MessageID())
		transaction, ok := transactionx.(*brokerPublishQOS2Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Pubrel(msg)
		return nil

	// Client PUBLISH QoS 1 transaction.
	case *msgs.PubackMessage:
		transactionx, _ := c.transactions.Get(msg.MessageID())
		transaction, ok := transactionx.(*publishQOS1Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Puback(msg)
		return nil

	// Client PUBLISH QoS 2 transaction.
	case *msgs.PubrecMessage:
		transactionx, _ := c.transactions.Get(msg.MessageID())
		transaction, ok := transactionx.(*publishQOS2Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		return transaction.Pubrec(msg)

	// Client PUBLISH QoS 2 transaction.
	case *msgs.PubcompMessage:
		transactionx, _ := c.transactions.Get(msg.MessageID())
		transaction, ok := transactionx.(*publishQOS2Transaction)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Pubcomp(msg)
		return nil

	case *msgs.DisconnectMessage:
		transactionx, ok := c.transactions.GetByType(msgs.DISCONNECT)
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
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Disconnect(msg)
		return nil

	case *msgs.WillTopicReqMessage:
		willTopic := msgs.NewWillTopicMessage(c.cfg.WillTopic, c.cfg.WillQOS, c.cfg.WillRetained)
		return c.send(willTopic)

	case *msgs.WillMsgReqMessage:
		willMsg := msgs.NewWillMsgMessage(c.cfg.WillPayload)
		return c.send(willMsg)

	case *msgs.PingrespMessage:
		transactionx, ok := c.transactions.GetByType(msgs.PINGREQ)
		if !ok {
			// Sleep transaction.
			transactionx, _ = c.transactions.GetByType(msgs.DISCONNECT)
		}
		transaction, ok := transactionx.(transactionWithPingresp)
		if !ok {
			c.log.Error("Unexpected transaction type %T for message: %v", transactionx, msg)
			return nil
		}
		transaction.Pingresp(msg)
		return nil

	default:
		return fmt.Errorf("Unhandled MQTT-SN message: %v", msgx)
	}
}
