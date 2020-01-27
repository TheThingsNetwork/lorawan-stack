// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var defaultBackoff = []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second}

const defaultJitter = 0.15

// RetryServer is a Server that attempts to automatically re-subscribe to the upstream server by
// proxying Subscribe calls.
type RetryServer struct {
	backoff []time.Duration
	jitter  float64

	upstream Server
}

// Option represents an option for the retry backend.
type Option interface {
	apply(*RetryServer)
}

// OptionFunc is an option represented by a function.
type OptionFunc func(*RetryServer)

func (f OptionFunc) apply(rs *RetryServer) {
	f(rs)
}

// WithBackoff configures the backoff interval for the resubscription attempts.
func WithBackoff(backoff []time.Duration) Option {
	return OptionFunc(func(rs *RetryServer) {
		rs.backoff = backoff
	})
}

// WithJitter configures the jitter to be added to the resubscription attempts.
func WithJitter(jitter float64) Option {
	return OptionFunc(func(rs *RetryServer) {
		rs.jitter = jitter
	})
}

// NewRetryServer creates a new RetryServer with the given upstream and options.
func NewRetryServer(upstream Server, opts ...Option) Server {
	rs := &RetryServer{
		backoff:  defaultBackoff,
		jitter:   defaultJitter,
		upstream: upstream,
	}
	for _, opt := range opts {
		opt.apply(rs)
	}
	return rs
}

// GetBaseConfig implements Server using the upstream Server.
func (rs RetryServer) GetBaseConfig(ctx context.Context) config.ServiceBase {
	return rs.upstream.GetBaseConfig(ctx)
}

// FillContext implements Server using the upstream Server.
func (rs RetryServer) FillContext(ctx context.Context) context.Context {
	return rs.upstream.FillContext(ctx)
}

// SendUp implements Server using the upstream Server.
func (rs RetryServer) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	return rs.upstream.SendUp(ctx, up)
}

func (rs RetryServer) shouldRetry(err error) bool {
	switch {
	case errors.IsFailedPrecondition(err),
		errors.IsUnauthenticated(err),
		errors.IsPermissionDenied(err),
		errors.IsInvalidArgument(err):
		return false
	default:
		return true
	}
}

// Subscribe implements Server by proxying the Subscription object between the upstream server and the frontend.
func (rs RetryServer) Subscribe(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*Subscription, error) {
	downstreamSub := NewSubscription(ctx, protocol, &ids)
	upstreamSub, err := rs.upstream.Subscribe(ctx, protocol, ids)
	if err != nil {
		return nil, err
	}
	go func() {
		logger := log.FromContext(ctx)
	nextUp:
		for {
			select {
			case up := <-upstreamSub.Up():
				err := downstreamSub.SendUp(up.Context, up.ApplicationUp)
				if err != nil {
					logger.WithError(err).Warn("Failed to send the uplink downstream")
				}
			case <-ctx.Done():
				err := ctx.Err()
				downstreamSub.Disconnect(err)
				upstreamSub.Disconnect(err)
				logger.WithError(err).Debug("Parent context canceled")
				return
			case <-upstreamSub.Context().Done():
				err := upstreamSub.Context().Err()
				if rs.shouldRetry(err) {
					logger.Debug("Upstream subscription canceled. Attempting to resubscribe")
					for _, backoff := range rs.backoff {
						delay := random.Jitter(backoff, rs.jitter)
						select {
						case <-ctx.Done():
							err := ctx.Err()
							logger.WithError(err).Debug("Parent context canceled while attempting to resubscribe")
							return
						case <-downstreamSub.Context().Done():
							err := downstreamSub.Context().Err()
							logger.WithError(err).Debug("Downstream subscription canceled while attempting to resubscribe")
							return
						case <-time.After(delay):
						}
						upstreamSub, err = rs.upstream.Subscribe(ctx, protocol, ids)
						if err == nil {
							logger.Debug("Resubscription successful")
							continue nextUp
						}
						logger.WithError(err).WithField("delay", delay).Debug("Resubscription failed")
					}
				}
				downstreamSub.Disconnect(err)
				logger.WithError(err).Debug("Upstream resubscription attempts failed. Downstream subscription canceled")
				return
			case <-downstreamSub.Context().Done():
				err := downstreamSub.Context().Err()
				upstreamSub.Disconnect(err)
				logger.WithError(err).Debug("Downstream subscription canceled")
				return
			}
		}
	}()
	return downstreamSub, nil
}

// DownlinkQueuePush implements Server using the upstream Server.
func (rs RetryServer) DownlinkQueuePush(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, downlinks []*ttnpb.ApplicationDownlink) error {
	return rs.upstream.DownlinkQueuePush(ctx, ids, downlinks)
}

// DownlinkQueueReplace implements Server using the upstream Server.
func (rs RetryServer) DownlinkQueueReplace(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, downlinks []*ttnpb.ApplicationDownlink) error {
	return rs.upstream.DownlinkQueueReplace(ctx, ids, downlinks)
}

// DownlinkQueueList implements Server using the upstream Server.
func (rs RetryServer) DownlinkQueueList(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error) {
	return rs.upstream.DownlinkQueueList(ctx, ids)
}
