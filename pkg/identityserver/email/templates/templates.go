// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

import (
	"bytes"
	"text/template"
)

var _ Renderer = new(DefaultRenderer)

// Template is the type that describes an email template.
type Template struct {
	Name    string
	Subject string
	Message string
}

// Renderer is the interface that describes an email template renderer.
type Renderer interface {
	// Render compiles a template with the provided data and returns the
	// subject and the content.
	Render(tmpl *Template, data interface{}) (string, string, error)
}

// DefaultRenderer is a renderer that relies on the `text/template` package to
// compile the template. It only compiles the `Message` field of the template.
type DefaultRenderer struct{}

// Render implements Renderer.
func (r DefaultRenderer) Render(tmpl *Template, data interface{}) (string, string, error) {
	t := template.New("")
	t, _ = t.Parse(tmpl.Message)

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", "", err
	}

	return tmpl.Subject, buf.String(), nil
}
