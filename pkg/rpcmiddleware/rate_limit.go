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

package rpcmiddleware

import (
	"context"
	"strconv"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"google.golang.org/grpc/metadata"
)

const (
	headerXRateLimit     = "x-rate-limit-limit"
	headerXRateAvailable = "x-rate-limit-available"
	headerXRateReset     = "x-rate-limit-reset"
	headerXRateRetry     = "x-rate-limit-retry"
)

func hasRateLimitHeaders(md metadata.MD) bool {
	return len(md.Get(headerXRateLimit)) > 0 &&
		len(md.Get(headerXRateAvailable)) > 0 &&
		len(md.Get(headerXRateReset)) > 0 &&
		len(md.Get(headerXRateRetry)) > 0
}

// GetTimeoutFromRateLimit returns a time.Duration that can be used to determine how much time should be waited before
// attempting to make another request. If the xrate headers specify that there are available requests, then the return
// value will be time until the limiter reset divided by the amount of available requests.
func GetTimeoutFromRateLimit(ctx context.Context) time.Duration {
	logger := log.FromContext(ctx)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || !hasRateLimitHeaders(md) {
		logger.Debug("Failed to obtain headers from the metadata")
		return -1
	}

	available, err := strconv.Atoi(md.Get(headerXRateAvailable)[0])
	if err != nil {
		logger.WithError(err).Debug("Failed to parse xrate available header")
		return -1
	}
	logger.Debug("Available:", available)

	if available == 0 {
		retryTimeout, err := time.ParseDuration(md.Get(headerXRateRetry)[0])
		if err != nil {
			logger.WithError(err).Debug("Failed to parse xrate retry header")
			return -1
		}
		logger.Debug("RetryTimeout:", retryTimeout*time.Second)

		return retryTimeout * time.Second
	}

	untilReset, err := time.ParseDuration(md.Get(headerXRateReset)[0])
	if err != nil {
		logger.WithError(err).Debug("Failed to parse xrate reset header")
		return -1
	}

	return (untilReset * time.Second) / time.Duration(available)
}
