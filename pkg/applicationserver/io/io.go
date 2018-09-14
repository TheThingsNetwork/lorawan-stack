// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package io

import (
	"context"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const bufferSize = 10

// Server represents the Application Server to gateway frontends.
type Server interface {
	// Connect connects an application or integration by its identifiers to the Application Server, and returns a
	// Connection for traffic and control.
	Connect(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*Connection, error)
}

// Connection is a connection to an application or integration managed by a frontend.
type Connection struct {
	ctx context.Context

	protocol string
	ttnpb.ApplicationIdentifiers

	upCh chan *ttnpb.ApplicationUp
}

func NewConnection(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) *Connection {
	return &Connection{
		ctx:                    ctx,
		protocol:               protocol,
		ApplicationIdentifiers: ids,
		upCh:                   make(chan *ttnpb.ApplicationUp, bufferSize),
	}
}

// Context returns the connection context.
func (c *Connection) Context() context.Context { return c.ctx }

// Protocol returns the protocol used for the connection, i.e. grpc, mqtt or http.
func (c *Connection) Protocol() string { return c.protocol }

var errBufferFull = errors.DefineResourceExhausted("buffer_full", "buffer is full")

// SendUp sends an upstream message.
// This method returns immediately, returning nil if the message is buffered, or with an error when the buffer is full.
func (c *Connection) SendUp(up *ttnpb.ApplicationUp) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.upCh <- up:
	default:
		return errBufferFull
	}
	return nil
}

// Up returns the upstream channel.
func (c *Connection) Up() <-chan *ttnpb.ApplicationUp {
	return c.upCh
}
