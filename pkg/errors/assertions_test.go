// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var testDescriptor = &ErrDescriptor{
	MessageFormat: "Test error",
	Code:          42,
}

func init() {
	testDescriptor.Register()
}

func TestShouldDescribe(t *testing.T) {
	a := assertions.New(t)

	// Happy flow.
	a.So(ShouldDescribe(testDescriptor.New(nil), testDescriptor), should.BeEmpty)
	a.So(ShouldNotDescribe(testDescriptor.New(nil), testDescriptor), should.NotBeEmpty)

	// Unknown error.
	a.So(ShouldDescribe(fmt.Errorf("unknown error"), testDescriptor), should.NotBeEmpty)
	a.So(ShouldNotDescribe(fmt.Errorf("unknown error"), testDescriptor), should.BeEmpty)

	// Wrong namespace or code.
	a.So(ShouldDescribe(New("test"), testDescriptor), should.NotBeEmpty)
	a.So(ShouldNotDescribe(New("test"), testDescriptor), should.BeEmpty)
}
