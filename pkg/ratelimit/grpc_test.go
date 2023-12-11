// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package ratelimit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smarty/assertions"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type serverStream struct {
	grpc.ServerStream
	ctx context.Context

	t *testing.T
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

func (*serverStream) RecvMsg(any) error {
	return nil
}

func (ss *serverStream) SetHeader(md metadata.MD) error {
	a := assertions.New(ss.t)
	a.So(md.Get("x-rate-limit-limit"), should.NotBeEmpty)
	a.So(md.Get("x-rate-limit-available"), should.NotBeEmpty)
	a.So(md.Get("x-rate-limit-reset"), should.NotBeEmpty)
	a.So(md.Get("x-rate-limit-retry"), should.NotBeEmpty)
	return nil
}

func grpcClusterContext() context.Context {
	return metadata.NewIncomingContext(test.Context(), metadata.Pairs(
		"authorization", fmt.Sprintf("%s %X", clusterauth.AuthType, []byte{0x00, 0x01, 0x02}),
	))
}

func grpcUnaryHandler(context.Context, any) (any, error) { return "response", nil }

func grpcStreamHandler(any, grpc.ServerStream) error { return nil }

type mockRequestWithKeyer struct {
	key string
}

func (r *mockRequestWithKeyer) RateLimitKey() string {
	return r.key
}

