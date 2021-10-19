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

package i18n

import (
	"regexp"
	"sort"
)

var messageFormatArgument = regexp.MustCompile(`\{[\s]*([a-z0-9_]+)`)

func messageFormatArguments(messageFormat string) (args []string) {
	for _, matches := range messageFormatArgument.FindAllStringSubmatch(messageFormat, -1) {
		if len(matches) == 2 {
			args = append(args, matches[1])
		}
	}
	m := make(map[string]struct{}, len(args))
	for _, arg := range args {
		m[arg] = struct{}{}
	}
	args = make([]string, 0, len(m))
	for arg := range m {
		args = append(args, arg)
	}
	sort.Strings(args)
	return
}
