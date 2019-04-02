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

// Package cayennelpp contains the CayenneLPP payload formatter message processors.
package cayennelpp

import (
	"bytes"
	"context"
	"runtime/trace"

	lpp "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type host struct {
}

type decodedMap map[string]interface{}

// New creates and returns a new CayenneLPP payload encoder and decoder.
func New() messageprocessors.PayloadEncodeDecoder {
	return &host{}
}

var (
	errInput  = errors.DefineInvalidArgument("input", "invalid input")
	errOutput = errors.Define("output", "invalid output")
)

// Encode encodes the message's DecodedPayload to FRMPayload using CayenneLPP encoding.
func (h *host) Encode(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationDownlink, script string) error {
	defer trace.StartRegion(ctx, "encode message").End()

	decoded := msg.DecodedPayload
	if decoded == nil {
		return nil
	}
	m, err := gogoproto.Map(decoded)
	if err != nil {
		return errInput.WithCause(err)
	}
	encoder := lpp.NewEncoder()
	for name, value := range m {
		key, channel, err := parseName(name)
		if err != nil {
			continue
		}
		switch key {
		case valueKey:
			if val, ok := value.(float64); ok {
				encoder.AddPort(channel, float32(val))
			}
		}
	}
	msg.FRMPayload = encoder.Bytes()
	return nil
}

// Decode decodes the message's FRMPayload to DecodedPayload using CayenneLPP decoding.
func (h *host) Decode(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, msg *ttnpb.ApplicationUplink, script string) error {
	defer trace.StartRegion(ctx, "decode message").End()

	decoder := lpp.NewDecoder(bytes.NewBuffer(msg.FRMPayload))
	m := decodedMap(make(map[string]interface{}))
	if err := decoder.DecodeUplink(m); err != nil {
		return errOutput.WithCause(err)
	}
	s, err := gogoproto.Struct(m)
	if err != nil {
		return errOutput.WithCause(err)
	}
	msg.DecodedPayload = s
	return nil
}

func (d decodedMap) DigitalInput(channel, value uint8) {
	d[formatName(digitalInputKey, channel)] = value
}

func (d decodedMap) DigitalOutput(channel, value uint8) {
	d[formatName(digitalOutputKey, channel)] = value
}

func (d decodedMap) AnalogInput(channel uint8, value float32) {
	d[formatName(analogInputKey, channel)] = value
}

func (d decodedMap) AnalogOutput(channel uint8, value float32) {
	d[formatName(analogOutputKey, channel)] = value
}

func (d decodedMap) Luminosity(channel uint8, value uint16) {
	d[formatName(luminosityKey, channel)] = value
}

func (d decodedMap) Presence(channel, value uint8) {
	d[formatName(presenceKey, channel)] = value
}

func (d decodedMap) Temperature(channel uint8, celsius float32) {
	d[formatName(temperatureKey, channel)] = celsius
}

func (d decodedMap) RelativeHumidity(channel uint8, rh float32) {
	d[formatName(relativeHumidityKey, channel)] = rh
}

func (d decodedMap) Accelerometer(channel uint8, x, y, z float32) {
	d[formatName(accelerometerKey, channel)] = map[string]float32{
		"x": x,
		"y": y,
		"z": z,
	}
}

func (d decodedMap) BarometricPressure(channel uint8, hpa float32) {
	d[formatName(barometricPressureKey, channel)] = hpa
}

func (d decodedMap) Gyrometer(channel uint8, x, y, z float32) {
	d[formatName(gyrometerKey, channel)] = map[string]float32{
		"x": x,
		"y": y,
		"z": z,
	}
}

func (d decodedMap) GPS(channel uint8, latitude, longitude, altitude float32) {
	d[formatName(gpsKey, channel)] = map[string]float32{
		"latitude":  latitude,
		"longitude": longitude,
		"altitude":  altitude,
	}
}
