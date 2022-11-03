package gateway

import "fmt"

// Transactions state constants.
type transactionState int

const (
	transactionDone transactionState = iota
	awaitingRegack
	awaitingPuback
	awaitingPubrec
	awaitingPubrel
	awaitingPubcomp
)

func (s transactionState) String() string {
	switch s {
	case transactionDone:
		return "done"
	case awaitingRegack:
		return "awaitingRegack"
	case awaitingPuback:
		return "awaitingPuback"
	case awaitingPubrec:
		return "awaitingPubrec"
	case awaitingPubrel:
		return "awaitingPubrel"
	case awaitingPubcomp:
		return "awaitingPubcomp"
	default:
		return fmt.Sprintf("invalid(%d)", s)
	}
}
