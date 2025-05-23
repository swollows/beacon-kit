// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, Berachain Foundation. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package types

import (
	"fmt"

	"github.com/berachain/beacon-kit/primitives/common"
	"github.com/berachain/beacon-kit/primitives/constants"
	"github.com/berachain/beacon-kit/primitives/constraints"
	"github.com/berachain/beacon-kit/primitives/crypto"
	sszutil "github.com/berachain/beacon-kit/primitives/encoding/ssz"
	"github.com/karalabe/ssz"
)

const sszConsolidationRequestSize = 116

// Compile-time check to ensure ConsolidationRequest implements the necessary interfaces.
var (
	_ ssz.StaticObject            = (*ConsolidationRequest)(nil)
	_ constraints.SSZMarshallable = (*ConsolidationRequest)(nil)
)

// ConsolidationRequest is introduced in Pectra but not used by us.
// We keep it so we can maintain parity tests with other SSZ implementations.
type ConsolidationRequest struct {
	SourceAddress common.ExecutionAddress
	SourcePubKey  crypto.BLSPubkey
	TargetPubKey  crypto.BLSPubkey
}

/* -------------------------------------------------------------------------- */
/*                       Consolidation Request SSZ                            */
/* -------------------------------------------------------------------------- */

func (c *ConsolidationRequest) ValidateAfterDecodingSSZ() error {
	return nil
}

func (c *ConsolidationRequest) DefineSSZ(codec *ssz.Codec) {
	ssz.DefineStaticBytes(codec, &c.SourceAddress)
	ssz.DefineStaticBytes(codec, &c.SourcePubKey)
	ssz.DefineStaticBytes(codec, &c.TargetPubKey)
}

func (c *ConsolidationRequest) SizeSSZ(_ *ssz.Sizer) uint32 {
	return sszConsolidationRequestSize
}

func (c *ConsolidationRequest) MarshalSSZ() ([]byte, error) {
	buf := make([]byte, ssz.Size(c))
	return buf, ssz.EncodeToBytes(buf, c)
}

// HashTreeRoot returns the hash tree root of the Deposits.
func (c *ConsolidationRequest) HashTreeRoot() common.Root {
	return ssz.HashSequential(c)
}

/* -------------------------------------------------------------------------- */
/*                       Consolidation Requests SSZ                           */
/* -------------------------------------------------------------------------- */

// Compile-time check to ensure ConsolidationRequests implements the necessary interfaces.
var _ constraints.SSZMarshaler = (*ConsolidationRequests)(nil)

// ConsolidationRequests is used for SSZ unmarshalling a list of ConsolidationRequest
type ConsolidationRequests []*ConsolidationRequest

// MarshalSSZ marshals the ConsolidationRequests object to SSZ format by encoding each consolidation request individually.
func (cr ConsolidationRequests) MarshalSSZ() ([]byte, error) {
	return sszutil.MarshalItemsEIP7685(cr)
}

// ValidateAfterDecodingSSZ validates the ConsolidationRequests object after decoding.
func (cr ConsolidationRequests) ValidateAfterDecodingSSZ() error {
	if len(cr) > constants.MaxConsolidationRequestsPerPayload {
		return fmt.Errorf(
			"invalid number of consolidation requests, got %d max %d",
			len(cr), constants.MaxConsolidationRequestsPerPayload,
		)
	}
	return nil
}

// DecodeConsolidationRequests decodes SSZ data by decoding each request individually.
func DecodeConsolidationRequests(data []byte) (ConsolidationRequests, error) {
	maxSize := constants.MaxConsolidationRequestsPerPayload * sszConsolidationRequestSize
	if len(data) > maxSize {
		return nil, fmt.Errorf(
			"invalid consolidation requests SSZ size, requests should not be more than the "+
				"max per payload, got %d max %d", len(data), maxSize,
		)
	}
	if len(data) < sszConsolidationRequestSize {
		return nil, fmt.Errorf(
			"invalid consolidation requests SSZ size, got %d expected at least %d",
			len(data), sszConsolidationRequestSize,
		)
	}
	if len(data)%sszConsolidationRequestSize != 0 {
		return nil, fmt.Errorf(
			"invalid data length: %d is not a multiple of consolidation request size %d",
			len(data), sszConsolidationRequestSize,
		)
	}

	// Use the EIP-7685 unmarshalItems helper.
	return sszutil.UnmarshalItemsEIP7685(
		data,
		sszConsolidationRequestSize,
		func() *ConsolidationRequest { return new(ConsolidationRequest) },
	)
}
