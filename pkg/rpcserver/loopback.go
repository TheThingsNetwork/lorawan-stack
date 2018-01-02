// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpcserver

import (
	"context"
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/rpcclient"
	"google.golang.org/grpc"
)

// StartLoopback starts the server on a local address and returns a connection to that address
func StartLoopback(ctx context.Context, s *grpc.Server, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go s.Serve(lis)
	return grpc.Dial(lis.Addr().String(), append(append(rpcclient.DefaultDialOptions(ctx), grpc.WithInsecure()), opts...)...)
}
