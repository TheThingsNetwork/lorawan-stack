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

package devicetemplates

import (
	"context"
	"io"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttnio "go.thethings.network/lorawan-stack/v3/pkg/util/io"
	"golang.org/x/net/html/charset"
)

// TTSJSON is the device template converter ID.
const TTSJSON = "the-things-stack"

type ttsJSON struct{}

// Format implements the devicetemplates.Converter interface.
func (*ttsJSON) Format() *ttnpb.EndDeviceTemplateFormat {
	return &ttnpb.EndDeviceTemplateFormat{
		Name:           "The Things Stack JSON",
		Description:    "File containing end devices in The Things Stack JSON format.",
		FileExtensions: []string{".json"},
	}
}

// Convert implements the devicetemplates.Converter interface.
func (*ttsJSON) Convert(_ context.Context, r io.Reader, f func(*ttnpb.EndDeviceTemplate) error) error {
	r, err := charset.NewReader(r, "application/json")
	if err != nil {
		return err
	}
	dec := ttnio.NewJSONDecoder(r)
	for {
		dev := &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{},
		}
		if err := dec.Decode(dev); err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			return nil
		}

		paths := ttnpb.NonZeroFields(dev, ttnpb.EndDeviceFieldPathsNestedWithoutWrappers...)
		paths = ttnpb.BottomLevelFields(paths)

		dev.Ids.DevAddr = nil
		paths = ttnpb.ExcludeFields(paths, "ids.dev_addr")

		tmpl := &ttnpb.EndDeviceTemplate{
			EndDevice: dev,
			FieldMask: ttnpb.FieldMask(paths...),
		}
		if err := f(tmpl); err != nil {
			return err
		}
	}
}

func init() {
	RegisterConverter(TTSJSON, &ttsJSON{})
}
