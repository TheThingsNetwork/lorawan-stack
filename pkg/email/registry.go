// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package email

import (
	"fmt"
	"html/template"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
)

// TemplateRegistry contains multiple email templates, identified by name.
type TemplateRegistry struct {
	fetcher  fetch.Interface
	registry sync.Map
	shared   *template.Template
}

// NewTemplateRegistry returns a new template registry that uses the given fetcher.
func NewTemplateRegistry(fetcher fetch.Interface, includes ...string) (*TemplateRegistry, error) {
	r := &TemplateRegistry{
		fetcher: fetcher,
		shared:  template.New("").Funcs(defaultFuncs),
	}
	for _, include := range includes {
		data, err := fetcher.File(include)
		if err != nil {
			return nil, err
		}
		shared, err := r.shared.New(include).Parse(string(data))
		if err != nil {
			return nil, err
		}
		r.shared = shared
	}
	return r, nil
}

type registeredTemplate struct {
	m     *MessageTemplate
	err   error
	ready chan struct{}
}

func (r *TemplateRegistry) getTemplate(data MessageData) (m *MessageTemplate, err error) {
	name := data.TemplateName()
	registeredI, ok := r.registry.LoadOrStore(name, &registeredTemplate{ready: make(chan struct{})})
	registered := registeredI.(*registeredTemplate)
	if ok {
		<-registered.ready
		return registered.m, registered.err
	}

	defer func() {
		registered.m, registered.err = m, err
		close(registered.ready)
	}()

	m = &MessageTemplate{Name: name}
	subject, html, text := data.DefaultTemplates()

	if r.fetcher != nil {
		subjectBytes, err := r.fetcher.File(fmt.Sprintf("%s.subject.txt", name))
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if len(subjectBytes) > 0 {
			subject = string(subjectBytes)
		}
		htmlBytes, err := r.fetcher.File(fmt.Sprintf("%s.html", name))
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if len(htmlBytes) > 0 {
			html = string(htmlBytes)
		}
		textBytes, err := r.fetcher.File(fmt.Sprintf("%s.txt", name))
		if err != nil && !errors.IsNotFound(err) {
			return nil, err
		}
		if len(textBytes) > 0 {
			text = string(textBytes)
		}
	}

	template, err := r.shared.Clone()
	if err != nil {
		return nil, err
	}

	m.SubjectTemplate, err = template.New(name + "_subject").Parse(subject)
	if err != nil {
		return nil, err
	}
	m.HTMLTemplate, err = template.New(name + "_html_body").Parse(html)
	if err != nil {
		return nil, err
	}
	if text != "" {
		m.TextTemplate, err = template.New(name + "_text_body").Parse(text)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// Render message template data into a message.
func (r *TemplateRegistry) Render(data MessageData) (*Message, error) {
	template, err := r.getTemplate(data)
	if err != nil {
		return nil, err
	}
	message, err := template.Execute(data)
	if err != nil {
		return nil, err
	}
	message.RecipientName, message.RecipientAddress = data.Recipient()
	return message, nil
}

// MessageData interface contains everything we need to create an email.Message.
// The DefaultTemplates should be able to execute using the MessageData itself.
type MessageData interface {
	TemplateName() string
	Recipient() (name, address string)
	DefaultTemplates() (subject, html, text string)
}
