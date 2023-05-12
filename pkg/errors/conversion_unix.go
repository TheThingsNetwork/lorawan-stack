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

//go:build linux || darwin
// +build linux darwin

package errors

import (
	"errors"
	"syscall"
)

func syscallErrorAttributes(err error) []any {
	if matched := (syscall.Errno)(0); errors.As(err, &matched) {
		// syscall.Errono do not contain any sensitive information and are safe to render.
		return []any{
			"error", matched.Error(),
			"timeout", matched.Timeout(),
		}
	}
	return nil
}
