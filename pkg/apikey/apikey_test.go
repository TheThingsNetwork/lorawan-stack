// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var (
	tenants = []string{
		"foo.bar.baz",
		"id.thethings.network",
		"a.very.long.tenant.id.that.is.really.long",
	}
)

func TestGenerateKey(t *testing.T) {
	a := assertions.New(t)

	// test good apps
	for _, tenant := range tenants {
		key := GenerateAPIKey(tenant)
		ten, err := KeyTenant(key)
		a.So(err, should.BeNil)

		fmt.Println(key)
		a.So(ten, should.Equal, tenant)
	}
}
