package client

import (
	msgs "github.com/energomonitor/bisquitt/messages"
	"github.com/energomonitor/bisquitt/transactions"
	"github.com/energomonitor/bisquitt/util"
)

// Transactions states constants type.
type transactionState int

// Transactions states constants.
const (
	transactionDone transactionState = iota
	awaitingPuback
	awaitingPubrec
	awaitingPubrel
	awaitingPubcomp
	awaitingDisconnect
	awaitingPingresp
)

type transaction struct {
	*transactions.RetryTransaction
	client *Client
	log    util.Logger
}

// Transactions involving DISCONNECT message.
type transactionWithDisconnect interface {
	Disconnect(*msgs.DisconnectMessage)
}

// Transactions involving PINGRESP message.
type transactionWithPingresp interface {
	Pingresp(pingresp *msgs.PingrespMessage)
}
