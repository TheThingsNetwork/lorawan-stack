// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package goproto

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type jsonMarshaler struct {
	Text string
}

func (m jsonMarshaler) MarshalJSON() ([]byte, error) {
	return bytes.ToUpper([]byte(`"` + m.Text + `"`)), nil
}

func (m *jsonMarshaler) UnmarshalJSON(b []byte) error {
	m.Text = string(bytes.ToLower(bytes.Trim(b, `"`)))
	return nil
}

func TestStructProto(t *testing.T) {
	a := assertions.New(t)

	ptr := "ptr"
	m := map[string]interface{}{
		"foo":            "bar",
		"ptr":            &ptr,
		"answer":         42,
		"answer.precise": 42.0,
		"works":          true,
		"empty":          nil,
		"list":           []string{"a", "b", "c"},
		"map":            map[string]string{"foo": "bar"},
		"eui":            types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		"jsonMarshaler":  &jsonMarshaler{Text: "testtext"},
	}
	o := MapFromProto(MapProto(m))
	mJSON, _ := json.Marshal(m)
	oJSON, _ := json.Marshal(o)
	a.So(string(oJSON), should.Resemble, string(mJSON))
}
