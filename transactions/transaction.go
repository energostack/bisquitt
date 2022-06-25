// Package transactions implements basic transactions types which are to be embedded
// in the specific transaction types.
//
// A transaction is a tool to implement asynchronous, concurrent, stateful event handling.
// In Bisquitt, it's used to keep track of the related packets (e.g. PUBLISH - PUBACK)
// in the interweaved packets stream.
package transactions

// Transaction is a common transactions interface.
//
// It is inspired by the Token type in Eclipse Paho
// https://github.com/eclipse/paho.mqtt.golang/blob/18bfbdece2e98c020293755a3468bf510a7b2497/token.go#L33
type Transaction interface {
	// Fail is called when the transaction completes unsuccessfully.
	Fail(error)

	// Success is called when the transaction completes successfully.
	Success()

	// Done returns a channel that is closed when the transaction completes.
	// Clients should call Err after the channel is closed to check if
	// the transaction completed successfully.
	Done() <-chan struct{}

	// Err returns the error the transaction failed with or nil on
	// successful completion.
	Err() error
}
