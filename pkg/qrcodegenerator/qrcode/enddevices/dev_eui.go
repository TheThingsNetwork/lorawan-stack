// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package enddevices

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const formatIDDevEUI = "dev_eui"

type devEUIData struct {
	DevEUI types.EUI64
}

// MarshalText implements Data.
func (d *devEUIData) MarshalText() (text []byte, err error) {
	return d.DevEUI.MarshalText()
}

// UnmarshalText implements Data.
func (d *devEUIData) UnmarshalText(text []byte) error {
	return d.DevEUI.UnmarshalText(text)
}

// Validate implements Data.
func (d *devEUIData) Validate() error {
	return nil
}

// Encode implements Data.
func (d *devEUIData) Encode(dev *ttnpb.EndDevice) error {
	if dev.Ids.DevEui == nil {
		return errNoDevEUI.New()
	}
	*d = devEUIData{
		DevEUI: types.MustEUI64(dev.Ids.DevEui).OrZero(),
	}
	return nil
}

// EndDeviceTemplate implements Data.
func (d *devEUIData) EndDeviceTemplate() *ttnpb.EndDeviceTemplate {
	return &ttnpb.EndDeviceTemplate{
		EndDevice: &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				DevEui: d.DevEUI.Bytes(),
			},
		},
		FieldMask: ttnpb.FieldMask("ids.dev_eui"),
	}
}

// FormatID implements Data.
func (d *devEUIData) FormatID() string {
	return formatIDDevEUI
}

type devEUIFormat struct{}

// Format implements Format.
func (*devEUIFormat) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:        "DevEUI",
		Description: "LoRaWAN® DevEUI (hex encoded)",
		FieldMask:   ttnpb.FieldMask("ids.dev_eui"),
	}
}

// New implements Format.
func (*devEUIFormat) New() Data {
	return new(devEUIData)
}
