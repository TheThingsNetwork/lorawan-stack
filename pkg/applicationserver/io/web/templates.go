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
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gopkg.in/yaml.v2"
)

const yamlFetchErrorCache = 1 * time.Minute

// TemplateStore contains the webhook templates.
type TemplateStore struct {
	Fetcher fetch.Interface

	templateIDs          []string
	templateIDsMu        sync.Mutex
	templateIDsError     error
	templateIDsErrorTime time.Time

	templates   map[string]queryResult
	templatesMu sync.Mutex
}

// NewTemplateStore creates a new template store that is backed by the provided fetcher.
func NewTemplateStore(fetcher fetch.Interface) (*TemplateStore, error) {
	return &TemplateStore{
		Fetcher:   fetcher,
		templates: make(map[string]queryResult),
	}, nil
}

// GetTemplate returns the template with the given identifiers.
func (ts *TemplateStore) GetTemplate(ctx context.Context, req *ttnpb.GetApplicationWebhookTemplateRequest) (*ttnpb.ApplicationWebhookTemplate, error) {
	template, err := ts.getTemplate(req.ApplicationWebhookTemplateIdentifiers)
	if err != nil {
		return nil, err
	}
	return applyWebhookTemplateFieldMask(nil, template, appendImplicitWebhookTemplatePaths(req.FieldMask.Paths...)...)
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
	data, err := ts.Fetcher.File("templates.yml")
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
	data, err := ts.Fetcher.File(fmt.Sprintf("%s.yml", ids.TemplateID))
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
