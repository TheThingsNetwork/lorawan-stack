// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package rpclog

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

func TestShouldSuppressLogEvaluatesCorrectly(t *testing.T) {
	ignoredError := errors.Define("ignored", "ignored")
	nonIgnoredError := errors.Define("non_ignored", "non_ignored")
	ignoredErrorKey := ignoredError.FullName()

	tests := []struct {
		inputCfg methodLogConfig
		inputErr error
		expected bool
	}{
		// test behavior when no error occurs
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: true,
				IgnoredErrors: map[string]struct{}{},
			},
			inputErr: nil,
			expected: true,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: true,
				IgnoredErrors: map[string]struct{}{
					ignoredErrorKey: {},
				},
			},
			inputErr: nil,
			expected: true,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: false,
				IgnoredErrors: map[string]struct{}{},
			},
			inputErr: nil,
			expected: false,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: false,
				IgnoredErrors: map[string]struct{}{
					ignoredErrorKey: {},
				},
			},
			inputErr: nil,
			expected: false,
		},
		// test behavior when an error occurs that is ignored
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: true,
				IgnoredErrors: map[string]struct{}{},
			},
			inputErr: ignoredError,
			expected: false,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: true,
				IgnoredErrors: map[string]struct{}{
					ignoredErrorKey: {},
				},
			},
			inputErr: ignoredError,
			expected: true,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: false,
				IgnoredErrors: map[string]struct{}{},
			},
			inputErr: ignoredError,
			expected: false,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: false,
				IgnoredErrors: map[string]struct{}{
					ignoredErrorKey: {},
				},
			},
			inputErr: ignoredError,
			expected: true,
		},
		// test behavior when an error occurs that is not ignored
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: true,
				IgnoredErrors: map[string]struct{}{},
			},
			inputErr: nonIgnoredError,
			expected: false,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: true,
				IgnoredErrors: map[string]struct{}{
					ignoredErrorKey: {},
				},
			},
			inputErr: nonIgnoredError,
			expected: false,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: false,
				IgnoredErrors: map[string]struct{}{},
			},
			inputErr: nonIgnoredError,
			expected: false,
		},
		{
			inputCfg: methodLogConfig{
				IgnoreSuccess: false,
				IgnoredErrors: map[string]struct{}{
					ignoredErrorKey: {},
				},
			},
			inputErr: nonIgnoredError,
			expected: false,
		},
	}

	for _, test := range tests {
		actual := shouldSuppressLog(test.inputCfg, test.inputErr)
		if actual != test.expected {
			t.Errorf("Expected %t, got %t", test.expected, actual)
		}
	}
}
