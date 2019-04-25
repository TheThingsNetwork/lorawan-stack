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

package jsonpb

import "github.com/grpc-ecosystem/grpc-gateway/runtime"

// TTN returns the default TTN JSONPb marshaler.
func TTN() *GoGoJSONPb {
	return &GoGoJSONPb{
		OrigName:    true,
		EnumsAsInts: true,
	}
}

// TTNEventStream returns a TTN JsonPb marshaler with double newlines for
// text/event-stream compatibility.
func TTNEventStream() runtime.Marshaler {
	return &ttnEventStream{GoGoJSONPb: TTN()}
}

type ttnEventStream struct {
	*GoGoJSONPb
}

func (s *ttnEventStream) ContentType() string { return "text/event-stream" }
func (s *ttnEventStream) Delimiter() []byte   { return []byte{'\n', '\n'} }
