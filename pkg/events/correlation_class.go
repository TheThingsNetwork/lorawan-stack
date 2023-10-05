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

package events

import (
	"context"
	"fmt"
	"strings"
)

type (
	correlationIDClass  string
	correlationIDPrefix string
)

var (
	correlationIDPrefixesByClass = make(map[correlationIDClass]map[correlationIDPrefix]struct{})
	correlationIDClassByPrefix   = make(map[correlationIDPrefix]correlationIDClass)
)

// RegisterCorrelationIDPrefix register a prefix which ensures that will be unique within the same class.
// The function returns a function that can be used to add the correlation ID to a context.
// If the prefix is already registered, the function panics.
// The function is not goroutine safe, and it is meant to be called during package initialization.
func RegisterCorrelationIDPrefix(
	class correlationIDClass, prefix correlationIDPrefix,
) func(context.Context, ...string) context.Context {
	byClass, ok := correlationIDPrefixesByClass[class]
	if !ok {
		byClass = make(map[correlationIDPrefix]struct{})
		correlationIDPrefixesByClass[class] = byClass
	}
	if _, ok := byClass[prefix]; ok {
		panic(fmt.Sprintf("prefix `%v` already registered in class `%v`", prefix, class))
	}
	if _, ok := correlationIDClassByPrefix[prefix]; ok {
		panic(fmt.Sprintf("prefix `%v` already registered in class `%v`", prefix, class))
	}
	byClass[prefix] = struct{}{}
	correlationIDClassByPrefix[prefix] = class
	return func(ctx context.Context, suffixes ...string) context.Context {
		if correlationIDClassPresent(ctx, byClass) {
			return ctx
		}
		suffix := strings.Join(suffixes, ":")
		if suffix == "" {
			suffix = NewCorrelationID()
		}
		return ContextWithCorrelationID(ctx, fmt.Sprintf("%v:%v", prefix, suffix))
	}
}

func correlationIDClassPresent(ctx context.Context, prefixes map[correlationIDPrefix]struct{}) bool {
	correlationIDs := CorrelationIDsFromContext(ctx)
	if len(correlationIDs) == 0 {
		return false
	}
	for prefix := range prefixes {
		prefix := string(prefix) + ":"
		for _, correlationID := range correlationIDs {
			if strings.HasPrefix(correlationID, prefix) {
				return true
			}
		}
	}
	return false
}
