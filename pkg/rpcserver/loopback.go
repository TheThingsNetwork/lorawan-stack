// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpcserver

import (
	"net"

	"google.golang.org/grpc"
)

// StartLoopback starts the server on a local address and returns a connection to that address
func StartLoopback(s *grpc.Server) (*grpc.ClientConn, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go s.Serve(lis)
	return grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
}
