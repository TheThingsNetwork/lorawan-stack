// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package webhandlers

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// Data contains data to render templates.
type Data struct {
	ErrorTitle          string
	ErrorMessage        string
	ErrorID             string
	CorrelationID       string
	BackendErrorDetails string
	Year                int
	IsGenericNotFound   bool
}

//go:embed "error_template.html.tmpl"
var errorTemplate string

// Template for rendering the static error.
var Template = func() *ErrorTemplate {
	return NewErrorTemplate(template.Must(template.New("error").Parse(errorTemplate)))
}()

// ErrorTemplate wraps the error template for the static error route.
type ErrorTemplate struct {
	template *template.Template
}

// NewErrorTemplate instantiates a new error template for non-frontend handled routes.
func NewErrorTemplate(t *template.Template) *ErrorTemplate {
	return &ErrorTemplate{template: t}
}

// ServeHTTP renders the non-frontend handled errors.
func (t *ErrorTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := RetrieveError(r)
	if err == nil {
		return
	}
	errMsg, _ := json.MarshalIndent(err, "", " ")
	errorID := "n/a"
	errorCorrelationID := "n/a"
	var errorTitle string
	if ttnErr, ok := errors.From(err); ok {
		errorID = ttnErr.FullName()
		errorCorrelationID = ttnErr.CorrelationID()
		errorTitle = ttnErr.FormatMessage(ttnErr.Attributes())
	}
	var errorMessage string
	isGenericNotFound := errors.Resemble(err, errRouteNotFound)
	switch errors.Code(err) {
	case http.StatusNotFound:
		if isGenericNotFound {
			errorTitle = "Page not found"
		}
		errorMessage = "The resource you requested cannot be found."
	case http.StatusUnauthorized:
		errorMessage = "You are not allowed to perform this action."
	default:
		errorMessage = "An unknown error occurred."
	}
	if err := t.template.Execute(w, Data{
		ErrorTitle:          errorTitle,
		ErrorMessage:        errorMessage,
		ErrorID:             errorID,
		CorrelationID:       errorCorrelationID,
		BackendErrorDetails: string(errMsg),
		Year:                time.Now().Year(),
		IsGenericNotFound:   isGenericNotFound,
	}); err != nil {
		log.FromContext(r.Context()).WithError(err).Warn("Failed to execute template")
	}
}
