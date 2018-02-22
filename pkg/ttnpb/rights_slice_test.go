// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
