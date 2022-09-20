package client

import (
	pkts "github.com/energomonitor/bisquitt/packets"
	pkts1 "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
)

type disconnectTransaction struct {
	*transaction
}

func newDisconnectTransaction(client *Client) *disconnectTransaction {
	tLog := client.log.WithTag("DISCONNECT")
	tLog.Debug("Created.")
	return &disconnectTransaction{
		transaction: &transaction{
			RetryTransaction: transactions.NewRetryTransaction(
				client.groupCtx, client.cfg.RetryDelay, client.cfg.RetryCount,
				func(lastPkt interface{}) error {
					tLog.Debug("Resend.")
					return client.send(lastPkt.(pkts.Packet))
				},
				func() {
					client.transactions.DeleteByType(pkts.DISCONNECT)
					tLog.Debug("Deleted.")
				},
			),
			client: client,
			log:    tLog,
		},
	}
}

func (t *disconnectTransaction) Disconnect(disconnect *pkts1.Disconnect) {
	t.Success()
}
