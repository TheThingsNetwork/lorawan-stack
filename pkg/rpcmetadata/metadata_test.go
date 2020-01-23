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
	. "go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
)

func TestMD(t *testing.T) {
	a := assertions.New(t)

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

	ctx := md1.ToOutgoingContext(test.Context())
	md, _ := metadata.FromOutgoingContext(ctx)
	ctx = metadata.NewIncomingContext(ctx, md)

	md2 := FromIncomingContext(ctx)
	a.So(md2.ID, should.Equal, md1.ID)
	a.So(md2.AuthType, should.BeEmpty)  // should be set by gRPC, not here
	a.So(md2.AuthValue, should.BeEmpty) // should be set by gRPC, not here
	a.So(md2.ServiceType, should.Equal, md1.ServiceType)
	a.So(md2.ServiceVersion, should.Equal, md1.ServiceVersion)
	a.So(md2.NetAddress, should.Equal, md1.NetAddress)
	a.So(md2.Host, should.Equal, md1.Host)
	a.So(md2.URI, should.Equal, md1.URI)

	a.So(md1.RequireTransportSecurity(), should.BeTrue)
	a.So(md2.RequireTransportSecurity(), should.BeFalse)

	{
		ctx := metadata.NewIncomingContext(test.Context(), metadata.Pairs("authorization", "Key foo"))
		md3 := FromIncomingContext(ctx)
		a.So(md3.AuthType, should.Equal, "Key")
		a.So(md3.AuthValue, should.Equal, "foo")
	}
}
