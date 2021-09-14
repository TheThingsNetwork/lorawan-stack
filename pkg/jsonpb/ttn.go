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

import (
	"encoding/json"
	"io"

	"github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// TTN returns the default TTN JSONPb marshaler.
func TTN() *TTNMarshaler {
	return &TTNMarshaler{
		GoGoJSONPb: &GoGoJSONPb{
			OrigName:    true,
			EnumsAsInts: true,
		},
	}
}

type TTNMarshaler struct {
	*GoGoJSONPb
}

func (*TTNMarshaler) ContentType() string { return "application/json" }

func (m *TTNMarshaler) Marshal(v interface{}) ([]byte, error) {
	if marshaler, ok := v.(jsonplugin.Marshaler); ok {
		return jsonplugin.MarshalerConfig{
			EnumsAsInts: true,
		}.Marshal(marshaler)
	}
	return m.GoGoJSONPb.Marshal(v)
}

func (m *TTNMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return &TTNEncoder{w: w, gogo: m.GoGoJSONPb}
}

type TTNEncoder struct {
	w    io.Writer
	gogo *GoGoJSONPb
}

func (e *TTNEncoder) Encode(v interface{}) error {
	if marshaler, ok := v.(jsonplugin.Marshaler); ok {
		b, err := jsonplugin.MarshalerConfig{
			EnumsAsInts: true,
		}.Marshal(marshaler)
		if err != nil {
			return err
		}
		_, err = e.w.Write(b)
		return err
	}
	return e.gogo.NewEncoder(e.w).Encode(v)
}

func (m *TTNMarshaler) Unmarshal(data []byte, v interface{}) error {
	if unmarshaler, ok := v.(jsonplugin.Unmarshaler); ok {
		return jsonplugin.UnmarshalerConfig{}.Unmarshal(data, unmarshaler)
	}
	return m.GoGoJSONPb.Unmarshal(data, v)
}

func (m *TTNMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return &NewDecoder{r: r, gogo: m.GoGoJSONPb}
}

type NewDecoder struct {
	r    io.Reader
	gogo *GoGoJSONPb
}

func (d *NewDecoder) Decode(v interface{}) error {
	if unmarshaler, ok := v.(jsonplugin.Unmarshaler); ok {
		var data json.RawMessage
		err := json.NewDecoder(d.r).Decode(&data)
		if err != nil {
			return err
		}
		return jsonplugin.UnmarshalerConfig{}.Unmarshal(data, unmarshaler)
	}
	return d.gogo.NewDecoder(d.r).Decode(v)
}

// TTNEventStream returns a TTN JsonPb marshaler with double newlines for
// text/event-stream compatibility.
func TTNEventStream() runtime.Marshaler {
	return &ttnEventStream{TTNMarshaler: TTN()}
}

type ttnEventStream struct {
	*TTNMarshaler
}

func (s *ttnEventStream) ContentType() string { return "text/event-stream" }

func (s *ttnEventStream) Delimiter() []byte { return []byte{'\n', '\n'} }
