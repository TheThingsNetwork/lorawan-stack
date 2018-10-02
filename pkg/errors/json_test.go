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

package errors_test

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	_ "go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestJSONConversion(t *testing.T) {
	a := assertions.New(t)

	errDef := errors.Define("test_json_conversion_err_def", "JSON Conversion Error", "foo")

	b, err := json.Marshal(errDef)
	a.So(err, should.BeNil)

	var unmarshaledDef errors.Definition
	err = json.Unmarshal(b, &unmarshaledDef)
	a.So(err, should.BeNil)
	a.So(unmarshaledDef, should.EqualErrorOrDefinition, errDef)

	errHello := errDef.WithAttributes("foo", "bar", "baz", "qux")
	errHelloExpected := errDef.WithAttributes("foo", "bar")

	b, err = json.Marshal(errHello)
	a.So(err, should.BeNil)

	var unmarshaled errors.Error
	err = json.Unmarshal(b, &unmarshaled)
	a.So(err, should.BeNil)
	a.So(unmarshaled, should.EqualErrorOrDefinition, errHelloExpected)
}
