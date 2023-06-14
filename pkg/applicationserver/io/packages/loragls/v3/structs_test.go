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

package loracloudgeolocationv3

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGatewayIDsConversion(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	expectedProto := &ttnpb.GatewayIdentifiers{
		GatewayId: "test-gateway",
	}
	expectedData := &GatewayIDs{
		GatewayID: "test-gateway",
	}

	actualData := &GatewayIDs{}
	if err := actualData.FromProto(expectedProto); err != nil {
		t.Fatalf("FromProto failed: %v", err)
	}

	actualProto := actualData.ToProto()
	a.So(actualProto, assertions.ShouldResemble, expectedProto)
	a.So(actualData, assertions.ShouldResemble, expectedData)
}

func TestLocationConversion(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	expectedProto := &ttnpb.Location{
		Latitude:  52.37403,
		Longitude: 4.88969,
		Altitude:  10,
		Accuracy:  65,
		Source:    ttnpb.LocationSource_SOURCE_GPS,
	}
	expectedData := &Location{
		Latitude:  52.37403,
		Longitude: 4.88969,
		Altitude:  10,
		Accuracy:  65,
		Source:    1,
	}

	actualData := &Location{}
	if err := actualData.FromProto(expectedProto); err != nil {
		t.Fatalf("FromProto failed: %v", err)
	}

	actualProto := actualData.ToProto()
	a.So(actualProto, assertions.ShouldResemble, expectedProto)
	a.So(actualData, assertions.ShouldResemble, expectedData)
}

func TestRxMetadataConversion(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	expectedProto := &ttnpb.RxMetadata{
		GatewayIds: &ttnpb.GatewayIdentifiers{
			GatewayId: "test-gateway",
		},
		FineTimestamp: 1234567890,
		Rssi:          -10,
		Snr:           5,
		Location: &ttnpb.Location{
			Latitude:  52.37403,
			Longitude: 4.88969,
			Altitude:  10,
			Accuracy:  65,
			Source:    ttnpb.LocationSource_SOURCE_GPS,
		},
	}
	expectedData := &RxMetadata{
		GatewayIDs: &GatewayIDs{
			GatewayID: "test-gateway",
		},
		FineTimestamp: 1234567890,
		RSSI:          -10,
		SNR:           5,
		Location: &Location{
			Latitude:  52.37403,
			Longitude: 4.88969,
			Altitude:  10,
			Accuracy:  65,
			Source:    1,
		},
	}

	actualData := &RxMetadata{}
	if err := actualData.FromProto(expectedProto); err != nil {
		t.Fatalf("FromProto failed: %v", err)
	}

	actualProto := actualData.ToProto()
	a.So(actualProto, assertions.ShouldResemble, expectedProto)
	a.So(actualData, assertions.ShouldResemble, expectedData)
}

func TestUplinkMessageConversion(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	now := time.Now().UTC().Truncate(time.Second)

	expectedProto := &ttnpb.ApplicationUplink{
		RxMetadata: []*ttnpb.RxMetadata{
			{
				GatewayIds: &ttnpb.GatewayIdentifiers{
					GatewayId: "test-gateway",
				},
				FineTimestamp: 1234567890,
				Rssi:          -10,
				Snr:           5,
				Location: &ttnpb.Location{
					Latitude:  52.37403,
					Longitude: 4.88969,
					Altitude:  10,
					Accuracy:  65,
					Source:    ttnpb.LocationSource_SOURCE_GPS,
				},
			},
		},
		ReceivedAt: timestamppb.New(now),
	}
	expectedData := &UplinkMetadata{
		RxMetadata: []*RxMetadata{
			{
				GatewayIDs: &GatewayIDs{
					GatewayID: "test-gateway",
				},
				FineTimestamp: 1234567890,
				RSSI:          -10,
				SNR:           5,
				Location: &Location{
					Latitude:  52.37403,
					Longitude: 4.88969,
					Altitude:  10,
					Accuracy:  65,
					Source:    1,
				},
			},
		},
		ReceivedAt: now,
	}

	actualData := &UplinkMetadata{}
	if err := actualData.FromApplicationUplink(expectedProto); err != nil {
		t.Fatalf("FromProto failed: %v", err)
	}

	actualProto := actualData.ToProto()
	a.So(actualProto, assertions.ShouldResemble, expectedProto)
	a.So(actualData, assertions.ShouldResemble, expectedData)
}
