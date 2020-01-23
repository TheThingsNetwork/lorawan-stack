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

package messages

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestType(t *testing.T) {
	a := assertions.New(t)
	msg := Version{
		Station:  "test",
		Firmware: "2.0.0",
		Package:  "test",
		Model:    "test",
		Protocol: 2,
	}

	data, err := json.Marshal(msg)
	a.So(err, should.BeNil)

	mt, err := Type(data)
	a.So(err, should.BeNil)
	a.So(mt, should.Equal, TypeUpstreamVersion)
}

func TestIsProduction(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Message          Version
		ExpectedResponse bool
	}{
		{
			Name:             "EmptyMessage",
			Message:          Version{},
			ExpectedResponse: false,
		},
		{
			Name: "EmptyMessage1",
			Message: Version{
				Features: "",
			},
			ExpectedResponse: false,
		},
		{
			Name: "NonProduction",
			Message: Version{
				Features: "gps rmtsh",
			},
			ExpectedResponse: false,
		},
		{
			Name: "Production",
			Message: Version{
				Features: "prod",
			},
			ExpectedResponse: true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(tc.Message.IsProduction(), should.Equal, tc.ExpectedResponse)
		})
	}
}
