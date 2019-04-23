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

import (
	"fmt"
	"io"
	"os"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

// PrintStack prints the error stack to w.
func PrintStack(w io.Writer, err error) {
	for i, err := range errors.Stack(err) {
		if i == 0 {
			fmt.Fprintln(w, err)
		} else {
			fmt.Fprintf(w, "--- %s\n", err)
		}
		for k, v := range errors.Attributes(err) {
			fmt.Fprintf(os.Stderr, "    %s=%v\n", k, v)
		}
		if ttnErr, ok := errors.From(err); ok {
			if correlationID := ttnErr.CorrelationID(); correlationID != "" {
				fmt.Fprintf(os.Stderr, "    correlation_id=%s\n", ttnErr.CorrelationID())
			}
		}
	}
}
