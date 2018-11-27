// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package blacklist

import "context"

type ctxKeyType struct{}

var ctxKey ctxKeyType

// FromContext returns the blacklists from the given context.
func FromContext(ctx context.Context) Blacklists {
	if blacklists, ok := ctx.Value(ctxKey).(Blacklists); ok {
		return blacklists
	}
	return nil
}

// NewContext returns a new context derived from parent with the given blacklists attached.
func NewContext(parent context.Context, blacklists ...*Blacklist) context.Context {
	if len(blacklists) == 0 {
		return parent
	}
	parentBlacklists := FromContext(parent)
	newBlacklists := make(Blacklists, 0, len(parentBlacklists)+len(blacklists))
	newBlacklists = append(newBlacklists, parentBlacklists...)
	newBlacklists = append(newBlacklists, blacklists...)
	return context.WithValue(parent, ctxKey, newBlacklists)
}
