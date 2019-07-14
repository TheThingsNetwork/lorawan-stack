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

package devicetemplates_test

import (
	"context"
	"io"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/devicetemplates"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestConverters(t *testing.T) {
	a := assertions.New(t)

	a.So(GetConverter("test"), should.BeNil)

	converter := &mockConverter{
		EndDeviceTemplateFormat: ttnpb.EndDeviceTemplateFormat{
			Name:        "Foo",
			Description: "Bar",
		},
	}
	RegisterConverter("test", converter)
	a.So(GetConverter("test"), should.Equal, converter)
}

type mockConverter struct {
	ttnpb.EndDeviceTemplateFormat
}

func (c *mockConverter) Format() *ttnpb.EndDeviceTemplateFormat {
	return &c.EndDeviceTemplateFormat
}

func (c *mockConverter) Convert(context.Context, io.Reader, chan<- *ttnpb.EndDeviceTemplate) error {
	return nil
}
