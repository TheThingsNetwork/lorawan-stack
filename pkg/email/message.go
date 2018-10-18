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

package email

import (
	"bytes"
	"html/template"

	"github.com/jaytaylor/html2text"
)

// MessageTemplate is the template for an email message.
type MessageTemplate struct {
	Name            string
	SubjectTemplate *template.Template
	HTMLTemplate    *template.Template
	TextTemplate    *template.Template
}

// Execute the message template, rendering it into a Message.
func (m MessageTemplate) Execute(data interface{}) (*Message, error) {
	var buf bytes.Buffer
	out := Message{
		TemplateName: m.Name,
	}

	err := m.SubjectTemplate.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	out.Subject = buf.String()

	if m.HTMLTemplate != nil {
		buf.Reset()
		err = m.HTMLTemplate.Execute(&buf, data)
		if err != nil {
			return nil, err
		}
		out.HTMLBody = buf.String()
		// TODO: Optimize the HTML for email (with something like premailer).
	}

	if m.TextTemplate != nil {
		buf.Reset()
		err = m.TextTemplate.Execute(&buf, data)
		if err != nil {
			return nil, err
		}
		out.TextBody = buf.String()
	}

	if out.TextBody == "" && out.HTMLBody != "" {
		out.TextBody, err = html2text.FromString(out.HTMLBody, html2text.Options{PrettyTables: true})
		if err != nil {
			return nil, err
		}
	}

	return &out, nil
}

// Message for sending over email.
type Message struct {
	TemplateName string

	RecipientName    string
	RecipientAddress string

	Subject  string
	HTMLBody string
	TextBody string
}
