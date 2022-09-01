package gateway

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
