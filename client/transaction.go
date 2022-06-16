package client

import (
	pkts "github.com/energomonitor/bisquitt/packets1"
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
	Disconnect(*pkts.Disconnect)
}

// Transactions involving PINGRESP message.
type transactionWithPingresp interface {
	Pingresp(pingresp *pkts.Pingresp)
}
