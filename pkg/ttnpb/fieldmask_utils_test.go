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

package ttnpb_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestTopLevelFields(t *testing.T) {
	a := assertions.New(t)
	paths := []string{
		"a",
		"b",
		"b.c",
		"b.c.d",
	}
	a.So(ttnpb.TopLevelFields(paths), should.Resemble, []string{"a", "b"})
}

func TestHasOnlyAllowedFields(t *testing.T) {
	a := assertions.New(t)
	allowed := []string{
		"a",
		"b.c",
		"d.e",
	}

	{
		requested := []string{
			"a",
			"b.c",
			"b.c.d", // lower level allowed
		}
		a.So(ttnpb.HasOnlyAllowedFields(requested, allowed...), should.BeTrue)
	}

	{
		requested := []string{
			"a",
			"e.f",
		}
		a.So(ttnpb.HasOnlyAllowedFields(requested, allowed...), should.BeFalse)
	}

	{
		requested := []string{
			"a",
			"d", // higher level not allowed
		}
		a.So(ttnpb.HasOnlyAllowedFields(requested, allowed...), should.BeFalse)
	}
}

func TestHasAnyField(t *testing.T) {
	a := assertions.New(t)
	requested := []string{
		"a",
		"b.c",
		"d",
	}
	a.So(ttnpb.HasAnyField(requested, "x", "a"), should.BeTrue)
	a.So(ttnpb.HasAnyField(requested, "x.y", "b"), should.BeFalse)
	a.So(ttnpb.HasAnyField(requested, "x", "b.c"), should.BeTrue)
	a.So(ttnpb.HasAnyField(requested, "x", "b.c.d"), should.BeTrue)
	a.So(ttnpb.HasAnyField(requested, "d"), should.BeTrue)
	a.So(ttnpb.HasAnyField(requested, "d.e", "b"), should.BeTrue)
}
