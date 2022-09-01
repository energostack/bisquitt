package client

import (
	"context"
	"fmt"

	pkts "github.com/energomonitor/bisquitt/packets"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
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
				client.transactions.DeleteByType(pkts.CONNECT)
				tLog.Debug("Deleted.")
			},
		),
		client: client,
	}
}

func (t *connectTransaction) Connack(connack *pkts1.Connack) {
	if connack.ReturnCode != pkts1.RC_ACCEPTED {
		t.Fail(fmt.Errorf("connection rejected: %s", connack.ReturnCode))
		return
	}
	t.client.setState(util.StateActive)
	t.Success()
}
