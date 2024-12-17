// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"context"
	"fmt"
	"html/template"
	"io/fs"
	"sort"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/jaytaylor/html2text"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type templateRegistryCtxKeyType struct{}

var templateRegistryCtxKey templateRegistryCtxKeyType

func templateRegistryFromContext(ctx context.Context) (TemplateRegistry, bool) {
	reg, ok := ctx.Value(templateRegistryCtxKey).(TemplateRegistry)
	return reg, ok
}

func newContextWithTemplateRegistry(parent context.Context, reg TemplateRegistry) context.Context {
	return context.WithValue(parent, templateRegistryCtxKey, reg)
}

// TemplateRegistry keeps track of email templates.
type TemplateRegistry interface {
	RegisteredTemplates() []*ttnpb.NotificationType
	GetTemplate(ctx context.Context, name string) *Template
}

// NewTemplateRegistry returns a new empty template registry.
func NewTemplateRegistry() MapTemplateRegistry {
	return make(MapTemplateRegistry)
}

// MapTemplateRegistry is a template registry implementation.
type MapTemplateRegistry map[string]*Template

// RegisterTemplate registers a template.
func (reg MapTemplateRegistry) RegisterTemplate(tmpl *Template) {
	reg[tmpl.Name] = tmpl
}

// RegisteredTemplates returns a sorted list of the names of all registered templates.
func (reg MapTemplateRegistry) RegisteredTemplates() []string {
	names := make([]string, 0, len(reg))
	for name := range reg {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetTemplate returns a registered template from the registry.
func (reg MapTemplateRegistry) GetTemplate(_ context.Context, name string) *Template {
	return reg[name]
}

var defaultTemplateRegistry = make(MapTemplateRegistry)

// RegisterTemplate registers a template on the default registry.
func RegisterTemplate(tmpl *Template) {
	defaultTemplateRegistry.RegisterTemplate(tmpl)
}

// RegisteredTemplates returns the names of the registered templates in the default registry.
func RegisteredTemplates() []string {
	return defaultTemplateRegistry.RegisteredTemplates()
}

// GetTemplate returns a registered template from the registry in the context (if available), otherwise falling back to the default registry.
func GetTemplate(ctx context.Context, name string) *Template {
	if reg, ok := templateRegistryFromContext(ctx); ok {
		if tmpl := reg.GetTemplate(ctx, name); tmpl != nil {
			return tmpl
		}
	}
	return defaultTemplateRegistry.GetTemplate(ctx, name)
}

// Template is the template for an email message.
type Template struct {
	Name            string
	SubjectTemplate *template.Template
	HTMLTemplate    *template.Template
	TextTemplate    *template.Template
}

var shared = template.New("").Funcs(sprig.FuncMap()).Funcs(defaultFuncs)

// FSTemplate defines the template files to parse from the file system.
type FSTemplate struct {
	SubjectTemplate      string
	HTMLTemplateBaseFile string
	HTMLTemplateFile     string
	TextTemplateBaseFile string
	TextTemplateFile     string
	IncludePatterns      []string
}

// NewTemplateFS parses a new email template by reading files on fsys.
func NewTemplateFS(fsys fs.FS, name string, opts FSTemplate) (*Template, error) {
	var (
		shared = template.Must(shared.Clone())
		err    error
		tmpl   = Template{Name: name}
	)
	if len(opts.IncludePatterns) > 0 {
		shared, err = shared.ParseFS(fsys, opts.IncludePatterns...)
		if err != nil {
			return nil, err
		}
	}
	tmpl.SubjectTemplate, err = template.Must(shared.Clone()).Parse(opts.SubjectTemplate)
	if err != nil {
		return nil, err
	}
	if opts.HTMLTemplateFile != "" {
		htmlTemplate := template.Must(shared.Clone())
		if opts.HTMLTemplateBaseFile != "" {
			htmlTemplate, err = htmlTemplate.ParseFS(fsys, opts.HTMLTemplateBaseFile)
			if err != nil {
				return nil, err
			}
		}
		htmlTemplate, err = htmlTemplate.ParseFS(fsys, opts.HTMLTemplateFile)
		if err != nil {
			return nil, err
		}
		if opts.HTMLTemplateBaseFile != "" {
			tmpl.HTMLTemplate = htmlTemplate.Lookup(opts.HTMLTemplateBaseFile)
		} else {
			tmpl.HTMLTemplate = htmlTemplate.Lookup(opts.HTMLTemplateFile)
		}
	}
	if opts.TextTemplateFile != "" {
		textTemplate := template.Must(shared.Clone())
		if opts.TextTemplateBaseFile != "" {
			textTemplate, err = textTemplate.ParseFS(fsys, opts.TextTemplateBaseFile)
			if err != nil {
				return nil, err
			}
		}
		textTemplate, err = textTemplate.ParseFS(fsys, opts.TextTemplateFile)
		if err != nil {
			return nil, err
		}
		if opts.TextTemplateBaseFile != "" {
			tmpl.TextTemplate = textTemplate.Lookup(opts.TextTemplateBaseFile)
		} else {
			tmpl.TextTemplate = textTemplate.Lookup(opts.TextTemplateFile)
		}
	}
	return &tmpl, nil
}

// TemplateData is the minimal interface Execute needs to render an email template.
type TemplateData interface {
	Network() *NetworkConfig
	ConsoleURL() string
	Receiver() *ttnpb.User
	ReceiverName() string
}

// TemplateDataBuilder is used to extend TemplateData.
type TemplateDataBuilder func(context.Context, TemplateData) (TemplateData, error)

// NewTemplateData returns new template data.
func NewTemplateData(networkConfig *NetworkConfig, receiver *ttnpb.User) TemplateData {
	return &templateData{
		networkConfig: networkConfig,
		receiver:      receiver,
	}
}

type templateData struct {
	networkConfig *NetworkConfig
	receiver      *ttnpb.User
}

func (d *templateData) Network() *NetworkConfig { return d.networkConfig }
func (d *templateData) ConsoleURL() string      { return d.networkConfig.ConsoleURL }
func (d *templateData) Receiver() *ttnpb.User   { return d.receiver }
func (d *templateData) ReceiverName() string {
	if name := d.Receiver().GetName(); name != "" {
		return name
	}
	if id := d.Receiver().GetIds().GetUserId(); id != "" {
		return id
	}
	return "user"
}

// Execute the message template, rendering it into a Message.
func (m Template) Execute(data TemplateData) (*Message, error) {
	var buf bytes.Buffer
	out := Message{
		TemplateName:     m.Name,
		RecipientName:    data.Receiver().GetName(),
		RecipientAddress: data.Receiver().GetPrimaryEmailAddress(),
	}

	err := m.SubjectTemplate.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute subject template: %w", err)
	}
	out.Subject = strings.TrimSpace(buf.String())

	if m.HTMLTemplate != nil {
		buf.Reset()
		err = m.HTMLTemplate.Execute(&buf, data)
		if err != nil {
			return nil, fmt.Errorf("failed to execute HTML template: %w", err)
		}
		out.HTMLBody = buf.String()
	}

	if m.TextTemplate != nil {
		buf.Reset()
		err = m.TextTemplate.Execute(&buf, data)
		if err != nil {
			return nil, fmt.Errorf("failed to execute text template: %w", err)
		}
		out.TextBody = buf.String()
	}

	if out.TextBody == "" && out.HTMLBody != "" {
		out.TextBody, err = html2text.FromString(out.HTMLBody, html2text.Options{PrettyTables: true})
		if err != nil {
			return nil, fmt.Errorf("failed to convert HTML to text: %w", err)
		}
	}

	return &out, nil
}
