// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package templates

import (
	"bytes"
	"html/template"
)

// Template is the interface of those things that are an email template.
type Template interface {
	// GetName returns the template's name.
	GetName() string

	// Render renders the subject and message of the template returning it.
	Render() (string, string, error)
}

// render renders subject and message using the given data.
func render(subject, message string, data interface{}) (string, string, error) {
	// TODO(gomezjdaniel): add styles to the HTML version.
	t, err := template.New("").Parse(subject)
	if err != nil {
		return "", "", err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", "", err
	}
	subject = buf.String()

	buf.Reset()

	t, err = template.New("").Parse(message)
	if err != nil {
		return "", "", err
	}

	if err := t.Execute(buf, data); err != nil {
		return "", "", err
	}

	return subject, buf.String(), nil
}
