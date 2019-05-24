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

package cluster_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	. "go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc/metadata"
)

func TestVerifySource(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	a := assertions.New(t)

	key := []byte{0x2A, 0x9C, 0x2C, 0x3C, 0x2A, 0x9C, 0x2A, 0x9C, 0x2A, 0x9C, 0x2A, 0x9C, 0x2A, 0x9C, 0x2A, 0x9C}

	c, err := New(ctx, &config.Cluster{
		Keys: []string{
			hex.EncodeToString(key),
		},
	})
	a.So(err, should.BeNil)

	t.Run("empty secret", func(t *testing.T) {
		a := assertions.New(t)

		ctx := c.WithVerifiedSource(ctx)
		a.So(errors.IsUnauthenticated(clusterauth.Authorized(ctx)), should.BeTrue)
	})

	t.Run("invalid secret type", func(t *testing.T) {
		a := assertions.New(t)

		md := metadata.Pairs("authorization", "Basic invalid-secret")
		ctx := metadata.NewIncomingContext(ctx, md)

		ctx = c.WithVerifiedSource(ctx)
		a.So(errors.IsInvalidArgument(clusterauth.Authorized(ctx)), should.BeTrue)
	})

	t.Run("valid secret", func(t *testing.T) {
		a := assertions.New(t)

		md := metadata.Pairs("authorization", fmt.Sprintf("ClusterKey %s", hex.EncodeToString(key)))
		ctx := metadata.NewIncomingContext(ctx, md)

		ctx = c.WithVerifiedSource(ctx)
		a.So(clusterauth.Authorized(ctx), should.BeNil)
	})

	t.Run("wrong secret", func(t *testing.T) {
		a := assertions.New(t)

		md := metadata.Pairs("authorization", "ClusterKey 0102030405060708")
		ctx := metadata.NewIncomingContext(ctx, md)

		ctx = c.WithVerifiedSource(ctx)
		a.So(errors.IsPermissionDenied(clusterauth.Authorized(ctx)), should.BeTrue)
	})
}
