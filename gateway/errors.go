package gateway

import "errors"

// This error is used to shut down the handler from a goroutine.
// It does not signalize error.
var Shutdown = errors.New("clean shutdown")

var TransactionCanceled = errors.New("transaction canceled")

var ErrTopicIDsExhausted = errors.New("no more TopicIDs available")
var ErrPacketIDsExhausted = errors.New("no more PacketIDs available")
var ErrMqttConnClosed = errors.New("MQTT broker closed connection")
var ErrIllegalPacketWhenDisconnected = errors.New("illegal packet in disconnected state")
var ErrMissingPrivateKey = errors.New("private key is missing")
var ErrTLSCertMissing = errors.New("TLS certificate is missing")
