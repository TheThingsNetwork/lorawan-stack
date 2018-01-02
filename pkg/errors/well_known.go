// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"context"
	"io"
)

// ErrEOF is the descriptor for the io.EOF error.
var ErrEOF = &ErrDescriptor{
	MessageFormat: io.EOF.Error(),
	Code:          Code(1),
	Namespace:     "io",
	Type:          OutOfRange,
}

// ErrContextCanceled is the descriptor for the context.Canceled error
var ErrContextCanceled = &ErrDescriptor{
	MessageFormat: context.Canceled.Error(),
	Code:          Code(1),
	Namespace:     "context",
	Type:          Canceled,
}

// ErrContextDeadlineExceeded is the descriptor for the context.DeadlineExceeded error
var ErrContextDeadlineExceeded = &ErrDescriptor{
	MessageFormat: context.DeadlineExceeded.Error(),
	Code:          Code(2),
	Namespace:     "context",
	Type:          Timeout,
}

func init() {
	ErrContextCanceled.Register()
	ErrContextDeadlineExceeded.Register()
	ErrEOF.Register()
}

// From lifts an error to be and Error.
func From(in error) Error {
	if err, ok := in.(Error); ok {
		return err
	}

	switch in {
	case io.EOF:
		return ErrEOF.New(nil)
	case context.Canceled:
		return ErrContextCanceled.New(nil)
	case context.DeadlineExceeded:
		return ErrContextDeadlineExceeded.New(nil)
	}

	return normalize(&Impl{
		info: info{
			Message: in.Error(),
			Code:    Code(0),
			Type:    Unknown,
		},
	})
}
