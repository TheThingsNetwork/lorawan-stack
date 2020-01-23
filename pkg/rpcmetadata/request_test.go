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

package rpcmetadata_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestRequestMetadata(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	md1 := MD{
		ID:             "some-id",
		AuthType:       "Key",
		AuthValue:      "foo",
		ServiceType:    "component",
		ServiceVersion: "1.2.3-dev",
		NetAddress:     "localhost",
		Host:           "hostfoo",
		URI:            "fooURI",
	}
	md2 := FromIncomingContext(ctx)

	{
		md, err := md1.GetRequestMetadata(test.Context())
		a.So(err, should.BeNil)
		a.So(md, should.Resemble, map[string]string{
			"id":            "some-id",
			"authorization": "Key foo",
		})
	}

	{
		md, err := md2.GetRequestMetadata(test.Context())
		a.So(err, should.BeNil)
		a.So(md, should.BeEmpty)
	}

	{
		ctx := metadata.NewIncomingContext(test.Context(), metadata.New(map[string]string{
			"id":            "some-id",
			"authorization": "Key foo",
			"host":          "test.local",
		}))
		callOpt, err := WithForwardedAuth(ctx, true)
		a.So(err, should.BeNil)
		requestMD, err := callOpt.(grpc.PerRPCCredsCallOption).Creds.GetRequestMetadata(ctx)
		a.So(err, should.BeNil)
		a.So(requestMD, should.Resemble, map[string]string{
			"id":            "some-id",
			"authorization": "Key foo",
		})
	}

	{
		ctx := metadata.NewIncomingContext(test.Context(), metadata.New(map[string]string{
			"id":   "some-id",
			"host": "test.local",
		}))
		_, err := WithForwardedAuth(ctx, true)
		a.So(errors.IsUnauthenticated(err), should.BeTrue)
	}
}
