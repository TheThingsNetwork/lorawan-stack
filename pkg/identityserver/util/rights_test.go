// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package util

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRightsIntersection(t *testing.T) {
	a := assertions.New(t)

	{
		b := []ttnpb.Right{ttnpb.Right(5), ttnpb.Right(1)}
		c := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(10)}
		a.So(RightsIntersection(b, c), should.Resemble, []ttnpb.Right{ttnpb.Right(1)})
	}

	{
		b := []ttnpb.Right{ttnpb.Right(54), ttnpb.Right(12)}
		c := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(10)}
		a.So(RightsIntersection(b, c), should.Resemble, []ttnpb.Right{})
	}
}

func TestRightsDifference(t *testing.T) {
	a := assertions.New(t)

	{
		b := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)}
		c := []ttnpb.Right{ttnpb.Right(2), ttnpb.Right(3)}
		a.So(RightsDifference(b, c), should.Resemble, []ttnpb.Right{ttnpb.Right(1)})
	}

	{
		b := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)}
		c := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)}
		a.So(RightsDifference(b, c), should.Resemble, []ttnpb.Right{})
	}

	{
		b := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)}
		c := []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2), ttnpb.Right(3)}
		a.So(RightsDifference(b, c), should.Resemble, []ttnpb.Right{})
	}
}
