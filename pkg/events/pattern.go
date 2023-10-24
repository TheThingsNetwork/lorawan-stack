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
	"regexp"
	"sort"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	errInvalidRegexp    = errors.DefineInvalidArgument("invalid_regexp", "invalid regexp")
	errNoMatchingEvents = errors.DefineInvalidArgument("no_matching_events", "no matching events for regexp `{regexp}`")
	errUnknownEventName = errors.DefineInvalidArgument("unknown_event_name", "unknown event `{name}`")
)

// NamesFromPatterns returns the event names which match the given patterns.
// The defined names are a set of event names which are used to match the patterns.
func NamesFromPatterns(definedNames map[string]struct{}, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	nameMap := make(map[string]struct{})
	for _, name := range patterns {
		if strings.HasPrefix(name, "/") && strings.HasSuffix(name, "/") {
			re, err := regexp.Compile(strings.Trim(name, "/"))
			if err != nil {
				return nil, errInvalidRegexp.WithCause(err)
			}
			var found bool
			for defined := range definedNames {
				if re.MatchString(defined) {
					nameMap[defined] = struct{}{}
					found = true
				}
			}
			if !found {
				return nil, errNoMatchingEvents.WithAttributes("regexp", re.String())
			}
		} else {
			var found bool
			for defined := range definedNames {
				if name == defined {
					nameMap[name] = struct{}{}
					found = true
					break
				}
			}
			if !found {
				return nil, errUnknownEventName.WithAttributes("name", name)
			}
		}
	}
	if len(nameMap) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(nameMap))
	for name := range nameMap {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}
