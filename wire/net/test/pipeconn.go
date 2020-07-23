// Copyright 2019 - See NOTICE file for copyright holders.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"net"

	"github.com/pkg/errors"

	"perun.network/go-perun/pkg/sync/atomic"
	"perun.network/go-perun/wire"
	wirenet "perun.network/go-perun/wire/net"
)

// TestConn is a testing connection.
type TestConn struct {
	closed *atomic.Bool
	conn   wirenet.Conn
}

// Send sends an envelope.
func (c *TestConn) Send(e *wire.Envelope) (err error) {
	if err = c.conn.Send(e); err != nil {
		c.Close()
	}
	return
}

// Recv receives an envelope.
func (c *TestConn) Recv() (e *wire.Envelope, err error) {
	if e, err = c.conn.Recv(); err != nil {
		c.Close()
	}
	return
}

// Close closes the TestConn.
func (c *TestConn) Close() error {
	if !c.closed.TrySet() {
		return errors.New("already closed")
	}
	return c.conn.Close()
}

// IsClosed returns whether the TestConn is already closed.
func (c *TestConn) IsClosed() bool {
	return c.closed.IsSet()
}

// NewTestConnPair creates endpoints that are connected via pipes.
func NewTestConnPair() (a wirenet.Conn, b wirenet.Conn) {
	closed := new(atomic.Bool)
	c0, c1 := net.Pipe()
	return &TestConn{closed, wirenet.NewIoConn(c0)}, &TestConn{closed, wirenet.NewIoConn(c1)}
}
