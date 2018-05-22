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

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestGRPCConversion(t *testing.T) {
	a := assertions.New(t)

	errDef := Define("test_grpc_conversion_err_def", "gRPC Conversion Error")
	a.So(FromGRPCStatus(errDef.GRPCStatus()).Definition, ShouldEqual, errDef)

	errHello := New("hello world").WithAttributes("foo", "bar")
	a.So(FromGRPCStatus(errHello.GRPCStatus()), ShouldEqual, errHello)
}

func TestJSONConversion(t *testing.T) {
	a := assertions.New(t)

	errHello := New("hello world").WithAttributes("foo", "bar")
	b, err := json.Marshal(errHello)
	a.So(err, should.BeNil)

	var unmarshaled Error
	err = json.Unmarshal(b, &unmarshaled)
	a.So(err, should.BeNil)
	a.So(unmarshaled, ShouldEqual, errHello)
}
