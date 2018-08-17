// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package assertions

import (
	"context"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestShouldHaveParentContext(t *testing.T) {
	for i, tc := range []struct {
		Actual,
		Expected interface{}
		Test func(actual interface{}, expected ...interface{}) string
	}{
		{
			Actual:   "string",
			Expected: context.Background(),
			Test:     should.NotBeEmpty,
		},
		{
			Actual:   context.Background(),
			Expected: "string",
			Test:     should.NotBeEmpty,
		},
		{
			Actual:   context.Background(),
			Expected: context.Background(),
			Test:     should.NotBeEmpty,
		},
		{
			Actual:   context.WithValue(context.Background(), struct{}{}, struct{}{}),
			Expected: context.Background(),
			Test:     should.BeEmpty,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			msg := ShouldHaveParentContext(tc.Actual, tc.Expected)
			a.So(msg, tc.Test)
		})
	}
}
