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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// Validator is a method that validates if an error should trigger the request retry.
type Validator func(error) bool

type options struct {
	max            uint
	timeout        time.Duration
	validators     []Validator
	enableMetadata bool
	jitter         float64
}

// Options is an option for the rpcretry clients.
type Option func(*options)

var (
	// Default Validators is a set of functions that validate errors that should trigger a retry of the request.
	DefaultValidators = []Validator{
		Validator(errors.IsResourceExhausted),
		Validator(errors.IsUnavailable),
	}

	defaultOptions = &options{
		max:            0,
		timeout:        100 * time.Millisecond,
		validators:     DefaultValidators,
		enableMetadata: true,
		jitter:         0.0,
	}
)

// WithMax sets the value of the maximum amount of times a request will be retried.
func WithMax(m uint) Option {
	return func(opt *options) {
		opt.max = m
	}
}

// WithDefaultTimeout sets the default timeout between request retries.
func WithDefaultTimeout(t time.Duration) Option {
	return func(opt *options) {
		opt.timeout = t
	}
}

// WithValidator sets validators that will be evaluated when validating if a request should be retried.
func WithValidator(validators ...Validator) Option {
	return func(opt *options) {
		opt.validators = validators
	}
}

// UseMetadata establishes if the x-rate-limit headers will be used to dinamically calculate the timeout between requests.
func UseMetadata(b bool) Option {
	return func(opt *options) {
		opt.enableMetadata = b
	}
}

// WithJitter determines the value of the fraciton used to create the deviation in the timeout between the requests.
func WithJitter(f float64) Option {
	return func(opt *options) {
		opt.jitter = f
	}
}
