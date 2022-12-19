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
	"bytes"
	"encoding/json"
	"io"
	"sort"

	"github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
)

// TTN returns the default JSONPb marshaler of The Things Stack.
func TTN() *TTNMarshaler {
	return &TTNMarshaler{
		GoGoJSONPb: &GoGoJSONPb{
			OrigName:    true,
			EnumsAsInts: true,
		},
	}
}

// TTNMarshaler is the JSON marshaler/unmarshaler that is used in grpc-gateway.
type TTNMarshaler struct {
	*GoGoJSONPb
}

// ContentType returns the content-type of the marshaler.
func (*TTNMarshaler) ContentType() string { return "application/json" }

// Marshal marshals v to JSON.
func (m *TTNMarshaler) Marshal(v any) ([]byte, error) {
	b, err := marshalAny(v, m.GoGoJSONPb)
	if err != nil {
		return nil, err
	}
	if m.GoGoJSONPb.Indent == "" {
		return b, nil
	}
	var buf bytes.Buffer
	if err = json.Indent(&buf, b, "", m.GoGoJSONPb.Indent); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshalAny(v any, fallback runtime.Marshaler) ([]byte, error) {
	if marshaler, ok := v.(jsonplugin.Marshaler); ok {
		b, err := jsonplugin.MarshalerConfig{
			EnumsAsInts: true,
		}.Marshal(marshaler)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	if kv, ok := v.(map[string]any); ok {
		return marshalMap(kv, fallback)
	}
	if kv, ok := v.(map[string]proto.Message); ok {
		return marshalMap(kv, fallback)
	}
	return fallback.Marshal(v)
}

func marshalMap[X any](kv map[string]X, fallback runtime.Marshaler) ([]byte, error) {
	keys := maps.Keys(kv)
	sort.Strings(keys)
	pluginMarshaler := jsonplugin.NewMarshalState(jsonplugin.MarshalerConfig{
		EnumsAsInts: true,
	})
	pluginMarshaler.WriteObjectStart()
	for _, k := range keys {
		pluginMarshaler.WriteObjectField(k)
		b, err := marshalAny(kv[k], fallback)
		if err != nil {
			return nil, err
		}
		_, err = pluginMarshaler.Write(b)
		if err != nil {
			return nil, err
		}
	}
	pluginMarshaler.WriteObjectEnd()
	return pluginMarshaler.Bytes()
}

// NewEncoder returns a new JSON encoder that writes values to w.
func (m *TTNMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return &TTNEncoder{w: w, gogo: m.GoGoJSONPb}
}

// TTNEncoder marshals values to JSON and writes them to an io.Writer.
type TTNEncoder struct {
	w    io.Writer
	gogo *GoGoJSONPb
}

// Encode marshals v to JSON and writes it to the writer.
func (e *TTNEncoder) Encode(v any) error {
	b, err := marshalAny(v, e.gogo)
	if err != nil {
		return err
	}
	if e.gogo.Indent == "" {
		_, err = e.w.Write(b)
		return err
	}
	var buf bytes.Buffer
	if err = json.Indent(&buf, b, "", e.gogo.Indent); err != nil {
		return err
	}
	_, err = io.Copy(e.w, &buf)
	return err
}

// Unmarshal unmarshals v from JSON data.
func (m *TTNMarshaler) Unmarshal(data []byte, v any) error {
	if unmarshaler, ok := v.(jsonplugin.Unmarshaler); ok {
		return jsonplugin.UnmarshalerConfig{}.Unmarshal(data, unmarshaler)
	}
	return m.GoGoJSONPb.Unmarshal(data, v)
}

// NewDecoder returns a new JSON decoder that reads data from r.
func (m *TTNMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return &TTNDecoder{d: json.NewDecoder(r), gogo: m.GoGoJSONPb}
}

// TTNDecoder reads JSON data from an io.Reader and unmarshals that into values.
type TTNDecoder struct {
	d    *json.Decoder
	gogo *GoGoJSONPb
}

// Decode reads a value from the reader and unmarshals v from JSON.
func (d *TTNDecoder) Decode(v any) error {
	if unmarshaler, ok := v.(jsonplugin.Unmarshaler); ok {
		var data json.RawMessage
		err := d.d.Decode(&data)
		if err != nil {
			return err
		}
		return jsonplugin.UnmarshalerConfig{}.Unmarshal(data, unmarshaler)
	}
	return GoGoDecoderWrapper{Decoder: d.d}.Decode(v)
}

// TTNEventStream returns a TTN JsonPb marshaler with double newlines for
// text/event-stream compatibility.
func TTNEventStream() runtime.Marshaler {
	return &ttnEventStream{TTNMarshaler: TTN()}
}

type ttnEventStream struct {
	*TTNMarshaler
}

func (*ttnEventStream) ContentType() string { return "text/event-stream" }

func (*ttnEventStream) Delimiter() []byte { return []byte{'\n', '\n'} }
