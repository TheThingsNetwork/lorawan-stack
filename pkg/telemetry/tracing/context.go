// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package tracing

import (
	"context"

	otrace "go.opentelemetry.io/otel/trace"
)

type tracerProviderKeyType struct{}

var tracerProviderKey = &tracerProviderKeyType{}

// NewContextWithTracerProvider returns a derived context with the tracer provider set.
func NewContextWithTracerProvider(ctx context.Context, tp otrace.TracerProvider) context.Context {
	return context.WithValue(ctx, tracerProviderKey, tp)
}

// FromContext returns the tracer provider that is attached to the context
// or returns a noop tracer provider if it does not exist.
func FromContext(ctx context.Context) otrace.TracerProvider {
	v := ctx.Value(tracerProviderKey)
	if v == nil {
		return otrace.NewNoopTracerProvider()
	}
	return v.(otrace.TracerProvider)
}
