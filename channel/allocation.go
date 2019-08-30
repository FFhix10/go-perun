// Copyright (c) 2019 The Perun Authors. All rights reserved.
// This file is part of go-perun. Use of this source code is governed by a
// MIT-style license that can be found in the LICENSE file.

package channel

import (
	"math/big"

	"perun.network/go-perun/log"
	"perun.network/go-perun/pkg/io"

	"github.com/pkg/errors"
)

// Allocation and associated types
type (
	// Allocation is the distribution of assets, were the channel to be finalized.
	//
	// Assets identify the assets held in the channel, like an address to the
	// deposit holder contract for this asset.
	//
	// OfParts holds the balance allocations to the participants.
	// Its outer dimension must match the size of the Params.parts slice.
	// Its inner dimension must match the size of Assets.
	// All asset distributions could have been saved as a single []SubAlloc, but this
	// would have saved the participants slice twice, wasting space.
	//
	// Locked holds the locked allocations to sub-app-channels.
	Allocation struct {
		// Assets are the asset types held in this channel
		Assets []Asset
		// OfParts is the allocation of assets to the Params.parts
		OfParts [][]Bal
		// Locked is the locked allocation to sub-app-channels. It is allowed to be
		// nil, in which case there's nothing locked.
		Locked []SubAlloc
	}

	// SubAlloc is the allocation of assets to a single receiver channel `ID`.
	// The size of the balances slice must be of the same size as the assets slice
	// of the channel Params.
	SubAlloc struct {
		ID   ID
		Bals []Bal
	}

	// Bal is a single asset's balance
	Bal = *big.Int

	// Asset identifies an asset. E.g., it may be the address of the multi-sig
	// where all participants' assets are deposited.
	// The same Asset should be shareable by multiple Allocation instances.
	Asset = io.Serializable
)

// Clone returns a deep copy of the Allocation object.
// If it is nil, it returns nil.
func (orig Allocation) Clone() (clone Allocation) {
	if orig.Assets != nil {
		clone.Assets = make([]Asset, len(orig.Assets))
		for i := 0; i < len(clone.Assets); i++ {
			clone.Assets[i] = orig.Assets[i]
		}
	}

	if orig.OfParts != nil {
		clone.OfParts = make([][]Bal, len(orig.OfParts))
		for i := 0; i < len(clone.OfParts); i++ {
			clone.OfParts[i] = CloneBals(orig.OfParts[i])
		}
	}

	if orig.Locked != nil {
		clone.Locked = make([]SubAlloc, len(orig.Locked))
		for i := 0; i < len(clone.Locked); i++ {
			clone.Locked[i] = SubAlloc{
				ID:   orig.Locked[i].ID,
				Bals: CloneBals(orig.Locked[i].Bals),
			}
		}
	}

	return clone
}

func CloneBals(orig []Bal) []Bal {
	if orig == nil {
		return nil
	}

	clone := make([]Bal, len(orig))
	for i := 0; i < len(clone); i++ {
		clone[i] = new(big.Int).Set(orig[i])
	}
	return clone
}

// valid checks that the asset-dimensions match and slices are not nil.
// Assets and OfParts cannot be of zero length.
func (a Allocation) valid() error {
	if len(a.Assets) == 0 || len(a.OfParts) == 0 {
		return errors.New("assets and participant balances must not be of length zero")
	}

	n := len(a.Assets)
	for i, pa := range a.OfParts {
		if len(pa) != n {
			return errors.Errorf("dimension mismatch of participant %d's balance vector", i)
		}
	}

	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
	for _, l := range a.Locked {
		if len(l.Bals) != n {
			return errors.Errorf("dimension mismatch of app-channel balance vector (ID: %x)", l.ID)
		}
	}

	return nil
}

// Sum returns the sum of each asset over all participant and locked allocations
// It runs an internal check that the dimensions of all slices are valid and
// panics if not.
func (a Allocation) Sum() []Bal {
	if err := a.valid(); err != nil {
		log.Panic(err)
	}

	n := len(a.Assets)
	totals := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		totals[i] = new(big.Int)
	}

	for _, bals := range a.OfParts {
		for i, bal := range bals {
			totals[i].Add(totals[i], bal)
		}
	}

	// Locked is allowed to have zero length, in which case there's nothing locked
	// and the loop is empty.
	for _, a := range a.Locked {
		for i, bal := range a.Bals {
			totals[i].Add(totals[i], bal)
		}
	}

	return totals
}

// summer returns sums of balances
type summer interface {
	Sum() []Bal
}

func equalSum(b0, b1 summer) (bool, error) {
	s0, s1 := b0.Sum(), b1.Sum()
	n := len(s0)
	if n != len(s1) {
		return false, errors.New("dimension mismatch")
	}

	for i := 0; i < n; i++ {
		if s0[i].Cmp(s1[i]) != 0 {
			return false, nil
		}
	}
	return true, nil
}
