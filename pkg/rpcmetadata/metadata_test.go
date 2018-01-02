// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpcmetadata

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
		Limit:          12,
		Offset:         34,
	}

	ctx := md1.ToOutgoingContext(context.Background())
	md, _ := metadata.FromOutgoingContext(ctx)
	ctx = metadata.NewIncomingContext(ctx, md)

	md2 := FromIncomingContext(ctx)
	a.So(md2.ID, should.Equal, md1.ID)
	a.So(md2.AuthType, should.BeEmpty)  // should be set by gRPC, not here
	a.So(md2.AuthValue, should.BeEmpty) // should be set by gRPC, not here
	a.So(md2.ServiceType, should.Equal, md1.ServiceType)
	a.So(md2.ServiceVersion, should.Equal, md1.ServiceVersion)
	a.So(md2.NetAddress, should.Equal, md1.NetAddress)
	a.So(md2.Limit, should.Equal, md1.Limit)
	a.So(md2.Offset, should.Equal, md1.Offset)

	a.So(md1.RequireTransportSecurity(), should.BeTrue)
	a.So(md2.RequireTransportSecurity(), should.BeFalse)

	{
		md, err := md1.GetRequestMetadata(context.Background())
		a.So(err, should.BeNil)
		a.So(md, should.Resemble, map[string]string{"authorization": "Key foo"})
	}

	{
		md, err := md2.GetRequestMetadata(context.Background())
		a.So(err, should.BeNil)
		a.So(md, should.BeEmpty)
	}

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Key foo"))
	md3 := FromIncomingContext(ctx)
	a.So(md3.AuthType, should.Equal, "Key")
	a.So(md3.AuthValue, should.Equal, "foo")

}
