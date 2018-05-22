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

package errors

import (
	"context"
	"io"
	"os"
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

// ErrOSNotExist is the descriptor for the os.ErrNotExist error.
var ErrOSNotExist = &ErrDescriptor{
	MessageFormat: os.ErrNotExist.Error(),
	Code:          Code(1),
	Namespace:     "os",
	Type:          NotFound,
}

func init() {
	ErrContextCanceled.Register()
	ErrContextDeadlineExceeded.Register()
	ErrEOF.Register()
	ErrOSNotExist.Register()
}

// From lifts an error to be an Error.
func From(in error) Error {
	if in == nil {
		return nil
	}
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
	case os.ErrNotExist:
		return ErrOSNotExist.New(nil)
	}

	return normalize(&Impl{
		info: info{
			Message: in.Error(),
			Code:    Code(0),
			Type:    Unknown,
		},
	})
}
