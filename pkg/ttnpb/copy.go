// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import "google.golang.org/protobuf/proto"

// Clone creates a deep copy of the given message.
func Clone[X proto.Message](in X) X {
	return proto.Clone(in).(X)
}

// CloneSlice creates a deep copy of the given slice of messages.
func CloneSlice[X proto.Message](in []X) []X {
	if in == nil {
		return nil
	}
	out := make([]X, len(in))
	for i, x := range in {
		out[i] = Clone(x)
	}
	return out
}
