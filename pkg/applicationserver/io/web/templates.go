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

package web

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	ttnweb "go.thethings.network/lorawan-stack/pkg/web"
	"gopkg.in/yaml.v2"
)

const (
	yamlFetchErrorCache = 1 * time.Minute
	directoryBaseURL    = "/as/webhook-templates/static"
)

// TemplatesConfig defines the configuration for the webhook templates registry.
type TemplatesConfig struct {
	Static    map[string][]byte `name:"-"`
	Directory string            `name:"directory" description:"Retrieve the webhook templates from the filesystem"`
	URL       string            `name:"url" description:"Retrieve the webhook templates from a web server"`
}

// TemplateStore contains the webhook templates.
type TemplateStore struct {
	fetcher        fetch.Interface
	mountDirectory string
	baseURL        string

	templateIDs          []string
	templateIDsMu        sync.Mutex
	templateIDsError     error
	templateIDsErrorTime time.Time

	templates   map[string]queryResult
	templatesMu sync.Mutex
}

// NewTemplateStore returns a new *web.TemplateStore based on the configuration.
// If no stores are provided, this method returns nil.
func (c TemplatesConfig) NewTemplateStore() (*TemplateStore, error) {
	var fetcher fetch.Interface
	var mountDirectory, baseURL string
	switch {
	case c.Static != nil:
		fetcher = fetch.NewMemFetcher(c.Static)
	case c.Directory != "":
		fetcher = fetch.FromFilesystem(c.Directory)
		mountDirectory = path.Join(c.Directory, "static")
		baseURL = directoryBaseURL
	case c.URL != "":
		fetcher = fetch.FromHTTP(c.URL, true)
		baseURL = path.Join(c.URL, "static")
	default:
		return nil, nil
	}
	return &TemplateStore{
		fetcher:        fetcher,
		mountDirectory: mountDirectory,
		baseURL:        baseURL,
		templates:      make(map[string]queryResult),
	}, nil
}

// RegisterRoutes implements ttnweb.Registerer.
func (ts *TemplateStore) RegisterRoutes(server *ttnweb.Server) {
	if ts.mountDirectory != "" {
		server.Static(directoryBaseURL, http.Dir(ts.mountDirectory))
	}
}

// prependBaseURL prepends the base URL and the template ID to the LogoURL, if it is available.
func (ts *TemplateStore) prependBaseURL(template *ttnpb.ApplicationWebhookTemplate) {
	if template.LogoURL == "" {
		return
	}
	template.LogoURL = path.Join(ts.baseURL, template.TemplateID, template.LogoURL)
}

// GetTemplate returns the template with the given identifiers.
func (ts *TemplateStore) GetTemplate(ctx context.Context, req *ttnpb.GetApplicationWebhookTemplateRequest) (*ttnpb.ApplicationWebhookTemplate, error) {
	template, err := ts.getTemplate(req.ApplicationWebhookTemplateIdentifiers)
	if err != nil {
		return nil, err
	}
	template, err = applyWebhookTemplateFieldMask(nil, template, appendImplicitWebhookTemplatePaths(req.FieldMask.Paths...)...)
	if err != nil {
		return nil, err
	}
	ts.prependBaseURL(template)
	return template, nil
}

// ListTemplates returns the available templates.
func (ts *TemplateStore) ListTemplates(ctx context.Context, req *ttnpb.ListApplicationWebhookTemplatesRequest) (*ttnpb.ApplicationWebhookTemplates, error) {
	ids, err := ts.getAllTemplateIDs()
	if err != nil {
		return nil, err
	}

	var templates ttnpb.ApplicationWebhookTemplates
	for _, id := range ids {
		template, err := ts.getTemplate(ttnpb.ApplicationWebhookTemplateIdentifiers{
			TemplateID: id,
		})
		if err != nil {
			return nil, err
		}

		template, err = applyWebhookTemplateFieldMask(nil, template, appendImplicitWebhookTemplatePaths(req.FieldMask.Paths...)...)
		if err != nil {
			return nil, err
		}

		ts.prependBaseURL(template)

		templates.Templates = append(templates.Templates, template)
	}
	return &templates, nil
}

type queryResult struct {
	t    *ttnpb.ApplicationWebhookTemplate
	err  error
	time time.Time
}

var (
	errFetchFailed = errors.Define("fetch", "fetching failed")
	errParseFile   = errors.DefineCorruption("parse_file", "could not parse file")
)

func (ts *TemplateStore) allTemplateIDs() (ids []string, err error) {
	data, err := ts.fetcher.File("templates.yml")
	if err != nil {
		return nil, errFetchFailed.WithCause(err)
	}
	err = yaml.Unmarshal(data, &ids)
	if err != nil {
		return nil, errParseFile.WithCause(err)
	}
	return ids, nil
}

func (ts *TemplateStore) getAllTemplateIDs() ([]string, error) {
	ts.templateIDsMu.Lock()
	defer ts.templateIDsMu.Unlock()
	if ts.templateIDs != nil {
		return ts.templateIDs, nil
	}
	if time.Since(ts.templateIDsErrorTime) < yamlFetchErrorCache {
		return nil, ts.templateIDsError
	}
	ids, err := ts.allTemplateIDs()
	if err != nil {
		ts.templateIDsError, ts.templateIDsErrorTime = err, time.Now()
		return nil, err
	}
	ts.templateIDs, ts.templateIDsError, ts.templateIDsErrorTime = ids, nil, time.Time{}
	return ids, err
}

func (ts *TemplateStore) template(ids ttnpb.ApplicationWebhookTemplateIdentifiers) (*ttnpb.ApplicationWebhookTemplate, error) {
	data, err := ts.fetcher.File(fmt.Sprintf("%s.yml", ids.TemplateID))
	if err != nil {
		return nil, errFetchFailed.WithCause(err)
	}
	template := &ttnpb.ApplicationWebhookTemplate{}
	err = yaml.Unmarshal(data, template)
	if err != nil {
		return nil, errParseFile.WithCause(err)
	}
	return template, nil
}

func (ts *TemplateStore) getTemplate(ids ttnpb.ApplicationWebhookTemplateIdentifiers) (t *ttnpb.ApplicationWebhookTemplate, err error) {
	ts.templatesMu.Lock()
	defer ts.templatesMu.Unlock()
	if cached, ok := ts.templates[ids.TemplateID]; ok && cached.err == nil && time.Since(cached.time) < yamlFetchErrorCache {
		return cached.t, cached.err
	}
	template, err := ts.template(ids)
	ts.templates[ids.TemplateID] = queryResult{
		t:    template,
		err:  err,
		time: time.Now(),
	}
	return template, err
}

func appendImplicitWebhookTemplatePaths(paths ...string) []string {
	return append(append(make([]string, 0, 2+len(paths)),
		"ids",
		"name",
	), paths...)
}

func applyWebhookTemplateFieldMask(dst, src *ttnpb.ApplicationWebhookTemplate, paths ...string) (*ttnpb.ApplicationWebhookTemplate, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationWebhookTemplate{}
	}
	return dst, dst.SetFields(src, paths...)
}
