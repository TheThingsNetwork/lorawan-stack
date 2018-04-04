// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package rpcserver

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

// StartLoopback starts the server on a local address and returns a connection to that address.
// This function does not add the default DialOptions.
func StartLoopback(ctx context.Context, s *grpc.Server, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go s.Serve(lis)
	return grpc.Dial(lis.Addr().String(), append(append([]grpc.DialOption{}, grpc.WithInsecure()), opts...)...)
}
