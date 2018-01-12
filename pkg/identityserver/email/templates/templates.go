// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

import (
	"bytes"
	"text/template"
)

// Template is the interface of those things that are an email template.
type Template interface {
	// Name returns the template's name.
	Name() string

	// Render renders the subject and message of the template returning it.
	Render() (string, string, error)
}

// render renders subject and message using the given data.
func render(subject, message string, data interface{}) (string, string, error) {
	t := template.New("")
	t, _ = t.Parse(subject)

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", "", err
	}
	subject = buf.String()

	buf.Reset()

	t, _ = t.Parse(message)
	if err := t.Execute(buf, data); err != nil {
		return "", "", err
	}

	return subject, buf.String(), nil
}
