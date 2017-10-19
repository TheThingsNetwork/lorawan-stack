// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestValidate(t *testing.T) {
	a := assertions.New(t)

	a.So(Field("", NotRequired, MinLength(5)), should.BeNil)
	a.So(Field("abc", NotRequired, MinLength(5)), should.NotBeNil)
	a.So(Field("abc@abc.com", NotRequired, MinLength(5)), should.BeNil)

	a.So(All(
		Field("", NotRequired, MinLength(10)),
		Field("alice", ID),
	), should.BeNil)

	a.So(All(
		Field("", Required).DescribeFieldName("LOL"),
		Field("", Email).DescribeFieldName("Email"),
		Field("", NotRequired),
		Field("foo@bar.com", NotRequired, Email),
		Field("foo-app", ID),
	), should.NotBeNil)

	a.So(All(
		Field("", Required).DescribeFieldName("Whatever"),
		Field("foo-app", ID),
	), should.NotBeNil)
}
