// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const bufferSize = 32

// Server represents the Application Server to gateway frontends.
type Server interface {
	// FillContext fills the given context.
	FillContext(ctx context.Context) context.Context
	// FillApplicationContext fills the given context and identifiers.
	FillApplicationContext(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (context.Context, ttnpb.ApplicationIdentifiers, error)
	// Subscribe subscribes an application or integration by its identifiers to the Application Server, and returns a
	// Subscription for traffic and control.
	Subscribe(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*Subscription, error)
	// DownlinkQueuePush pushes the given downlink messages to the end device's application downlink queue.
	DownlinkQueuePush(context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
	// DownlinkQueueReplace replaces the end device's application downlink queue with the given downlink messages.
	DownlinkQueueReplace(context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
	// DownlinkQueueList lists the application downlink queue of the given end device.
	DownlinkQueueList(context.Context, ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error)
}

// ContextualApplicationUp represents an ttnpb.ApplicationUp with its context.
type ContextualApplicationUp struct {
	context.Context
	*ttnpb.ApplicationUp
}

// Subscription is a subscription to an application or integration managed by a frontend.
type Subscription struct {
	ctx       context.Context
	cancelCtx errorcontext.CancelFunc

	protocol string
	ids      *ttnpb.ApplicationIdentifiers

	upCh chan *ContextualApplicationUp
}

// NewSubscription instantiates a new application or integration subscription.
func NewSubscription(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers) *Subscription {
	ctx, cancelCtx := errorcontext.New(ctx)
	return &Subscription{
		ctx:       ctx,
		cancelCtx: cancelCtx,
		protocol:  protocol,
		ids:       ids,
		upCh:      make(chan *ContextualApplicationUp, bufferSize),
	}
}

// Context returns the subscription context.
func (s *Subscription) Context() context.Context { return s.ctx }

// Disconnect marks the subscription as disconnected and cancels the context.
func (s *Subscription) Disconnect(err error) {
	s.cancelCtx(err)
}

// Protocol returns the protocol used for the subscription, i.e. grpc, mqtt or http.
func (s *Subscription) Protocol() string { return s.protocol }

// ApplicationIDs returns the application identifiers, if the subscription represents any specific.
func (s *Subscription) ApplicationIDs() *ttnpb.ApplicationIdentifiers { return s.ids }

var errBufferFull = errors.DefineResourceExhausted("buffer_full", "buffer is full")

// SendUp sends an upstream message.
// This method returns immediately, returning nil if the message is buffered, or with an error when the buffer is full.
func (s *Subscription) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case s.upCh <- &ContextualApplicationUp{
		Context:       ctx,
		ApplicationUp: up,
	}:
	default:
		return errBufferFull
	}
	return nil
}

// Up returns the upstream channel.
func (s *Subscription) Up() <-chan *ContextualApplicationUp {
	return s.upCh
}

// CleanDownlinks returns a copy of the given downlink items with only the fields that can be set by the application.
func CleanDownlinks(items []*ttnpb.ApplicationDownlink) []*ttnpb.ApplicationDownlink {
	res := make([]*ttnpb.ApplicationDownlink, 0, len(items))
	for _, item := range items {
		res = append(res, &ttnpb.ApplicationDownlink{
			FPort:          item.FPort,
			FRMPayload:     item.FRMPayload,
			DecodedPayload: item.DecodedPayload,
			ClassBC:        item.ClassBC,
			Priority:       item.Priority,
			Confirmed:      item.Confirmed,
			CorrelationIDs: item.CorrelationIDs,
		})
	}
	return res
}
