// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package cayennelpp_test

import (
	"context"
	"testing"

	lpp "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	"github.com/TheThingsNetwork/ttn/pkg/gogoproto"
	"github.com/TheThingsNetwork/ttn/pkg/messageprocessors/cayennelpp"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestEncode(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	host := cayennelpp.New()

	model := &ttnpb.EndDeviceVersion{
		EndDeviceModel: ttnpb.EndDeviceModel{
			BrandID: "The Things Products",
			ModelID: "The Things Uno",
		},

		HardwareVersion: "1.0",
		FirmwareVersion: "1.0.0",
	}

	// Happy flow.
	{
		message := &ttnpb.DownlinkMessage{
			Payload: ttnpb.Message{
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						DecodedPayload: &types.Struct{
							Fields: map[string]*types.Value{
								"value_2": {
									Kind: &types.Value_NumberValue{
										NumberValue: -50.51,
									},
								},
							},
						},
					},
				},
			},
		}

		output, err := host.Encode(ctx, message, model, "")
		a.So(err, should.BeNil)
		a.So(output.Payload.GetMACPayload().FRMPayload, should.Resemble, []byte{2, 236, 69})
	}

	// Test resilience against custom fields from the user. Should be fine.
	{
		message := &ttnpb.DownlinkMessage{
			Payload: ttnpb.Message{
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						DecodedPayload: &types.Struct{
							Fields: map[string]*types.Value{
								"custom": {
									Kind: &types.Value_NumberValue{
										NumberValue: 8,
									},
								},
								"digital_in_8": {
									Kind: &types.Value_StringValue{
										StringValue: "shouldn't be a string",
									},
								},
								"custom_5": {
									Kind: &types.Value_NumberValue{
										NumberValue: 5,
									},
								},
								"accelerometer_1": {
									Kind: &types.Value_StructValue{
										StructValue: &types.Struct{
											Fields: map[string]*types.Value{
												"x": {
													Kind: &types.Value_StringValue{
														StringValue: "test",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		output, err := host.Encode(ctx, message, model, "")
		a.So(err, should.BeNil)
		a.So(output.Payload.GetMACPayload().FRMPayload, should.BeEmpty)
	}
}

func TestDecode(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	host := cayennelpp.New()

	model := &ttnpb.EndDeviceVersion{
		EndDeviceModel: ttnpb.EndDeviceModel{
			BrandID: "The Things Products",
			ModelID: "The Things Uno",
		},
		HardwareVersion: "1.0",
		FirmwareVersion: "1.0.0",
	}

	message := &ttnpb.UplinkMessage{
		Payload: ttnpb.Message{
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FRMPayload: []byte{
						1, lpp.DigitalInput, 255,
						2, lpp.DigitalOutput, 100,
						3, lpp.AnalogInput, 21, 74,
						4, lpp.AnalogOutput, 234, 182,
						5, lpp.Luminosity, 1, 244,
						6, lpp.Presence, 50,
						7, lpp.Temperature, 255, 100,
						8, lpp.RelativeHumidity, 99,
						9, lpp.Accelerometer, 254, 88, 0, 15, 6, 130,
						10, lpp.BarometricPressure, 41, 239,
						11, lpp.Gyrometer, 1, 99, 2, 49, 254, 102,
						12, lpp.GPS, 7, 253, 135, 0, 190, 245, 0, 8, 106,
					},
				},
			},
		},
	}

	output, err := host.Decode(ctx, message, model, "")
	a.So(err, should.BeNil)
	m, err := gogoproto.Map(output.Payload.GetMACPayload().DecodedPayload)
	a.So(err, should.BeNil)
	a.So(m, should.HaveLength, 12)
	a.So(m["digital_in_1"], should.Equal, 255)
	a.So(m["digital_out_2"], should.Equal, 100)
	a.So(m["analog_in_3"], should.AlmostEqual, 54.5, 0.00001)
	a.So(m["analog_out_4"], should.AlmostEqual, -54.5, 0.00001)
	a.So(m["luminosity_5"], should.Equal, 500)
	a.So(m["presence_6"], should.Equal, 50)
	a.So(m["temperature_7"], should.AlmostEqual, -15.6, 0.00001)
	a.So(m["relative_humidity_8"], should.AlmostEqual, 49.5, 0.00001)
	a.So(m["accelerometer_9"].(map[string]interface{})["x"], should.AlmostEqual, -0.424, 0.00001)
	a.So(m["accelerometer_9"].(map[string]interface{})["y"], should.AlmostEqual, 0.015, 0.00001)
	a.So(m["accelerometer_9"].(map[string]interface{})["z"], should.AlmostEqual, 1.666, 0.00001)
	a.So(m["barometric_pressure_10"], should.AlmostEqual, 1073.5, 0.00001)
	a.So(m["gyrometer_11"].(map[string]interface{})["x"], should.AlmostEqual, 3.55, 0.00001)
	a.So(m["gyrometer_11"].(map[string]interface{})["y"], should.AlmostEqual, 5.61, 0.00001)
	a.So(m["gyrometer_11"].(map[string]interface{})["z"], should.AlmostEqual, -4.10, 0.00001)
	a.So(m["gps_12"].(map[string]interface{})["latitude"], should.AlmostEqual, 52.3655, 0.00001)
	a.So(m["gps_12"].(map[string]interface{})["longitude"], should.AlmostEqual, 4.8885, 0.00001)
	a.So(m["gps_12"].(map[string]interface{})["altitude"], should.AlmostEqual, 21.54, 0.00001)
}
