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

package rpcserver

import (
	"context"
	"net"
	"time"

	"go.thethings.network/lorawan-stack/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const inProcess = "in-process"

type inProcessAuthInfo struct{}

func (inProcessAuthInfo) AuthType() string { return inProcess }

type inProcessCredentials struct {
	ServerName string
}

func (inProcessCredentials) ClientHandshake(_ context.Context, _ string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return conn, inProcessAuthInfo{}, nil
}

func (inProcessCredentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	return conn, inProcessAuthInfo{}, nil
}

func (c inProcessCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		SecurityProtocol: inProcess,
		SecurityVersion:  version.TTN,
		ServerName:       c.ServerName,
	}
}

func (c *inProcessCredentials) Clone() credentials.TransportCredentials { return c }

func (c *inProcessCredentials) OverrideServerName(serverName string) error {
	c.ServerName = serverName
	return nil
}

func newInProcessListener(parent context.Context) *inProcessListener {
	ctx, cancel := context.WithCancel(parent)
	return &inProcessListener{
		ctx:    ctx,
		cancel: cancel,
		ch:     make(chan net.Conn),
	}
}

type inProcessListener struct {
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan net.Conn
}

func (l inProcessListener) Accept() (net.Conn, error) {
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	case conn := <-l.ch:
		return conn, nil
	}
}

func (l inProcessListener) Close() error {
	l.cancel()
	return nil
}

type inProcessAddr string

func (inProcessAddr) Network() string  { return "in-process" }
func (a inProcessAddr) String() string { return string(a) }

func (l inProcessListener) Addr() net.Addr { return inProcessAddr("in-process") }

func inProcessDialer(lis *inProcessListener) func(string, time.Duration) (net.Conn, error) {
	return func(addr string, timeout time.Duration) (net.Conn, error) {
		server, client := net.Pipe()
		select {
		case <-time.After(timeout):
			return nil, context.DeadlineExceeded
		case lis.ch <- server:
			return client, nil
		}
	}
}

// StartLoopback starts the server on a local address and returns a connection to that address.
// This function does not add the default DialOptions.
func StartLoopback(ctx context.Context, s *grpc.Server, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	lis := newInProcessListener(ctx)
	go s.Serve(lis)
	return grpc.Dial(
		lis.Addr().String(),
		append([]grpc.DialOption{
			grpc.WithDialer(inProcessDialer(lis)),
			grpc.WithTransportCredentials(&inProcessCredentials{}),
		}, opts...)...)
}
