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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

type Validator func(error) bool

type options struct {
	max               uint
	timeout           time.Duration
	validators        []Validator
	enableXrateHeader bool
}

type Option func(*options)

var (
	DefaultValidators = []Validator{
		Validator(errors.IsResourceExhausted),
		Validator(errors.IsUnavailable),
	}

	defaultOptions = &options{
		max:               0,
		timeout:           100 * time.Millisecond,
		validators:        DefaultValidators,
		enableXrateHeader: true,
	}
)

func WithMax(m uint) Option {
	return func(opt *options) {
		opt.max = m
	}
}

func WithDefaultTimeout(t time.Duration) Option {
	return func(opt *options) {
		opt.timeout = t
	}
}

func WithValidator(validators ...Validator) Option {
	return func(opt *options) {
		opt.validators = validators
	}
}

func DisableXRateHeader() Option {
	return func(opt *options) {
		opt.enableXrateHeader = false
	}
}
