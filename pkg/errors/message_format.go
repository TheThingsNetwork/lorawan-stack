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

package errors

import "github.com/gotnospirit/messageformat"

var formatter, _ = messageformat.New()

// FormatMessage formats the message using the given attributes.
func (d Definition) FormatMessage(attributes map[string]interface{}) string {
	if len(attributes) == 0 {
		return d.messageFormat
	}
	parsedMessageFormat := d.parsedMessageFormat
	if parsedMessageFormat == nil {
		parsedMessageFormat, _ = formatter.Parse(d.messageFormat)
	}
	if parsedMessageFormat != nil {
		formatted, err := parsedMessageFormat.FormatMap(attributes)
		if err == nil {
			return formatted
		}
	}
	return d.messageFormat
}
