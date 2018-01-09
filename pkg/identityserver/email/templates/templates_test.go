// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ Renderer = new(DefaultRenderer)

func TestDefaultRenderer(t *testing.T) {
	a := assertions.New(t)

	tmpl := &Template{
		Subject: "Hi",
		Message: "<b>{{.name}}!</b>",
	}

	renderer := new(DefaultRenderer)
	subject, body, err := renderer.Render(tmpl, map[string]interface{}{
		"name": "john",
	})
	a.So(err, should.BeNil)
	a.So(subject, should.Equal, "Hi")
	a.So(body, should.Equal, "<b>john!</b>")
}
