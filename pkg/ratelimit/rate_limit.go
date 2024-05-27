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

package ratelimit

import (
	"context"
	"fmt"

	"github.com/throttled/throttled/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// Interface can be used to rate limit access to a Resource.
type Interface interface {
	// RateLimit limits access on a Resource.
	//
	// The RateLimit operations returns true if the rate limit for the requested resource has been exceeded, along
	// with metadata for the rate limiting operation.
	RateLimit(resource Resource) (limit bool, result Result)
}

// NoopRateLimiter does not enforce any rate limits.
type NoopRateLimiter struct{}

// RateLimit implements ratelimit.Interface.
func (*NoopRateLimiter) RateLimit(Resource) (bool, Result) {
	return false, Result{}
}

type rateLimiter struct {
	ctx     context.Context
	limiter throttled.RateLimiterCtx
}

// RateLimit implements ratelimit.Interface.
func (l *rateLimiter) RateLimit(resource Resource) (bool, Result) {
	ok, result, err := l.limiter.RateLimitCtx(l.ctx, resource.Key(), 1)
	if err != nil {
		log.FromContext(l.ctx).
			WithError(err).
			WithFields(logFields(resource)).
			Error("Rate limiter failed")
		// NOTE: The missing return here is intentional - the underlying implementation
		// will still return a valid ok and result even on error.
	}
	return ok, Result{
		Limit:      result.Limit,
		Remaining:  result.Remaining,
		RetryAfter: result.RetryAfter,
		ResetAfter: result.ResetAfter,
	}
}

// muxRateLimiter is a ratelimit.Interface that supports multiple rate limiting profiles.
// If no rate limiting profile is set for a resource class, then no rate limits are applied.
type muxRateLimiter struct {
	defaultLimiter Interface
	limiters       map[string]Interface
}

// RateLimit implements ratelimit.Interface.
func (l *muxRateLimiter) RateLimit(resource Resource) (bool, Result) {
	for _, c := range resource.Classes() {
		if limiter, ok := l.limiters[c]; ok {
			return limiter.RateLimit(resource)
		}
	}
	return l.defaultLimiter.RateLimit(resource)
}

// Require checks that the rate limit for a Resource has not been exceeded.
func Require(limiter Interface, resource Resource) error {
	if limit, result := limiter.RateLimit(resource); limit {
		return errRateLimitExceeded.WithAttributes(
			"key", resource.Key(),
			"rate", result.Limit,
		)
	}
	return nil
}

// RateLimitKeyer can be implemented by request messages. If implemented, the
// string returned by the RateLimitKey() method is appended to the key used for
// rate limiting.
type RateLimitKeyer interface {
	RateLimitKey() string
}

func logFields(resource Resource) log.Fielder {
	fields := append(
		make([]any, 0, 2+2*len(resource.Classes())),
		"resource_key", resource.Key(),
	)
	for i, c := range resource.Classes() {
		fields = append(fields, fmt.Sprintf("resource_class_%d", i), c)
	}
	return log.Fields(fields...)
}
