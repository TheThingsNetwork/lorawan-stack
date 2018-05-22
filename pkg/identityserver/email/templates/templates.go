// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package templates provides a generic interface that templates used in emails
// must implements.
//
// Templates used by the email provider are defined here.
package templates

import (
	"bytes"
	"fmt"
	"html/template"
)

var templates = make(templatesRegistry)

type templatesRegistry map[string]*template.Template

func (r templatesRegistry) Register(templateName, subject, body string) {
	t, err := template.New("").Parse(subject)
	if err != nil {
		panic(err)
	}
	r[fmt.Sprintf("subject:%s", templateName)] = t

	t, err = template.New("").Parse(body)
	if err != nil {
		panic(err)
	}
	r[fmt.Sprintf("body:%s", templateName)] = t
}

func (r templatesRegistry) Subject(templateName string) *template.Template {
	return r[fmt.Sprintf("subject:%s", templateName)]
}

func (r templatesRegistry) Body(templateName string) *template.Template {
	return r[fmt.Sprintf("body:%s", templateName)]
}

// Template is the interface of email templates.
type Template interface {
	// GetName returns the template's name.
	GetName() string

	// Render renders the subject and message of the template returning it.
	Render() (string, string, error)
}

// render renders subject and message using the given data.
func render(templateName string, data interface{}) (string, string, error) {
	// TODO(gomezjdaniel): add styles to the HTML version.
	buf := new(bytes.Buffer)
	if err := templates.Subject(templateName).Execute(buf, data); err != nil {
		return "", "", err
	}
	subject := buf.String()

	buf.Reset()

	if err := templates.Body(templateName).Execute(buf, data); err != nil {
		return "", "", err
	}

	return subject, buf.String(), nil
}
