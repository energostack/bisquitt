package client

import (
	"context"
	"fmt"

	msgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

type connectTransaction struct {
	*transactions.TimedTransaction
	client *Client
}

func newConnectTransaction(ctx context.Context, client *Client) *connectTransaction {
	tLog := client.log.WithTag("CONNECT")
	tLog.Debug("Created.")
	return &connectTransaction{
		TimedTransaction: transactions.NewTimedTransaction(
			ctx, client.cfg.ConnectTimeout,
			func() {
				client.transactions.DeleteByType(msgs.CONNECT)
				tLog.Debug("Deleted.")
			},
		),
		client: client,
	}
}

func (t *connectTransaction) Connack(connack *msgs.ConnackMessage) {
	if connack.ReturnCode != msgs.RC_ACCEPTED {
		t.Fail(fmt.Errorf("connection rejected: %s", connack.ReturnCode))
		return
	}
	t.client.setState(util.StateActive)
	t.Success()
}
