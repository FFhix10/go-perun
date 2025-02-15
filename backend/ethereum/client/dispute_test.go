// Copyright 2020 - See NOTICE file for copyright holders.
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

package client_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"perun.network/go-perun/backend/ethereum/channel/test"
	ctest "perun.network/go-perun/backend/ethereum/client/test"
	"perun.network/go-perun/client"
	clienttest "perun.network/go-perun/client/test"
	"perun.network/go-perun/log"
	"perun.network/go-perun/wire"
	pkgtest "polycry.pt/poly-go/test"
)

const (
	twoPartyTestTimeout = 10 * time.Second
	TxFinalityDepth     = 3
)

func TestDisputeMalloryCarol(t *testing.T) {
	log.Info("Starting dispute test")
	rng := pkgtest.Prng(t)

	const A, B = 0, 1 // Indices of Mallory and Carol
	var (
		name  = [2]string{"Mallory", "Carol"}
		setup [2]clienttest.RoleSetup
		role  [2]clienttest.Executer
	)

	s := test.NewSetup(t, rng, 2, ctest.BlockInterval, TxFinalityDepth)
	setup = ctest.MakeRoleSetups(s, name)

	role[A] = clienttest.NewMallory(t, setup[A])
	role[B] = clienttest.NewCarol(t, setup[B])

	execConfig := &clienttest.MalloryCarolExecConfig{
		BaseExecConfig: clienttest.MakeBaseExecConfig(
			[2]wire.Address{setup[A].Identity.Address(), setup[B].Identity.Address()},
			s.Asset,
			[2]*big.Int{big.NewInt(100), big.NewInt(1)},
			client.WithoutApp(),
		),
		NumPayments: [2]int{5, 0},
		TxAmounts:   [2]*big.Int{big.NewInt(20), big.NewInt(0)},
	}

	ctx, cancel := context.WithTimeout(context.Background(), twoPartyTestTimeout)
	defer cancel()
	err := clienttest.ExecuteTwoPartyTest(ctx, role, execConfig)
	require.NoError(t, err)

	// Assert correct final balances
	netTransfer := big.NewInt(int64(execConfig.NumPayments[A])*execConfig.TxAmounts[A].Int64() -
		int64(execConfig.NumPayments[B])*execConfig.TxAmounts[B].Int64())
	finalBal := [2]*big.Int{
		new(big.Int).Sub(execConfig.InitBals()[A], netTransfer),
		new(big.Int).Add(execConfig.InitBals()[B], netTransfer),
	}
	// reset context timeout
	ctx, cancel = context.WithTimeout(context.Background(), ctest.DefaultTimeout)
	defer cancel()
	for i, bal := range finalBal {
		b, err := s.SimBackend.BalanceAt(ctx, common.Address(*s.Recvs[i]), nil)
		require.NoError(t, err)
		assert.Zero(t, b.Cmp(bal), "ETH balance mismatch")
	}

	log.Info("Dispute test done")
}
