// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package rpcretry_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpcretry"
	"go.thethings.network/lorawan-stack/v3/pkg/util/rpctest"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	exhaustedErr = errors.DefineResourceExhausted("rpcretry_unary_resource_exhausted", "mock error of a resource exhausted scenario")
	internalErr  = errors.DefineInternal("rpcretry_unary_internal", "mock error that represents an error that should not be retried")
)

type unaryService struct {
	rpctest.FooBarServer
	md metadata.MD

	counter uint
	err     error
	sleep   time.Duration
}

func (fs *unaryService) Unary(ctx context.Context, foo *rpctest.Foo) (*rpctest.Bar, error) {
	fs.counter++
	time.Sleep(fs.sleep)

	if err := grpc.SendHeader(ctx, fs.md); err != nil {
		return nil, err
	}
	return &rpctest.Bar{Message: "bar"}, fs.err
}

func Test_UnaryClientInterceptor(t *testing.T) {
	type Service struct {
		err   error
		sleep time.Duration
		md    metadata.MD
	}
	type Client struct {
		retries       uint
		retryTimeout  time.Duration
		useMetadata   bool
		jitter        float64
		contextFiller func(ctx context.Context) context.Context
	}

	for _, tt := range []struct {
		name              string
		service           Service
		client            Client
		errAssertion      func(error) bool
		expectedReqAmount int
	}{
		{
			name:              "no error",
			client:            Client{retries: 5, retryTimeout: 3 * test.Delay},
			service:           Service{sleep: 0},
			expectedReqAmount: 1,
		},
		{
			name:              "unretriable error",
			client:            Client{retries: 5, retryTimeout: 3 * test.Delay},
			service:           Service{err: internalErr.New()},
			errAssertion:      errors.IsInternal,
			expectedReqAmount: 1,
		},
		{
			name:              "retriable error",
			client:            Client{retries: 5, retryTimeout: 3 * test.Delay},
			service:           Service{err: exhaustedErr.New()},
			errAssertion:      errors.IsResourceExhausted,
			expectedReqAmount: 6,
		},
		{
			name: "timeout error",
			client: Client{
				retries:      5,
				retryTimeout: 10 * test.Delay,
				contextFiller: func(ctx context.Context) context.Context {
					ctx, _ = context.WithTimeout(ctx, 5*test.Delay)
					return ctx
				},
			},
			service:           Service{sleep: 5 * test.Delay},
			errAssertion:      errors.IsDeadlineExceeded,
			expectedReqAmount: 1,
		},
		{
			name: "timeout from metadata rate-limiter",
			client: Client{
				retries:     1,
				useMetadata: true,
			},
			service: Service{
				err: exhaustedErr.New(),
				md:  ratelimit.Result{Limit: 10, Remaining: 8, RetryAfter: time.Second, ResetAfter: time.Second}.GRPCHeaders(),
			},
			errAssertion:      errors.IsResourceExhausted,
			expectedReqAmount: 2,
		},
	} {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}

		server := grpc.NewServer()
		testService := &unaryService{
			err:   tt.service.err,
			sleep: tt.service.sleep,
			md:    tt.service.md,
		}
		rpctest.RegisterFooBarServer(server, testService)
		go server.Serve(lis)

		cc, err := grpc.DialContext(
			test.Context(), lis.Addr().String(), grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(rpcretry.UnaryClientInterceptor(
				rpcretry.WithMax(tt.client.retries),
				rpcretry.WithDefaultTimeout(tt.client.retryTimeout),
				rpcretry.UseMetadata(tt.client.useMetadata),
				rpcretry.WithJitter(tt.client.jitter),
			)),
		)
		if err != nil {
			t.Fail()
		}
		defer cc.Close()

		client := rpctest.NewFooBarClient(cc)
		t.Run(tt.name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := test.Context()
			if tt.client.contextFiller != nil {
				ctx = tt.client.contextFiller(ctx)
			}

			resp, err := client.Unary(ctx, &rpctest.Foo{Message: "foo"})
			if tt.errAssertion != nil {
				a.So(tt.errAssertion(err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
				a.So(resp.Message, should.Equal, "bar")
			}
			a.So(testService.counter, should.Equal, tt.expectedReqAmount)
		})
	}
}
