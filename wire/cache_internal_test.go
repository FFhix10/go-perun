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

package wire

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wtest "perun.network/go-perun/wallet/test"
	"polycry.pt/poly-go/test"
)

// newRandomEnvelope - copy from wire/test for internal tests.
func newRandomEnvelope(rng *rand.Rand, m Msg) *Envelope {
	return &Envelope{
		Sender:    wtest.NewRandomAddress(rng),
		Recipient: wtest.NewRandomAddress(rng),
		Msg:       m,
	}
}

func TestCache(t *testing.T) {
	assert, require := assert.New(t), require.New(t)
	rng := test.Prng(t)

	c := MakeCache()
	require.Zero(c.Size())

	ping0 := newRandomEnvelope(rng, NewPingMsg())
	pong := newRandomEnvelope(rng, NewPongMsg())
	ping1 := newRandomEnvelope(rng, NewPingMsg())
	ping2 := newRandomEnvelope(rng, NewPingMsg())
	// we want to uniquely identify messages by their timestamp
	require.False(ping0.Msg.(*PingMsg).Created.Equal(ping1.Msg.(*PingMsg).Created))

	assert.False(c.Put(ping0), "Put into cache without predicate")
	assert.Zero(c.Size())

	isPing := func(e *Envelope) bool { return e.Msg.Type() == Ping }
	c.Cache(&isPing)
	assert.True(c.Put(ping0), "Put into cache with predicate")
	assert.Equal(1, c.Size())
	assert.False(c.Put(pong), "Put into cache with non-matching prediacte")
	assert.Equal(1, c.Size())
	assert.True(c.Put(ping1), "Put into cache with predicate")
	assert.Equal(2, c.Size())

	empty := c.Messages(func(*Envelope) bool { return false })
	assert.Len(empty, 0)

	c.Release(&isPing)
	assert.False(c.Put(ping2), "Put into cache with canceled predicate")
	assert.Equal(2, c.Size())
	assert.Len(c.preds, 0, "internal: Put should have removed canceled predicate")

	msgs := c.Messages(func(e *Envelope) bool {
		return e.Msg.Type() == Ping &&
			e.Msg.(*PingMsg).Created.Equal(ping0.Msg.(*PingMsg).Created)
	})
	assert.Equal(1, c.Size())
	require.Len(msgs, 1)
	assert.Same(msgs[0], ping0)

	c.Cache(&isPing)
	c.Flush()
	assert.Equal(0, c.Size())
	assert.False(c.Put(ping0), "flushed cache should not hold any predicates")
}
