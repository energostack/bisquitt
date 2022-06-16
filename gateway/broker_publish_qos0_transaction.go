package gateway

import (
	"context"
	"fmt"

	snPkts "github.com/energomonitor/bisquitt/packets1"
	"github.com/energomonitor/bisquitt/transactions"
)

type brokerPublishQOS0Transaction struct {
	brokerPublishTransactionBase
}

func newBrokerPublishQOS0Transaction(ctx context.Context, h *handler, msgID uint16) *brokerPublishQOS0Transaction {
	tLog := h.log.WithTag(fmt.Sprintf("PUBLISH0(%d)", msgID))
	tLog.Debug("Created.")
	t := &brokerPublishQOS0Transaction{
		brokerPublishTransactionBase: brokerPublishTransactionBase{
			handler: h,
			log:     tLog,
		},
	}
	t.RetryTransaction = transactions.NewRetryTransaction(
		ctx, h.cfg.RetryDelay, h.cfg.RetryCount, t.resend,
		func() {
			h.transactions.Delete(msgID)
			tLog.Debug("Deleted.")
		},
	)
	return t
}

func (t *brokerPublishQOS0Transaction) Regack(snRegack *snPkts.RegackMessage) error {
	return t.regack(snRegack, transactionDone)
}
