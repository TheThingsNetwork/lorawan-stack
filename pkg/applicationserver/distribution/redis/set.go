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

package redis

import (
	"context"

	"github.com/go-redis/redis/v7"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type set struct {
	ctx    context.Context
	cancel errorcontext.CancelFunc

	init    chan struct{}
	initErr error

	ps *redis.PubSub

	distribution.Distributor
}

func (s *set) run() (err error) {
	logger := log.FromContext(s.ctx)
	defer func() {
		s.cancel(err)
		if err := s.ps.Close(); err != nil {
			logger.WithError(err).Warn("Failed to close pub/sub")
		}
	}()
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case msg, ok := <-s.ps.Channel():
			if !ok {
				return
			}

			up := &ttnpb.ApplicationUp{}
			if err := ttnredis.UnmarshalProto(msg.Payload, up); err != nil {
				logger.WithError(err).Warn("Failed to unmarshal upstream message")
				continue
			}

			if err := s.SendUp(s.ctx, up); err != nil {
				logger.WithError(err).Warn("Failed to send upstream message")
				continue
			}
		}
	}
}

func (s *set) setup(ctx context.Context, ps *redis.PubSub) error {
	s.ctx, s.cancel = errorcontext.New(ctx)
	s.Distributor = distribution.NewSubscriptionSet(ctx)
	s.ps = ps
	go s.run()
	return nil
}
