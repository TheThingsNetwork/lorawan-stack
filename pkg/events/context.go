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

package events

import "context"

// ContextMarshaler interface for marshaling/unmarshaling contextual information
// to/from events.
type ContextMarshaler interface {
	MarshalContext(context.Context) []byte
	UnmarshalContext(context.Context, []byte) (context.Context, error)
}

var contextMarshalers = map[string]ContextMarshaler{}

// RegisterContextMarshaler registers a ContextMarshaler with the given name.
// This should only be called from init funcs.
func RegisterContextMarshaler(name string, m ContextMarshaler) {
	contextMarshalers[name] = m
}

func unmarshalContext(ctx context.Context, data map[string][]byte) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	for name, payload := range data {
		m, ok := contextMarshalers[name]
		if !ok {
			continue
		}
		ctx, err = m.UnmarshalContext(ctx, payload)
		if err != nil {
			return nil, err
		}
		delete(data, name)
	}
	return ctx, nil
}

func marshalContext(ctx context.Context) (map[string][]byte, error) {
	data := make(map[string][]byte, len(contextMarshalers))
	for name, m := range contextMarshalers {
		payload := m.MarshalContext(ctx)
		if payload == nil {
			continue
		}
		data[name] = payload
	}
	if len(data) == 0 {
		return nil, nil
	}
	return data, nil
}
