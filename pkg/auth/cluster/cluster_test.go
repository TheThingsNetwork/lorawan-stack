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
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
)

func TestAuthorized(t *testing.T) {
	a := assertions.New(t)

	if !a.So(func() { cluster.Authorized(context.Background()) }, should.Panic) {
		t.Fatal("Authorized should fail if there is no authentication data")
	}

	authorizedCtx := cluster.NewContext(context.Background(), nil)
	a.So(cluster.Authorized(authorizedCtx), should.BeNil)
	unauthorizedCtx := cluster.NewContext(context.Background(), errors.New("Unauthorized"))
	a.So(cluster.Authorized(unauthorizedCtx), should.NotBeNil)
}

func TestVerify(t *testing.T) {
	a := assertions.New(t)

	keys := [][]byte{
		{0x00, 0xaa, 0x12},
	}
	ctxWithAuthorization := cluster.VerifySource(context.Background(), keys)
	a.So(cluster.Authorized(ctxWithAuthorization), should.NotBeNil)

	md := metadata.MD{}

	for _, tc := range []struct {
		key     []byte
		success bool
	}{
		{
			key:     keys[0],
			success: true,
		},
		{
			key:     []byte{0x00, 0x00},
			success: false,
		},
	} {
		md["authorization"] = []string{fmt.Sprintf("%s %X", cluster.AuthType, tc.key)}
		ctxWithMetadata := metadata.NewIncomingContext(context.Background(), md)
		ctxWithAuthorization = cluster.VerifySource(ctxWithMetadata, keys)
		if tc.success {
			a.So(cluster.Authorized(ctxWithAuthorization), should.BeNil)
		} else {
			a.So(cluster.Authorized(ctxWithAuthorization), should.NotBeNil)
		}
	}
}

func ExampleAuthorized() {
	var ( // Assume this comes from a hypothetical inter-cluster RPC call.
		ctx context.Context
	)

	if err := cluster.Authorized(ctx); err != nil {
		// return err
	}
}
