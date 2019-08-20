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

package component

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

func (c *Component) initRights() {
	fetcher := rights.NewAccessFetcher(func(ctx context.Context) *grpc.ClientConn {
		peer := c.GetPeer(ctx, ttnpb.ClusterRole_ACCESS, nil)
		if peer == nil {
			return nil
		}
		return peer.Conn()
	}, c.config.GRPC.AllowInsecureForCredentials)

	if c.config.Rights.TTL > 0 {
		fetcher = rights.NewInMemoryCache(fetcher, c.config.Rights.TTL, c.config.Rights.TTL)
	} else {
		c.Logger().Warn("No rights TTL configured")
	}

	c.rightsFetcher = fetcher
	c.AddContextFiller(func(ctx context.Context) context.Context {
		return rights.NewContextWithFetcher(ctx, fetcher)
	})
}
