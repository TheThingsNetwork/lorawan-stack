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

package ttnpb

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestIntersectRights(t *testing.T) {
	a := assertions.New(t)

	{
		b := []Right{Right(5), Right(1)}
		c := []Right{Right(1), Right(10)}
		a.So(IntersectRights(b, c), should.Resemble, []Right{Right(1)})
	}

	{
		b := []Right{Right(54), Right(12)}
		c := []Right{Right(1), Right(10)}
		a.So(IntersectRights(b, c), should.Resemble, []Right{})
	}
}

func TestDifferenceRights(t *testing.T) {
	a := assertions.New(t)

	{
		b := []Right{Right(1), Right(2)}
		c := []Right{Right(2), Right(3)}
		a.So(DifferenceRights(b, c), should.Resemble, []Right{Right(1)})
	}

	{
		b := []Right{Right(1), Right(2)}
		c := []Right{Right(1), Right(2)}
		a.So(DifferenceRights(b, c), should.Resemble, []Right{})
	}

	{
		b := []Right{Right(1), Right(2)}
		c := []Right{Right(1), Right(2), Right(3)}
		a.So(DifferenceRights(b, c), should.Resemble, []Right{})
	}
}

func TestIncludesRights(t *testing.T) {
	a := assertions.New(t)

	a.So(IncludesRights([]Right{Right(1)}), should.BeTrue)
	a.So(IncludesRights([]Right{Right(1)}, Right(1)), should.BeTrue)
	a.So(IncludesRights([]Right{Right(1), Right(4)}, Right(4), Right(1)), should.BeTrue)

	a.So(IncludesRights([]Right{Right(1)}, Right(2)), should.BeFalse)
	a.So(IncludesRights([]Right{Right(1)}, Right(2), Right(3)), should.BeFalse)
	a.So(IncludesRights([]Right{Right(1), Right(2)}, Right(2), Right(3)), should.BeFalse)
}
