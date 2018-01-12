// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRender(t *testing.T) {
	a := assertions.New(t)

	subject, message, err := render("Hello {{.Name}}", "<b>{{.Name}}!</b>", struct {
		Name string
	}{"john"})
	a.So(err, should.BeNil)
	a.So(subject, should.Equal, "Hello john")
	a.So(message, should.Equal, "<b>john!</b>")
}
