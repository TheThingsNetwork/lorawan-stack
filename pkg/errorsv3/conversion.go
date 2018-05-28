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

import "google.golang.org/grpc/status"

// From returns an *Error if it can be derived from the given input.
func From(err error) (out *Error, ok bool) {
	if err == nil {
		return nil, true
	}
	defer func() {
		if out != nil {
			var copy = *out
			out = &copy
		}
	}()
	switch err := err.(type) {
	case Error:
		return &err, true
	case *Error:
		if err == nil {
			return nil, true
		}
		return err, true
	default:
		if se, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
			err := FromGRPCStatus(se.GRPCStatus())
			return &err, true
		}
	}
	return nil, false
}
