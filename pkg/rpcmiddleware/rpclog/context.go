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

package rpclog

import "context"

type requestFieldsKeyType struct{}

var requestFieldsKey requestFieldsKeyType

type requestFieldsValue struct {
	fields map[string]any
}

func requestFieldsFromContext(ctx context.Context) (*requestFieldsValue, bool) {
	value, ok := ctx.Value(requestFieldsKey).(*requestFieldsValue)
	return value, ok
}

func newContextWithRequestFields(parent context.Context) context.Context {
	return context.WithValue(parent, requestFieldsKey, &requestFieldsValue{
		fields: make(map[string]any),
	})
}

// AddField adds a log field to the fields in the request context.
// Not safe for concurrent use.
func AddField(ctx context.Context, key string, value any) {
	if v, ok := requestFieldsFromContext(ctx); ok {
		v.fields[key] = value
	}
}
