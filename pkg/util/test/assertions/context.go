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
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/util/test"
)

const (
	needContext             = "This assertion requires context.Context as comparison type (you provided %T)."
	shouldHaveParentContext = "Expected context to have parent '%v' (but it didn't)!"
)

// ShouldHaveParentContext takes as argument a context.Context and context.Context.
// If the arguments are valid and the actual context has the expected context as
// parent, this function returns an empty string.
// Otherwise, it returns a string describing the error.
func ShouldHaveParentContext(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf(needExactValues, 1, len(expected))
	}

	ctx, ok := actual.(context.Context)
	if !ok {
		return fmt.Sprintf(needContext, actual)
	}

	parent, ok := expected[0].(context.Context)
	if !ok {
		return fmt.Sprintf(needContext, expected[0])
	}

	if !test.ContextHasParent(ctx, parent) {
		return fmt.Sprintf(shouldHaveParentContext, parent)
	}
	return success
}
