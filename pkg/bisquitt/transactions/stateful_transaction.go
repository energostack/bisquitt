package transactions

// StatefulTransaction is a common interface for transactions with an inner
// state (e.g. there are multiple steps to be taken to successfully complete
// the transaction).
type StatefulTransaction interface {
	Transaction
	// Proceed is to be called when the transaction progresses to a state
	// "state" with user state data "data".
	//
	// Example:
	//
	//   transaction.Proceed(awaitingPuback, publishMessage)
	//
	// The PUBLISH message was received and we are waiting for the
	// corresponding PUBACK message.
	Proceed(state interface{}, data interface{})
}
