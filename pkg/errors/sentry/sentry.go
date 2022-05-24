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

// Package sentry converts errors to Sentry events.
package sentry

import (
	"strings"

	"github.com/getsentry/sentry-go"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

// NewEvent creates a new Sentry event for the given error.
func NewEvent(err error) *sentry.Event {
	evt := sentry.NewEvent()
	if err == nil {
		return evt
	}

	evt.Level = sentry.LevelError

	errStack := errors.Stack(err)

	messages := make([]string, 0, len(errStack))

	for i, err := range errStack {
		messages = append(messages, err.Error())
		exception := sentry.Exception{Value: err.Error()}
		if ttnErr, ok := errors.From(err); ok && ttnErr != nil {
			exception.Module = ttnErr.Namespace()
			exception.Type = ttnErr.Name()
			if i == 0 { // We set the namespace, name and ID from the first error in the chain.
				evt.Tags["error.namespace"] = ttnErr.Namespace()
				evt.Tags["error.name"] = ttnErr.Name()
				if correlationID := ttnErr.CorrelationID(); correlationID != "" {
					evt.EventID = sentry.EventID(correlationID)
				}
			}
			evt.Contexts[ttnErr.FullName()+" attributes"] = ttnErr.Attributes()
			if stackTrace := sentry.ExtractStacktrace(err); stackTrace != nil {
				exception.Stacktrace = stackTrace
			}
		}
		evt.Exception = append(evt.Exception, exception)
	}

	evt.Message = strings.Join(messages, "\n--- ")

	return evt
}
