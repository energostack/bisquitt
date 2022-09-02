// Package util includes common tools for various Bisquitt packages.
package util

import (
	"context"
	"net"
	"time"

	dtlsProtocol "github.com/pion/dtls/v2/pkg/protocol"
)

// ConnWithContext is a net.Conn wrapper which implements io.ReadWriteCloser
// interface with Read and Write methods cancellable using context.Context.
type ConnWithContext struct {
	conn    net.Conn
	ctx     context.Context
	timeout time.Duration
}

// NewConnWithContext creates a new ConnWithContext instance.
//
// The timeout parameter sets a "time granularity". Since Go does not support
// immediate cancellation of Read nor Write call, these calls will be canceled
// in at most "timeout" time after ctx cancellation. Lower timeout causes slightly
// higher overhead.
func NewConnWithContext(ctx context.Context, conn net.Conn, timeout time.Duration) *ConnWithContext {
	return &ConnWithContext{
		conn:    conn,
		ctx:     ctx,
		timeout: timeout,
	}
}

func (c *ConnWithContext) Read(p []byte) (int, error) {
AGAIN:
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		// continue
	}

	err := c.conn.SetReadDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return 0, err
	}

	n, err := c.conn.Read(p)
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Temporary() && e.Timeout() {
				goto AGAIN
			}
		case *dtlsProtocol.TimeoutError:
			goto AGAIN
		}
	}
	return n, err
}

func (c *ConnWithContext) Write(b []byte) (int, error) {
AGAIN:
	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	default:
		// continue
	}

	err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return 0, err
	}

	n, err := c.conn.Write(b)
	if err != nil {
		switch e := err.(type) {
		case net.Error:
			if e.Temporary() && e.Timeout() {
				goto AGAIN
			}
		case *dtlsProtocol.TimeoutError:
			goto AGAIN
		}
	}
	return n, err
}

func (c *ConnWithContext) Close() error {
	return c.conn.Close()
}
