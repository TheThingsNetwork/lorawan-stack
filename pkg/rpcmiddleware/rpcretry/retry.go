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

package rpcretry

import (
	"context"
	"strconv"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	headerXRateLimit     = "x-rate-limit-limit"
	headerXRateAvailable = "x-rate-limit-available"
	headerXRateReset     = "x-rate-limit-reset"
	headerXRateRetry     = "x-rate-limit-retry"

	AttemptMetadataKey = "x-retry-attempty"
)

func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	callOpts := evaluateClientOpt(opts...)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logger := log.FromContext(ctx)

		var md metadata.MD
		var err error
		err = invoker(ctx, method, req, reply, cc, append(opts, grpc.Header(&md))...)
		var available, reset, retry int
		available, _ = strconv.Atoi(md.Get(headerXRateAvailable)[0])
		reset, _ = strconv.Atoi(md.Get(headerXRateReset)[0])
		retry, _ = strconv.Atoi(md.Get(headerXRateRetry)[0])
		log.FromContext(ctx).Debugf("Available: %d, Reset: %d, Retry: %d", available, reset, retry)

		if err == nil || !isRetriable(err, callOpts) {
			return err
		}

		// retries after the initial request
		for attempt := uint(1); attempt <= callOpts.max; attempt++ {
			retryTimeout := callOpts.timeout
			if callOpts.enableXrateHeader {
				if headerTimeout := getHeaderLimiterTimeout(ctx, md); headerTimeout > 0 {
					retryTimeout = headerTimeout
				}
			}

			logger.WithField("attempt", attempt).Infof("Failed request, waiting %v until next attempt", retryTimeout)
			err = waitRetryBackoff(ctx, retryTimeout)
			if err != nil {
				logger.WithError(err).Debug("An unexpected error occurred while in the timeout for the next request retry")
				return err
			}

			callCtx := context.WithValue(ctx, AttemptMetadataKey, attempt)
			err = invoker(callCtx, method, req, reply, cc, append(opts, grpc.Header(&md))...)
			if err == nil {
				return nil
			}
			logger.WithField("attempt", attempt).WithError(err).Debug("Failed request retry")

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

func waitRetryBackoff(ctx context.Context, timeout time.Duration) error {
	timer := time.NewTicker(timeout)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
	}
	return nil
}

func hasRateLimitHeaders(md metadata.MD) bool {
	return len(md.Get(headerXRateLimit)) > 0 &&
		len(md.Get(headerXRateAvailable)) > 0 &&
		len(md.Get(headerXRateReset)) > 0 &&
		len(md.Get(headerXRateRetry)) > 0
}

func getHeaderLimiterTimeout(ctx context.Context, md metadata.MD) time.Duration {
	if !hasRateLimitHeaders(md) {
		return -1
	}
	var available, reset, retry int
	available, _ = strconv.Atoi(md.Get(headerXRateAvailable)[0])
	reset, _ = strconv.Atoi(md.Get(headerXRateReset)[0])
	retry, _ = strconv.Atoi(md.Get(headerXRateRetry)[0])

	log.FromContext(ctx).Debugf("Available: %d, Reset: %d, Retry: %d", available, reset, retry)

	if available == 0 {
		return time.Duration(reset) * time.Second
	}

	return time.Duration(retry) * time.Second

	//// When setting the retry header with a value bigger than zero the limiter signals that the amount of requests
	//// permitted have been reached, meaning that the client should wait until the reset
	//if retry > 0 {
	//	// Waits until reset is finished
	//	return time.Duration(retry) * time.Second
	//}

	//// Defines timeout as the amount of time remaining until the reset divided by the amount of requests tha can still be
	//// made, that allows to disperse the requests in a more even manner. For the sake of avoiding multiple concurrent
	//// retries to have the same timeout jitter is used.
	//timeout := time.Minute - (time.Duration(reset) * time.Second)
	//return backoffutils.JitterUp(timeout, 0.1)
}
