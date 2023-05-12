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

package rpcretry

import (
	"context"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/backoffutils"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor returns a new unary client interceptor that retries the execution of external gRPC calls, the
// retry attempt will only occur if any of the validators define the error as a trigger.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	callOpts := evaluateClientOpt(opts...)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var md metadata.MD
		var err error
		logger := log.FromContext(ctx)

		for attempt := uint(0); attempt <= callOpts.max; attempt++ {
			retryTimeout := callOpts.timeout
			if callOpts.enableMetadata {
				if headerTimeout := getHeaderLimiterTimeout(ctx, md, callOpts); headerTimeout > 0 {
					retryTimeout = headerTimeout
				}
			}

			err = waitRetryBackoff(ctx, retryTimeout, attempt)
			if err != nil {
				logger.WithError(err).Error("An unexpected error occurred while in the timeout for the next request retry")
				return err
			}

			err = invoker(ctx, method, req, reply, cc, append(opts, grpc.Header(&md))...)
			if err == nil {
				return nil
			}

			if !isRetriable(err, callOpts) {
				return err
			}
		}

		return err
	}
}

func evaluateClientOpt(opts ...Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions

	for _, f := range opts {
		f(optCopy)
	}
	return optCopy
}

func isRetriable(err error, opt *options) bool {
	if err == nil {
		return false
	}
	for _, check := range opt.validators {
		if check(err) {
			return true
		}
	}
	return false
}

func waitRetryBackoff(ctx context.Context, timeout time.Duration, attempt uint) error {
	if attempt == 0 {
		return nil
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"attempt", attempt,
		"timeout", timeout,
	)).Debug("Failed request, waiting until next attempt")
	timer := time.NewTicker(timeout)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
	}
	return nil
}

func getHeaderLimiterTimeout(ctx context.Context, md metadata.MD, opt *options) time.Duration {
	if len(md.Get("x-rate-limit-reset")) <= 0 {
		return -1
	}
	var reset int
	reset, err := strconv.Atoi(md.Get("x-rate-limit-reset")[0])
	if err != nil {
		return -1
	}

	timeout := time.Duration(reset) * time.Second
	if opt.jitter > 0 {
		timeout = backoffutils.JitterUp(timeout, opt.jitter)
	}
	return timeout
}