func TestGRPC(t *testing.T) {
	t.Parallel()

	const (
		unaryMethod  = "/Service/UnaryMethod"
		streamMethod = "/Service/StreamMethod"
	)

	t.Run("UnaryServerInterceptor", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name    string
			limiter *mockLimiter
			cluster bool
			assert  func(t *testing.T, limiter *mockLimiter, resp any, err error)
			request any
		}{
			{
				name:    "Cluster",
				limiter: &mockLimiter{limit: true},
				cluster: true,
				assert: func(t *testing.T, limiter *mockLimiter, resp any, err error) {
					t.Helper()

					a := assertions.New(t)
					a.So(resp, should.Resemble, "response")
					a.So(err, should.BeNil)
					a.So(limiter.calledWithResource, should.BeNil)
				},
			},
			{
				name:    "Pass",
				limiter: &mockLimiter{},
				assert: func(t *testing.T, limiter *mockLimiter, resp any, err error) {
					t.Helper()

					a := assertions.New(t)
					a.So(resp, should.Resemble, "response")
					a.So(err, should.BeNil)

					a.So(limiter.calledWithResource, should.NotBeNil)
					a.So(limiter.calledWithResource.Key(), should.ContainSubstring, unaryMethod)
					a.So(limiter.calledWithResource.Key(), should.ContainSubstring, authTokenID)
					a.So(
						limiter.calledWithResource.Classes(),
						should.Resemble,
						[]string{fmt.Sprintf("grpc:method:%s", unaryMethod), "grpc:method"},
					)
				},
			},
			{
				name:    "Limit",
				limiter: &mockLimiter{limit: true},
				assert: func(t *testing.T, limiter *mockLimiter, resp any, err error) {
					t.Helper()

					a := assertions.New(t)
					a.So(resp, should.BeNil)
					a.So(errors.IsResourceExhausted(err), should.BeTrue)

					a.So(limiter.calledWithResource, should.NotBeNil)
				},
			},
			{
				name:    "Keyer",
				limiter: &mockLimiter{limit: true},
				request: &mockRequestWithKeyer{key: "test_keyer"},
				assert: func(t *testing.T, limiter *mockLimiter, resp any, err error) {
					t.Helper()

					a := assertions.New(t)
					a.So(limiter.calledWithResource.Key(), should.ContainSubstring, "test_keyer")
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				intercept := ratelimit.UnaryServerInterceptor(tc.limiter)

				ctx := tokenContext(authTokenID)
				if tc.cluster {
					ctx = grpcClusterContext()
				}
				info := &grpc.UnaryServerInfo{FullMethod: unaryMethod}
				resp, err := intercept(ctx, tc.request, info, grpcUnaryHandler)
				tc.assert(t, tc.limiter, resp, err)
			})
		}
	})
	t.Run("StreamServerInterceptor", func(t *testing.T) {
		t.Parallel()

		t.Run("Streams", func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				limiter *mockLimiter
				cluster bool
				assert  func(t *testing.T, limiter *mockLimiter, err error)
			}{
				{
					name:    "Cluster",
					limiter: &mockLimiter{limit: true},
					cluster: true,
					assert: func(t *testing.T, limiter *mockLimiter, err error) {
						t.Helper()

						a := assertions.New(t)
						a.So(err, should.BeNil)
						a.So(limiter.calledWithResource, should.BeNil)
					},
				},
				{
					name:    "Pass",
					limiter: &mockLimiter{result: ratelimit.Result{Limit: 10}},
					assert: func(t *testing.T, limiter *mockLimiter, err error) {
						t.Helper()

						a := assertions.New(t)
						a.So(err, should.BeNil)

						a.So(limiter.calledWithResource.Key(), should.ContainSubstring, streamMethod)
						a.So(limiter.calledWithResource.Key(), should.ContainSubstring, authTokenID)
						a.So(
							limiter.calledWithResource.Classes(),
							should.Resemble,
							[]string{fmt.Sprintf("grpc:stream:accept:%s", streamMethod), "grpc:stream:accept"},
						)
					},
				},
				{
					name:    "Limit",
					limiter: &mockLimiter{limit: true, result: ratelimit.Result{Limit: 10}},
					assert: func(t *testing.T, limiter *mockLimiter, err error) {
						t.Helper()

						assertions.New(t).So(errors.IsResourceExhausted(err), should.BeTrue)
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					intercept := ratelimit.StreamServerInterceptor(tc.limiter)
					ss := &serverStream{t: t, ctx: tokenContext(authTokenID)}
					if tc.cluster {
						ss.ctx = grpcClusterContext()
					}
					info := &grpc.StreamServerInfo{FullMethod: streamMethod}

					err := intercept(nil, ss, info, grpcStreamHandler)
					tc.assert(t, tc.limiter, err)
				})
			}
		})

		t.Run("Traffic", func(t *testing.T) {
			a := assertions.New(t)
			limiter := muxMockLimiter{
				"grpc:stream:accept": &mockLimiter{result: ratelimit.Result{Limit: 10}},
				"grpc:stream:up":     &mockLimiter{},
			}
			intercept := ratelimit.StreamServerInterceptor(limiter)
			ss := &serverStream{t: t, ctx: tokenContext(authTokenID)}
			info := &grpc.StreamServerInfo{FullMethod: streamMethod}

			keyFromFirstStream := ""
			_ = intercept(nil, ss, info, func(req any, stream grpc.ServerStream) error {
				// Assert traffic limiter unused
				a.So(limiter["grpc:stream:up"].calledWithResource, should.BeNil)

				// Receive message
				a.So(stream.RecvMsg(nil), should.BeNil)
				keyFromFirstStream = limiter["grpc:stream:up"].calledWithResource.Key()

				a.So(limiter["grpc:stream:up"].calledWithResource.Key(), should.ContainSubstring, streamMethod)
				a.So(
					limiter["grpc:stream:up"].calledWithResource.Classes(),
					should.Resemble,
					[]string{fmt.Sprintf("grpc:stream:up:%s", streamMethod), "grpc:stream:up"},
				)

				// Enable rate limits
				limiter["grpc:stream:up"].limit = true

				// Receive message must fail
				a.So(errors.IsResourceExhausted(stream.RecvMsg(nil)), should.BeTrue)

				return nil
			})

			_ = intercept(nil, ss, info, func(req any, stream grpc.ServerStream) error {
				// receive message to use rate limiter
				a.So(errors.IsResourceExhausted(stream.RecvMsg(nil)), should.BeTrue)

				// Assert a different rate limiting key was used
				a.So(limiter["grpc:stream:up"].calledWithResource.Key(), should.ContainSubstring, streamMethod)
				a.So(limiter["grpc:stream:up"].calledWithResource.Key(), should.NotEqual, keyFromFirstStream)

				return nil
			})
		})
	})
}
