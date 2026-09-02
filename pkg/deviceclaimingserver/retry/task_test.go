// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

package retry_test

import (
	"fmt"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/retry"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var errTestFunction = errors.DefineResourceExhausted("test_function_failed", "test function failed")

func TestTask(t *testing.T) { // nolint:paralleltest
	var attempt int

	a, ctx := test.New(t)
	for _, tc := range []struct { // nolint:paralleltest
		Name           string
		f              func() (bool, error)
		ErrorAssertion func(err error) bool
	}{
		{
			Name: "CompletedSuccessfully",
			f: func() (bool, error) {
				return false, nil
			},
		},
		{
			Name: "CompletedWithError",
			f: func() (bool, error) {
				return false, errTestFunction.New()
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsResourceExhausted(err)
			},
		},
		{
			Name: "Timeout",
			f: func() (bool, error) {
				return true, nil
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsInternal(err)
			},
		},
		{
			Name: "CompletedSuccessfullySecondAttempt",
			f: func() (bool, error) {
				if attempt == 0 {
					attempt++
					return true, nil
				}
				attempt = 0 // reset for the next case
				return false, nil
			},
		},
		{
			Name: "CompletedWithErrorSecondAttempt",
			f: func() (bool, error) {
				if attempt == 0 {
					attempt++
					return true, nil
				}
				return false, errTestFunction.New()
			},
			ErrorAssertion: func(err error) bool {
				return errors.IsResourceExhausted(err)
			},
		},
	} {
		t.Run(fmt.Sprintf("Setup/%s", tc.Name), func(t *testing.T) {
			testTask := retry.Task{
				Name:        tc.Name,
				F:           tc.f,
				WaitTime:    (1 << 3) * test.Delay,
				MaxAttempts: 3,
				Jitter:      0.2,
			}
			err := testTask.Do(ctx)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(err), should.BeTrue)
			}
		})
	}
}
