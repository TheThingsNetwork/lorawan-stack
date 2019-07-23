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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Converter converts a binary file in end device templates.
type Converter interface {
	Format() *ttnpb.EndDeviceTemplateFormat
	Convert(context.Context, io.Reader, chan<- *ttnpb.EndDeviceTemplate) error
}

var converters = map[string]Converter{}

// GetConverter returns the converter by ID.
func GetConverter(id string) Converter {
	return converters[id]
}

// RegisterConverter registers the given converter.
// Existing registrations with the same ID will be overwritten.
// This function is not goroutine-safe.
func RegisterConverter(id string, c Converter) {
	converters[id] = c
}
