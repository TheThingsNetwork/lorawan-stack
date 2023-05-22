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
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// DefaultBufferSize is the default size of a subscription uplink buffer.
const DefaultBufferSize = 128

// PubSub represents the Application Server Pub/Sub capabilities to application frontends.
type PubSub interface {
	// Publish publishes upstream traffic to the Application Server.
	Publish(ctx context.Context, up *ttnpb.ApplicationUp) error
	// Subscribe subscribes an application or integration by its identifiers to the Application Server, and returns a
	// Subscription for traffic and control. If the cluster parameter is true, the subscription receives all of the
	// traffic of the application. Otherwise, only traffic that was processed locally is sent.
	Subscribe(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers, cluster bool) (*Subscription, error)
}

// DownlinkQueueOperator represents the Application Server downlink queue operations to application frontends.
type DownlinkQueueOperator interface {
	// DownlinkQueuePush pushes the given downlink messages to the end device's application downlink queue.
	DownlinkQueuePush(context.Context, *ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
	// DownlinkQueueReplace replaces the end device's application downlink queue with the given downlink messages.
	DownlinkQueueReplace(context.Context, *ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
	// DownlinkQueueList lists the application downlink queue of the given end device.
	DownlinkQueueList(context.Context, *ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error)
}

// Cluster represents the Application Server cluster peers to application frontends.
type Cluster interface {
	// GetPeers returns peers with the given role.
	GetPeers(ctx context.Context, role ttnpb.ClusterRole) ([]cluster.Peer, error)
	// GetPeer returns a peer with the given role, and a responsibility for the
	// given identifiers. If the identifiers are nil, this function returns a random
	// peer from the list that would be returned by GetPeers.
	GetPeer(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (cluster.Peer, error)
	// GetPeerConn returns the gRPC client connection of a peer, if the peer is available as
	// as per GetPeer.
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
}

// EndDeviceRegistry represents the Application Server end device registry to application frontends.
type EndDeviceRegistry interface {
	// GetEndDevice retrieves the end device from the Application Server end device registry.
	// This call will be delegated to the underlying end device registry, and should not be
	// used on the hot path. It exists for provisioning purposes.
	GetEndDevice(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error)
}

// Server represents the Application Server to application frontends.
type Server interface {
	task.Starter
	httpclient.Provider
	PubSub
	DownlinkQueueOperator
	Cluster
	EndDeviceRegistry
	// FromRequestContext decouples the lifetime of the provided context from the values found in the context.
	FromRequestContext(context.Context) context.Context
	// GetBaseConfig returns the component configuration.
	GetBaseConfig(ctx context.Context) config.ServiceBase
	// FillContext fills the given context.
	// This method should only be used for request contexts.
	FillContext(ctx context.Context) context.Context
	// RateLimiter returns the rate limiter instance.
	RateLimiter() ratelimit.Interface
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

	upCh    chan *ContextualApplicationUp
	publish func(context.Context, context.Context, chan<- *ContextualApplicationUp, *ContextualApplicationUp) error
}

// SubscriptionOption is an option for a Subscription.
type SubscriptionOption interface {
	// apply is unexposed in order to ensure that options
	// are not applied after the Subscription has been created.
	apply(*Subscription)
}

type subscriptionOptionFunc func(s *Subscription)

func (f subscriptionOptionFunc) apply(s *Subscription) { f(s) }

// WithBlocking controls if the Publish call is blocking or not.
func WithBlocking(blocking bool) SubscriptionOption {
	return subscriptionOptionFunc(func(s *Subscription) {
		if blocking {
			s.publish = blockingPublish
		} else {
			s.publish = nonBlockingPublish
		}
	})
}

// WithBufferSize controls the size of the subscription buffer.
func WithBufferSize(bufferSize int) SubscriptionOption {
	return subscriptionOptionFunc(func(s *Subscription) {
		s.upCh = make(chan *ContextualApplicationUp, bufferSize)
	})
}

// NewSubscription instantiates a new application or integration subscription.
func NewSubscription(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers, opts ...SubscriptionOption) *Subscription {
	ctx, cancelCtx := errorcontext.New(ctx)
	s := &Subscription{
		ctx:       ctx,
		cancelCtx: cancelCtx,
		protocol:  protocol,
		ids:       ids,
		upCh:      make(chan *ContextualApplicationUp, DefaultBufferSize),
		publish:   nonBlockingPublish,
	}
	for _, opt := range opts {
		opt.apply(s)
	}
	return s
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

// Publish publishes an upstream message.
func (s *Subscription) Publish(ctx context.Context, up *ttnpb.ApplicationUp) error {
	ctxUp := &ContextualApplicationUp{
		Context:       ctx,
		ApplicationUp: up,
	}
	return s.publish(ctx, s.ctx, s.upCh, ctxUp)
}

func blockingPublish(ctx context.Context, subCtx context.Context, upCh chan<- *ContextualApplicationUp, up *ContextualApplicationUp) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-subCtx.Done():
		return subCtx.Err()
	case upCh <- up:
		return nil
	}
}

var errBufferFull = errors.DefineResourceExhausted("buffer_full", "buffer is full")

func nonBlockingPublish(ctx context.Context, subCtx context.Context, upCh chan<- *ContextualApplicationUp, up *ContextualApplicationUp) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-subCtx.Done():
		return subCtx.Err()
	case upCh <- up:
		return nil
	default:
		return errBufferFull.New()
	}
}

// Up returns the upstream channel.
func (s *Subscription) Up() <-chan *ContextualApplicationUp {
	return s.upCh
}

// Pipe pipes the output of the Subscription to the provided handler.
func (s *Subscription) Pipe(
	ctx context.Context,
	ts task.Starter,
	name string,
	submit func(context.Context, *ttnpb.ApplicationUp) error,
) {
	f := func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-s.ctx.Done():
				return s.ctx.Err()
			case up := <-s.upCh:
				if err := submit(up.Context, up.ApplicationUp); err != nil {
					log.FromContext(up.Context).WithError(err).Warn("Failed to submit message")
				}
			}
		}
	}
	ts.StartTask(&task.Config{
		Context: ctx,
		ID:      fmt.Sprintf("pipe_%v", name),
		Func:    f,
		Restart: task.RestartOnFailure,
		Backoff: task.DefaultBackoffConfig,
	})
}

// CleanDownlinks returns a copy of the given downlink items with only the fields that can be set by the application.
func CleanDownlinks(items []*ttnpb.ApplicationDownlink) []*ttnpb.ApplicationDownlink {
	res := make([]*ttnpb.ApplicationDownlink, 0, len(items))
	for _, item := range items {
		res = append(res, &ttnpb.ApplicationDownlink{
			SessionKeyId:   item.SessionKeyId, // SessionKeyID must be set when skipping application payload crypto.
			FPort:          item.FPort,
			FCnt:           item.FCnt, // FCnt must be set when skipping application payload crypto.
			FrmPayload:     item.FrmPayload,
			DecodedPayload: item.DecodedPayload,
			ClassBC:        item.ClassBC,
			Priority:       item.Priority,
			Confirmed:      item.Confirmed,
			CorrelationIds: item.CorrelationIds,
			ConfirmedRetry: item.ConfirmedRetry,
		})
	}
	return res
}
