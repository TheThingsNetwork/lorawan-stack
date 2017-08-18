// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package udp

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
		return err
	}
	*et = ExpandedTime(t)
	return nil
}
