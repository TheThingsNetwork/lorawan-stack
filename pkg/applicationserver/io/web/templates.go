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

	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gopkg.in/yaml.v2"
)

// TemplateStore contains the webhook templates.
type TemplateStore struct {
	Fetcher fetch.Interface

	templates      sync.Map
	templateIDs    *[]string
	templateIDsMu  sync.Mutex
	templateIDsErr error
}

// NewTemplateStore creates a new template store that is backed by the provided fetcher.
func NewTemplateStore(fetcher fetch.Interface) (*TemplateStore, error) {
	return &TemplateStore{
		Fetcher: fetcher,
	}, nil
}

// Get returns the template with the given identifiers.
func (ts *TemplateStore) Get(ctx context.Context, req *ttnpb.GetApplicationWebhookTemplateRequest) (*ttnpb.ApplicationWebhookTemplate, error) {
	template, err := ts.getTemplate(req.ApplicationWebhookTemplateIdentifiers)
	if err != nil {
		return nil, err
	}
	return applyWebhookTemplateFieldMask(nil, template, appendImplicitWebhookTemplatePaths(req.FieldMask.Paths...)...)
}

// List returns the available templates.
func (ts *TemplateStore) List(ctx context.Context, req *ttnpb.ListApplicationWebhookTemplatesRequest) (*ttnpb.ApplicationWebhookTemplates, error) {
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

type registeredTemplate struct {
	t     *ttnpb.ApplicationWebhookTemplate
	err   error
	ready chan struct{}
}

func (ts *TemplateStore) getAllTemplateIDs() (ids []string, err error) {
	ts.templateIDsMu.Lock()
	defer ts.templateIDsMu.Unlock()
	if ts.templateIDs != nil {
		return *ts.templateIDs, ts.templateIDsErr
	}
	defer func() {
		ts.templateIDs, ts.templateIDsErr = &ids, err
	}()

	data, err := ts.Fetcher.File("templates.yml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &ids)
	if err != nil {
		return nil, err
	}
	return ids, err
}

func (ts *TemplateStore) getTemplate(ids ttnpb.ApplicationWebhookTemplateIdentifiers) (t *ttnpb.ApplicationWebhookTemplate, err error) {
	registeredI, ok := ts.templates.LoadOrStore(ids.TemplateID, &registeredTemplate{ready: make(chan struct{})})
	registered := registeredI.(*registeredTemplate)
	if ok {
		<-registered.ready
		return registered.t, registered.err
	}
	defer func() {
		registered.t, registered.err = t, err
		close(registered.ready)
	}()

	data, err := ts.Fetcher.File(fmt.Sprintf("%s.yml", ids.TemplateID))
	if err != nil {
		return nil, err
	}

	template := &ttnpb.ApplicationWebhookTemplate{}
	err = yaml.Unmarshal(data, template)
	if err != nil {
		return nil, err
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
