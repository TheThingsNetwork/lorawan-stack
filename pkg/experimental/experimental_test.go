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

package experimental_test

import (
	"context"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestExperimentalFeatures(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	r := NewRegistry()

	ctx := NewContextWithRegistry(context.Background(), r)

	feature := DefineFeature("experimental.feature", false)
	a.So(feature.GetValue(ctx), should.BeFalse)
	a.So(AllFeatures(ctx)["experimental.feature"], should.BeFalse)
	a.So(feature.GetValue(context.Background()), should.BeFalse)
	a.So(AllFeatures(context.Background())["experimental.feature"], should.BeFalse)

	r.EnableFeatures("experimental.feature")
	a.So(feature.GetValue(ctx), should.BeTrue)
	a.So(AllFeatures(ctx)["experimental.feature"], should.BeTrue)
	a.So(feature.GetValue(context.Background()), should.BeFalse)
	a.So(AllFeatures(context.Background())["experimental.feature"], should.BeFalse)

	EnableFeatures("experimental.feature")
	r.DisableFeatures("experimental.feature")

	a.So(feature.GetValue(ctx), should.BeFalse)
	a.So(AllFeatures(ctx)["experimental.feature"], should.BeFalse)
	a.So(feature.GetValue(context.Background()), should.BeTrue)
	a.So(AllFeatures(context.Background())["experimental.feature"], should.BeTrue)
}
