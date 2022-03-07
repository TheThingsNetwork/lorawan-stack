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

// Package qrcode implements working with QR codes.
package qrcode

import (
	"encoding"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Data represents QR code data.
type Data interface {
	Validate() error
	// EndDeviceInfo returns the format ID as a string and the End Device Template corresponding to the QR code data.
	EndDeviceInfo() (string, *ttnpb.EndDeviceTemplate)
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}
