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

package sentry

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// NewEvent creates a new Sentry event for the given error.
func NewEvent(err error) *sentry.Event {
	evt := sentry.NewEvent()
	if err == nil {
		return evt
	}

	evt.Message = err.Error()

	// Error Tags.
	if ttnErr, ok := errors.From(err); ok && ttnErr != nil {
		evt.Tags["error.namespace"] = ttnErr.Namespace()
		evt.Tags["error.name"] = ttnErr.Name()
		if correlationID := ttnErr.CorrelationID(); correlationID != "" {
			evt.EventID = sentry.EventID(correlationID)
		}
	}

	errStack := errors.Stack(err)

	// Error Attributes.
	for k, v := range errors.Attributes(errStack...) {
		if val := fmt.Sprint(v); len(val) < 64 {
			evt.Extra["error.attributes."+k] = val
		}
	}

	// Error Stack.
	for _, err := range errStack {
		exception := sentry.Exception{
			Value: err.Error(),
		}
		if ttnErr, ok := errors.From(err); ok && ttnErr != nil {
			exception.Type = ttnErr.Name()
			exception.Module = ttnErr.Namespace()
			exception.Value = ttnErr.FormatMessage(ttnErr.PublicAttributes())
		}
		if stackTrace := sentry.ExtractStacktrace(err); stackTrace != nil {
			exception.Stacktrace = stackTrace
		}
		evt.Exception = append(evt.Exception, exception)
	}

	return evt
}
