package client

import (
	pkts1 "github.com/energostack/bisquitt/packets1"
	"github.com/energostack/bisquitt/transactions"
	"github.com/energostack/bisquitt/util"
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

// Transactions involving DISCONNECT packet.
type transactionWithDisconnect interface {
	Disconnect(*pkts1.Disconnect)
}

// Transactions involving PINGRESP packet.
type transactionWithPingresp interface {
	Pingresp(pingresp *pkts1.Pingresp)
}
