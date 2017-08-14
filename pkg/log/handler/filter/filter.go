// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package filter

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/log"
)

// Filtered is a log.Handler that only logs fields that match the
// specified filters.
type Filtered struct {
	log.Handler
	filter Filter
}

// Filter something that can filter a log entry.
type Filter interface {
	Filter(log.Entry) bool
}

// FilterFunc is a wrapper type for simple functions that filter.
type FilterFunc func(log.Entry) bool

// Filter implements Filter.
func (fn FilterFunc) Filter(e log.Entry) bool {
	return fn(e)
}

// Wrap returns a new filtered logger that allows all log entries to be passed.
func Wrap(handler log.Handler, filters ...Filter) *Filtered {
	return &Filtered{
		Handler: handler,
		filter:  And(filters...),
	}
}

// SetFilters sets the filters.
func (f *Filtered) SetFilters(filters ...Filter) {
	f.filter = And(filters...)
}

// HandleLog implements log.Handler.
func (f *Filtered) HandleLog(e log.Entry) error {
	if f.filter.Filter(e) {
		return f.Handler.HandleLog(e)
	}
	return nil
}

// All is a filter allows all log entries to be passed.
var All = FilterFunc(func(e log.Entry) bool {
	return true
})

// And is a combinator that combines filters in such a way that the log entry is only
// passed is all filters pass it.
func And(filters ...Filter) Filter {
	return FilterFunc(func(e log.Entry) bool {
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
	return FilterFunc(func(e log.Entry) bool {
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
	return FilterFunc(func(e log.Entry) bool {
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
