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
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockResource struct {
	key     string
	classes []string
}

func (r *mockResource) Key() string       { return r.key }
func (r *mockResource) Classes() []string { return r.classes }

func TestRateLimit(t *testing.T) {
	a := assertions.New(t)

	limiter, err := ratelimit.New(test.Context(), config.RateLimiting{
		Profiles: []config.RateLimitingProfile{
			{
				Name:         "Default profile",
				MaxPerMin:    maxRate,
				MaxBurst:     maxRate,
				Associations: []string{"default"},
			},
			{
				Name:         "Multiple associations",
				MaxPerMin:    maxRate,
				MaxBurst:     maxRate,
				Associations: []string{"assoc1", "assoc2"},
			},
			{
				Name:         "Override",
				MaxPerMin:    overrideRate,
				MaxBurst:     overrideRate,
				Associations: []string{"override"},
			},
		},
	}, config.BlobConfig{}, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	t.Run("Limit", func(t *testing.T) {
		for _, resource := range []*mockResource{
			{"key1", []string{"default"}},
			{"key2", []string{"default"}},
		} {
			for i := uint(0); i < maxRate; i++ {
				limit, result := limiter.RateLimit(resource)

				a.So(limit, should.BeFalse)
				a.So(result.Limit, should.Equal, maxRate)
				a.So(result.Remaining, should.Equal, maxRate-i-1)
			}

			limit, result := limiter.RateLimit(resource)
			a.So(limit, should.BeTrue)
			a.So(result.Limit, should.Equal, maxRate)
			a.So(result.Remaining, should.Equal, 0)
			a.So(result.ResetAfter, should.NotBeZeroValue)
			a.So(result.RetryAfter, should.NotBeZeroValue)
		}
	})

	t.Run("Profile", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			classes  []string
			validate func(t *testing.T, limit bool, result ratelimit.Result)
		}{
			{
				name:    "Profile",
				classes: []string{"override"},
				validate: func(t *testing.T, limit bool, result ratelimit.Result) {
					assertions.New(t).So(result.Limit, should.Equal, overrideRate)
				},
			},
			{
				name:    "IgnoreMissing",
				classes: []string{"unknown-class"},
				validate: func(t *testing.T, limit bool, result ratelimit.Result) {
					a := assertions.New(t)
					a.So(limit, should.BeFalse)
					a.So(result.IsZero(), should.BeTrue)
				},
			},
			{
				name:    "Priority",
				classes: []string{"unknown-class", "default", "override"},
				validate: func(t *testing.T, limit bool, result ratelimit.Result) {
					assertions.New(t).So(result.Limit, should.Equal, maxRate)
				},
			},
			{
				name:    "Multiple",
				classes: []string{"assoc2"},
				validate: func(t *testing.T, limit bool, result ratelimit.Result) {
					assertions.New(t).So(result.Limit, should.Equal, maxRate)
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				resource := &mockResource{key: "key", classes: tc.classes}
				limit, result := limiter.RateLimit(resource)
				tc.validate(t, limit, result)
			})
		}
	})

	t.Run("Require", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			limiter   *mockLimiter
			assertErr func(error) bool
		}{
			{"Pass", &mockLimiter{}, func(err error) bool { return err == nil }},
			{"Limit", &mockLimiter{limit: true}, errors.IsResourceExhausted},
		} {
			t.Run(tc.name, func(t *testing.T) {
				a := assertions.New(t)
				resource := &mockResource{key: "resource", classes: []string{"one", "two"}}

				a.So(tc.assertErr(ratelimit.Require(tc.limiter, resource)), should.BeTrue)
			})
		}
	})

	t.Run("ExternalConfig", func(t *testing.T) {
		conf := config.RateLimiting{
			ConfigSource: "directory",
			Directory:    "testdata",
			Memory: config.RateLimitingMemory{
				MaxSize: 1024,
			},
			Profiles: []config.RateLimitingProfile{
				{
					MaxPerMin:    100,
					Associations: []string{"assoc1"},
				},
			},
		}

		a := assertions.New(t)
		limiter, err := ratelimit.New(test.Context(), conf, config.BlobConfig{}, nil)
		a.So(err, should.BeNil)

		resource := &mockResource{key: "key", classes: []string{"assoc1"}}
		limit, result := limiter.RateLimit(resource)
		a.So(limit, should.BeFalse)
		a.So(result.Limit, should.Equal, 200)
	})
}
