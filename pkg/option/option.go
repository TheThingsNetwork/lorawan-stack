// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package option

// Option is the interface of all functional options
type Option interface {
	// Apply applies the option to the structure if possible, or does nothing when
	// not applicable
	Apply(interface{}) error
}

// Fn is a function that can apply itself as an option
type Fn func(interface{}) error

// Apply implements Option
func (opt Fn) Apply(to interface{}) error {
	return opt(to)
}

// Options is a list of options that get applied in order
type Options []Option

// Apply implements Option
func (opts Options) Apply(to interface{}) error {
	for _, opt := range opts {
		err := opt.Apply(to)
		if err != nil {
			return err
		}
	}

	return nil
}
