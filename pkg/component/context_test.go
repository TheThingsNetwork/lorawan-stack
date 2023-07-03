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

package component_test

import (
	"context"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestContextDecoupling(t *testing.T) {
	a := assertions.New(t)

	c, err := component.New(test.GetLogger(t), &component.Config{})
	a.So(err, should.BeNil)

	dl := time.Now().Add(1 * time.Second)

	ctx := test.Context()
	ctx = context.WithValue(ctx, "key", "value")
	ctx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	ctx = c.FromRequestContext(ctx)
	deadline, set := ctx.Deadline()
	a.So(deadline, should.NotEqual, dl)
	a.So(set, should.BeFalse)

	select {
	case <-ctx.Done():
		t.Fatal("Context expired")
	case <-time.After(time.Second):
		a.So(ctx.Err(), should.BeNil)
	}

	val := ctx.Value("key")
	strVal, ok := val.(string)
	a.So(ok, should.BeTrue)
	a.So(strVal, should.Equal, "value")
}
