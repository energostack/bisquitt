// The MQTT broker watches the connection state using keepalive (PING* packets)
// _after_ the connection is established (after a CONNECT packet is received by
// the broker).  Before the connection is established, the MQTT-SN gateway
// must watch the connection itself because a malicious client could leave the
// connection half-established (=> possible DoS attack vulnerability).
// Hence, we must use time-limited connectTransaction.

package gateway

import (
	"context"
	"errors"
	"fmt"

	mqPkts "github.com/eclipse/paho.mqtt.golang/packets"

	snPkts "github.com/energostack/bisquitt/packets"
	snPkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
	"github.com/energostack/bisquitt/util"
)

var Cancelled = errors.New("transaction cancelled")

type connectTransaction struct {
	*transactions.TimedTransaction
	handler       *handler1
	log           util.Logger
	authEnabled   bool
	mqConnect     *mqPkts.ConnectPacket
	authenticated bool
}

func newConnectTransaction(ctx context.Context, h *handler1, authEnabled bool, mqConnect *mqPkts.ConnectPacket) *connectTransaction {
	tLog := h.log.WithTag("CONNECT")
	tLog.Debug("Created.")
	return &connectTransaction{
		TimedTransaction: transactions.NewTimedTransaction(
			ctx, connectTransactionTimeout,
			func() {
				h.transactions.DeleteByType(snPkts.CONNECT)
				tLog.Debug("Deleted.")
			},
		),
		handler:     h,
		log:         tLog,
		authEnabled: authEnabled,
		mqConnect:   mqConnect,
	}
}

func (t *connectTransaction) Start(ctx context.Context) error {
	t.handler.group.Go(func() error {
		select {
		case <-t.Done():
			if err := t.Err(); err != nil {
				if err == Cancelled {
					return nil
				}
				return fmt.Errorf("CONNECT: %s", err)
			}
			t.log.Debug("CONNECT transaction finished successfully.")
			return nil
		case <-ctx.Done():
			t.log.Debug("CONNECT transaction cancelled.")
			return nil
		}
	})

	if t.authEnabled {
		t.log.Debug("Waiting for AUTH packet.")
		return nil
	}

	if t.mqConnect.WillFlag {
		// Continue with WILLTOPICREQ.
		return t.handler.snSend(snPkts1.NewWillTopicReq())
	}

	return t.handler.mqttSend(t.mqConnect)
}

func (t *connectTransaction) Auth(snPkt *snPkts1.Auth) error {
	// Extract username and password from PLAIN data.
	if snPkt.Method == snPkts1.AUTH_PLAIN {
		user, password, err := snPkt.DecodePlain()
		if err != nil {
			t.Fail(err)
			return err
		}
		t.mqConnect.UsernameFlag = true
		t.mqConnect.Username = user
		t.mqConnect.PasswordFlag = true
		t.mqConnect.Password = password
	} else {
		if err := t.SendConnack(snPkts1.RC_NOT_SUPPORTED); err != nil {
			return err
		}
		err := fmt.Errorf("unknown auth method: %#v", snPkt.Method)
		t.Fail(err)
		return err
	}

	if t.mqConnect.WillFlag {
		// Continue with WILLTOPICREQ.
		return t.handler.snSend(snPkts1.NewWillTopicReq())
	}

	// All information successfully gathered - send MQTT connect.
	return t.handler.mqttSend(t.mqConnect)
}

func (t *connectTransaction) WillTopic(snWillTopic *snPkts1.WillTopic) error {
	t.mqConnect.WillQos = snWillTopic.QOS
	t.mqConnect.WillRetain = snWillTopic.Retain
	t.mqConnect.WillTopic = snWillTopic.WillTopic

	// Continue with WILLMSGREQ.
	return t.handler.snSend(snPkts1.NewWillMsgReq())
}

func (t *connectTransaction) WillMsg(snWillMsg *snPkts1.WillMsg) error {
	t.mqConnect.WillMessage = snWillMsg.WillMsg

	// All information successfully gathered - send MQTT connect.
	return t.handler.mqttSend(t.mqConnect)
}

func (t *connectTransaction) Connack(mqConnack *mqPkts.ConnackPacket) error {
	if mqConnack.ReturnCode != mqPkts.Accepted {
		// We misuse RC_CONGESTION here because MQTT-SN spec v. 1.2 does not define
		// any suitable return code.
		if err := t.SendConnack(snPkts1.RC_CONGESTION); err != nil {
			return err
		}
		returnCodeStr, ok := mqPkts.ConnackReturnCodes[mqConnack.ReturnCode]
		if !ok {
			returnCodeStr = "unknown code!"
		}
		err := fmt.Errorf(
			"CONNECT refused by MQTT broker with return code %d (%s).",
			mqConnack.ReturnCode, returnCodeStr)
		t.Fail(err)
		return err
	}

	// Must be set before snSend to avoid race condition in tests.
	t.handler.setState(util.StateActive)
	if err := t.SendConnack(snPkts1.RC_ACCEPTED); err != nil {
		t.Fail(err)
		return err
	}
	t.Success()
	return nil
}

// Inform client that the CONNECT request was refused.
func (t *connectTransaction) SendConnack(code snPkts1.ReturnCode) error {
	snConnack := snPkts1.NewConnack(code)
	if err := t.handler.snSend(snConnack); err != nil {
		t.Fail(err)
		return err
	}
	return nil
}
