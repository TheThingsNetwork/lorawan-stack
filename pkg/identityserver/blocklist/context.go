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

package blocklist

import "context"

type ctxKeyType struct{}

var ctxKey ctxKeyType

// FromContext returns the blocklists from the given context.
func FromContext(ctx context.Context) Blocklists {
	if blocklists, ok := ctx.Value(ctxKey).(Blocklists); ok {
		return blocklists
	}
	return nil
}

// NewContext returns a new context derived from parent with the given blocklists attached.
func NewContext(parent context.Context, blocklists ...*Blocklist) context.Context {
	if len(blocklists) == 0 {
		return parent
	}
	parentBlocklists := FromContext(parent)
	newBlocklists := make(Blocklists, 0, len(parentBlocklists)+len(blocklists))
	newBlocklists = append(newBlocklists, parentBlocklists...)
	newBlocklists = append(newBlocklists, blocklists...)
	return context.WithValue(parent, ctxKey, newBlocklists)
}
