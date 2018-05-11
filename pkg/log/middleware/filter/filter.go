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

// Package filter implements a pkg/log.Handler that only logs fields that match the specified filters
package filter

import (
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/log"
)

// Filtered is a log.Handler that only logs fields that match the
// specified filters.
type Filtered struct {
	filter Filter
}

// Filter something that can filter a log entry.
type Filter interface {
	Filter(log.Entry) bool
}

// Func is a wrapper type for simple functions that filter.
type Func func(log.Entry) bool

// Filter implements Filter.
func (fn Func) Filter(e log.Entry) bool {
	return fn(e)
}

// SetFilters sets the filters.
func (f *Filtered) SetFilters(filters ...Filter) {
	f.filter = And(filters...)
}

// Wrap  implements log.Middleware.
func (f *Filtered) Wrap(next log.Handler) log.Handler {
	return log.HandlerFunc(func(entry log.Entry) error {
		if f.filter.Filter(entry) {
			return next.HandleLog(entry)
		}

		return nil
	})
}

// All is a filter allows all log entries to be passed.
var All = Func(func(e log.Entry) bool {
	return true
})

// And is a combinator that combines filters in such a way that the log entry is only
// passed is all filters pass it.
func And(filters ...Filter) Filter {
	return Func(func(e log.Entry) bool {
		for _, filter := range filters {
			if !filter.Filter(e) {
				return false
			}
		}

		return true
	})
}

// Or is a combinator for filters that passes a log entry only if one of the filters passes it.
func Or(filters ...Filter) Filter {
	return Func(func(e log.Entry) bool {
		for _, filter := range filters {
			if filter.Filter(e) {
				return true
			}
		}

		return false
	})
}

// Field returns a filter that passes a log entry if and only if the supplied field name is present and the matcher
// returns true when called on the field value.
func Field(field string, matcher func(interface{}) bool) Filter {
	return Func(func(e log.Entry) bool {
		fields := e.Fields().Fields()
		val, ok := fields[field]
		return ok && matcher(val)
	})
}

// FieldString returns a filter that passes only if the string representation of the field value for the field equals the passed string.
func FieldString(field string, value string) Filter {
	return Field(field, func(val interface{}) bool {
		if stringer, ok := val.(fmt.Stringer); ok {
			return stringer.String() == value
		}

		return fmt.Sprintf("%v", val) == value
	})
}
