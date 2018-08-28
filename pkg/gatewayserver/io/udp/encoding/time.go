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

package encoding

import "time"

// CompactTime is a time.Time that is encoded as ISO 8601 'compact' format
type CompactTime time.Time

// CompactFormat formats a time as ISO 8601 'compact' with µs precision
const CompactFormat = "2006-01-02T15:04:05.999999Z07:00"

// MarshalJSON implements the json.Marshaler interface.
func (ct CompactTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(ct).UTC().Format(`"` + CompactFormat + `"`)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ct *CompactTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"`+CompactFormat+`"`, string(data))
	if err != nil {
		return err
	}
	*ct = CompactTime(t)
	return nil
}

// ExpandedTime is a time.Time that is encoded as ISO 8601 'expanded' format
type ExpandedTime time.Time

// ExpandedFormat formats a time as ISO 8601 'expanded'
const ExpandedFormat = "2006-01-02 15:04:05 MST"

// MarshalJSON implements the json.Marshaler interface.
func (et ExpandedTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(et).UTC().Format(`"` + ExpandedFormat + `"`)), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (et *ExpandedTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(`"`+ExpandedFormat+`"`, string(data))
	if err != nil {
		return errTimestamp.WithCause(err)
	}
	*et = ExpandedTime(t)
	return nil
}
