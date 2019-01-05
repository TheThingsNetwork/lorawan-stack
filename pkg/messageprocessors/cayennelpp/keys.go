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

package cayennelpp

import (
	"fmt"
	"strconv"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

const (
	valueKey              = "value"
	digitalInputKey       = "digital_in"
	digitalOutputKey      = "digital_out"
	analogInputKey        = "analog_in"
	analogOutputKey       = "analog_out"
	luminosityKey         = "luminosity"
	presenceKey           = "presence"
	temperatureKey        = "temperature"
	relativeHumidityKey   = "relative_humidity"
	accelerometerKey      = "accelerometer"
	barometricPressureKey = "barometric_pressure"
	gyrometerKey          = "gyrometer"
	gpsKey                = "gps"
)

func formatName(key string, channel uint8) string {
	return fmt.Sprintf("%s_%d", key, channel)
}

var (
	errInvalidKeyName = errors.DefineInvalidArgument("key_name", "invalid key name `{name}`")
	errInvalidChannel = errors.DefineInvalidArgument("channel", "invalid channel `{channel}`")
)

func parseName(name string) (string, uint8, error) {
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		return "", 0, errInvalidKeyName.WithAttributes("name", name)
	}
	key := strings.Join(parts[:len(parts)-1], "_")
	if key == "" {
		return "", 0, errInvalidKeyName.WithAttributes("name", name)
	}
	channel, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return "", 0, err
	}
	if channel < 0 || channel > 255 {
		return "", 0, errInvalidChannel.WithAttributes("channel", channel)
	}
	return key, uint8(channel), nil
}
