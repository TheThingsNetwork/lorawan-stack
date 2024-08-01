// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package mux

import "go.thethings.network/lorawan-stack/v3/pkg/events"

// Matcher matches event names.
type Matcher interface {
	Matches(name string) bool
}

// MatcherFunc is a function that implements the Matcher interface.
type MatcherFunc func(name string) bool

// Matches implements Matcher.
func (f MatcherFunc) Matches(name string) bool {
	return f(name)
}

var (
	// MatchAll is a matcher that matches all event names.
	MatchAll Matcher = MatcherFunc(func(string) bool {
		return true
	})
	// MatchNone is a matcher that matches no event names.
	MatchNone Matcher = MatcherFunc(func(string) bool {
		return false
	})
)

// MatchNames is a matcher that matches specific event names.
func MatchNames(names ...string) Matcher {
	return MatcherFunc(func(name string) bool {
		for _, n := range names {
			if n == name {
				return true
			}
		}
		return false
	})
}

// MatchPatterns is a matcher that matches event names based on patterns.
func MatchPatterns(patterns ...string) (Matcher, error) {
	definedNames := make(map[string]struct{})
	for _, def := range events.All().Definitions() {
		definedNames[def.Name()] = struct{}{}
	}
	names, err := events.NamesFromPatterns(definedNames, patterns)
	if err != nil {
		return nil, err
	}
	return MatchNames(names...), nil
}
