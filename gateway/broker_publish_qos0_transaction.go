package gateway

import (
	"context"
	"fmt"

	snPkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
)

type brokerPublishQOS0Transaction struct {
	brokerPublishTransactionBase
}

func newBrokerPublishQOS0Transaction(ctx context.Context, h *handler1, msgID uint16) *brokerPublishQOS0Transaction {
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

func (t *brokerPublishQOS0Transaction) Regack(snRegack *snPkts1.Regack) error {
	return t.regack(snRegack, transactionDone)
}
