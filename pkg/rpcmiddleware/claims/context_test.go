// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package claims

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestClaimsContext(t *testing.T) {
	a := assertions.New(t)

	c := FromContext(context.Background())
	a.So(c, should.Resemble, new(auth.Claims))

	c = &auth.Claims{
		Client: "foo",
	}
	ctx := NewContext(context.Background(), c)
	a.So(FromContext(ctx), should.Resemble, c)
}
