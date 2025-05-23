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
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"cosmossdk.io/store"
	"github.com/berachain/beacon-kit/beacon/blockchain"
	cometbft "github.com/berachain/beacon-kit/consensus/cometbft/service"
	"github.com/berachain/beacon-kit/log"
	service "github.com/berachain/beacon-kit/node-core/services/registry"
	"github.com/berachain/beacon-kit/node-core/types"
)

// Compile-time assertion that node implements the NodeI interface.
var _ types.Node = (*node)(nil)

// node is the hard-type representation of the beacon-kit node.
type node struct {
	// logger is the node's logger.
	logger log.Logger
	// registry is the node's service registry.
	registry *service.Registry
	// shutdownTimeout is the maximum time to wait for the node to gracefully shutdown before forcing an exit.
	shutdownTimeout time.Duration
}

// New returns a new node.
func New[NodeT types.Node](shutdownTimeout time.Duration, registry *service.Registry, logger log.Logger) NodeT {
	n := &node{
		shutdownTimeout: shutdownTimeout,
		registry:        registry,
		logger:          logger,
	}

	//nolint:errcheck // should be safe
	return types.Node(n).(NodeT)
}

// Start starts the node.
func (n *node) Start(
	ctx context.Context,
) error {
	cctx, cancelFn := context.WithCancel(ctx)

	stop := make(chan struct{})
	sigc := make(chan os.Signal, 1)

	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	// make sure we only call shutdownFunc once
	var once sync.Once

	shutdownFunc := func(err error) {
		now := time.Now()
		n.logger.Error("Shutdown initiated", "timeout", n.shutdownTimeout.String(), "error", err)

		cancelFn()
		n.registry.StopAll()
		close(stop)

		n.logger.Info("Node shutdown completed", "duration", time.Since(now).String())
	}

	// listen to signals in a separate goroutine
	go func() {
		sig := <-sigc

		timeout := time.AfterFunc(n.shutdownTimeout, func() {
			n.logger.Error("Shutdown timeout exceeded, forcing exit", "timeout", n.shutdownTimeout.String())
			os.Exit(1)
		})
		defer timeout.Stop()

		once.Do(func() {
			shutdownFunc(fmt.Errorf("shutdown initiated by signal: %s", sig.String()))
		})
	}()

	err := n.registry.StartAll(cctx)
	if err != nil {
		once.Do(func() {
			shutdownFunc(fmt.Errorf("failed to start services: %w", err))
		})
		return err
	}

	// we wait here until the signal handler has shutdown the node
	<-stop

	return nil
}

// CommitMultiStore returns the CommitMultiStore from cometbft service.
func (n *node) CommitMultiStore() store.CommitMultiStore {
	var cometService *cometbft.Service
	err := n.registry.FetchService(&cometService)
	if err != nil || cometService == nil { // appease nilaway
		err = fmt.Errorf("failed to fetch cometbft service: %w", err)
		panic(err)
	}
	return cometService.CommitMultiStore()
}

// StorageBackend returns the storage backend from the blockchain service.
func (n *node) StorageBackend() blockchain.StorageBackend {
	var blockchainService *blockchain.Service
	err := n.registry.FetchService(&blockchainService)
	if err != nil || blockchainService == nil { // appease nilaway
		err = fmt.Errorf("failed to fetch blockchain service: %w", err)
		panic(err)
	}
	return blockchainService.StorageBackend()
}
