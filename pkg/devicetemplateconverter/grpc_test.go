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

package devicetemplateconverter_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/component/test"
	. "go.thethings.network/lorawan-stack/pkg/devicetemplateconverter"
	"go.thethings.network/lorawan-stack/pkg/devicetemplates"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestConvertEndDeviceTemplate(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	devicetemplates.RegisterConverter("test", &mockConverter{
		EndDeviceTemplateFormat: ttnpb.EndDeviceTemplateFormat{
			Name:        "Test",
			Description: "Test",
		},
		ConvertFunc: func(ctx context.Context, r io.Reader, ch chan<- *ttnpb.EndDeviceTemplate) error {
			reader := bufio.NewReader(r)
			for {
				b, err := reader.ReadByte()
				if err != nil {
					if err == io.EOF {
						close(ch)
						return nil
					}
					return err
				}
				ch <- &ttnpb.EndDeviceTemplate{
					EndDevice: ttnpb.EndDevice{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							DeviceID: fmt.Sprintf("sn-%d", b),
						},
					},
					FieldMask: pbtypes.FieldMask{
						Paths: []string{"ids.device_id"},
					},
				}
			}
		},
	})

	c := NewComponent(t, &component.Config{})
	test.Must(New(c, &Config{
		Enabled: []string{"test"},
	}))
	test.Must(c.Start(), nil)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_TEMPLATE_CONVERTER)

	client := ttnpb.NewEndDeviceTemplateConverterClient(c.LoopbackConn())

	formats, err := client.ListFormats(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
	a.So(formats.Formats, should.HaveSameElementsDeep, map[string]*ttnpb.EndDeviceTemplateFormat{
		"test": {
			Name:        "Test",
			Description: "Test",
		},
	})

	stream, err := client.Convert(ctx, &ttnpb.ConvertEndDeviceTemplateRequest{
		FormatID: "test",
		Data:     []byte{0x1, 0x2},
	})
	a.So(err, should.BeNil)
	tmpls := make([]*ttnpb.EndDeviceTemplate, 0, 2)
	for {
		tmpl, err := stream.Recv()
		if err == io.EOF {
			break
		}
		tmpls = append(tmpls, tmpl)
	}
	a.So(tmpls, should.Resemble, []*ttnpb.EndDeviceTemplate{
		{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "sn-1",
				},
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids.device_id"},
			},
		},
		{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "sn-2",
				},
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids.device_id"},
			},
		},
	})
}
