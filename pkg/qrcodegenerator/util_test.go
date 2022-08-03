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

package qrcodegenerator_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/enddevices"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

type mock struct {
	ids *ttnpb.EndDeviceIdentifiers
}

func (mock) Validate() error { return nil }

func (m *mock) Encode(dev *ttnpb.EndDevice) error {
	*m = mock{
		ids: dev.Ids,
	}
	return nil
}

func (*mock) EndDeviceTemplate() *ttnpb.EndDeviceTemplate { return nil }

func (*mock) FormatID() string {
	return "test"
}

func (m mock) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(
		"%s:%s",
		strings.ToUpper(hex.EncodeToString(m.ids.JoinEui)),
		strings.ToUpper(hex.EncodeToString(m.ids.DevEui))),
	), nil
}

func (*mock) UnmarshalText([]byte) error { return nil }

type mockFormat struct{}

func (mockFormat) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:        "Test",
		Description: "Test",
		FieldMask:   ttnpb.FieldMask("ids"),
	}
}

func (mockFormat) New() enddevices.Data {
	return new(mock)
}
