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

package rpcretry_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpcretry"
	"go.thethings.network/lorawan-stack/v3/pkg/util/rpctest"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	goodPing     = &rpctest.Foo{Message: "mock"}
	retryTimeout = 50 * time.Millisecond
)

type failingService struct {
	rpctest.FooBarServer

	reqCounter uint
	unaryErr   error
	reqSleep   time.Duration
}

func (fs *failingService) Unary(ctx context.Context, foo *rpctest.Foo) (*rpctest.Bar, error) {
	fs.reqCounter++
	if fs.reqSleep > 0 {
		time.Sleep(fs.reqSleep)
	}
	return &rpctest.Bar{Message: "bar"}, fs.unaryErr
}

func Test_UnaryClientInterceptor(t *testing.T) {

	// unary tests
	{

		exaushtedErr := errors.DefineResourceExhausted("rpcretry_unary_resource_exausted", "mock error of a resource exaushted scenario")
		internalErr := errors.DefineInternal("rpcretry_unary_internal", "mock error that represents an error that should not be retried")

		for _, tt := range []struct {
			name    string
			service struct {
				unaryErr error
				sleep    time.Duration
			}
			client struct {
				retries uint
				timeout time.Duration
			}
			contextFiller     func(ctx context.Context) context.Context
			errAssertion      func(error) bool
			expectedReqAmount int
		}{
			{
				name: "no error",
				client: struct {
					retries uint
					timeout time.Duration
				}{5, 3 * test.Delay},
				service: struct {
					unaryErr error
					sleep    time.Duration
				}{nil, 0},

				expectedReqAmount: 1,
			},
			{
				name: "unretriable error",
				client: struct {
					retries uint
					timeout time.Duration
				}{5, 3 * test.Delay},
				service: struct {
					unaryErr error
					sleep    time.Duration
				}{internalErr.New(), 0},

				errAssertion:      errors.IsInternal,
				expectedReqAmount: 1,
			},
			{
				name: "retriable error",
				client: struct {
					retries uint
					timeout time.Duration
				}{5, 3 * test.Delay},
				service: struct {
					unaryErr error
					sleep    time.Duration
				}{exaushtedErr.New(), 0},

				errAssertion:      errors.IsResourceExhausted,
				expectedReqAmount: 6,
			},
			// TODO: Finish.
			// {
			// 	name: "timeout error",
			// 	client: struct {
			// 		retries uint
			// 		timeout time.Duration
			// 	}{5, 10 * test.Delay},
			// 	service: struct {
			// 		unaryErr error
			// 		sleep    time.Duration
			// 	}{nil, 5 * test.Delay},

			// 	contextFiller: func(ctx context.Context) context.Context {
			// 		ctx, _ = context.WithTimeout(ctx, 5*test.Delay)
			// 		return ctx
			// 	},
			// 	errAssertion:      errors.IsDeadlineExceeded,
			// 	expectedReqAmount: 1,
			// },
			// TODO Add  metadata testcace
		} {

			lis, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				panic(err)
			}

			server := grpc.NewServer()
			failService := &failingService{
				unaryErr: tt.service.unaryErr,
				reqSleep: tt.service.sleep,
			}
			rpctest.RegisterFooBarServer(server, failService)
			go server.Serve(lis)

			ctx := context.Background()
			if tt.contextFiller != nil {
				ctx = tt.contextFiller(ctx)
			}

			cc, err := grpc.DialContext(
				ctx,
				lis.Addr().String(),
				grpc.WithInsecure(),
				grpc.WithUnaryInterceptor(
					rpcretry.UnaryClientInterceptor(
						rpcretry.WithMax(tt.client.retries),
						rpcretry.WithDefaultTimeout(tt.client.timeout),
					),
				),
			)
			if err != nil {
				t.Fail()
			}
			defer cc.Close()

			client := rpctest.NewFooBarClient(cc)

			t.Run(tt.name, func(t *testing.T) {
				a := assertions.New(t)

				resp, err := client.Unary(test.Context(), &rpctest.Foo{Message: "foo"})
				if tt.errAssertion != nil {
					t.Log(err)
					a.So(tt.errAssertion(err), should.BeTrue)
				} else {
					a.So(err, should.BeNil)
					a.So(resp.Message, should.Equal, "bar")
				}
				a.So(failService.reqCounter, should.Equal, tt.expectedReqAmount)
			})
		}
	}
}
