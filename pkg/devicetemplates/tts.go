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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttnio "go.thethings.network/lorawan-stack/v3/pkg/util/io"
)

type tts struct{}

// Format implements the devicetemplates.Converter interface.
func (t *tts) Format() *ttnpb.EndDeviceTemplateFormat {
	return &ttnpb.EndDeviceTemplateFormat{
		Name:           "The Things Stack JSON",
		Description:    "JSON file containing end devices exported in the v3 JSON format.",
		FileExtensions: []string{".json"},
	}
}

// Convert implements the devicetemplates.Converter interface.
func (t *tts) Convert(ctx context.Context, r io.Reader, ch chan<- *ttnpb.EndDeviceTemplate) error {
	defer close(ch)

	dec := ttnio.NewJSONDecoder(r)

	for {
		dev := ttnpb.EndDevice{}
		paths, err := dec.Decode(&dev)
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		paths = append(paths, "supports_join")

		tmpl := &ttnpb.EndDeviceTemplate{
			EndDevice: dev,
			FieldMask: pbtypes.FieldMask{
				Paths: paths,
			},
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ch <- tmpl:
		}
	}
}

func init() {
	RegisterConverter("the-things-stack", &tts{})
}
