// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GatewayIDs contains the gateway identifiers stored by the package.
type GatewayIDs struct {
	GatewayID string `json:"gateway_id"`
}

// ToProto converts the GatewayIDs to a protobuf representation.
func (g GatewayIDs) ToProto() *ttnpb.GatewayIdentifiers {
	return &ttnpb.GatewayIdentifiers{
		GatewayId: g.GatewayID,
	}
}

// FromProto converts the GatewayIDs from a protobuf representation.
func (g *GatewayIDs) FromProto(pb *ttnpb.GatewayIdentifiers) error {
	g.GatewayID = pb.GatewayId
	return nil
}

// Location contains the location stored by the package.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  int32   `json:"altitude"`
	Accuracy  int32   `json:"accuracy"`
	Source    int32   `json:"source"`
}

// ToProto converts the Location to a protobuf representation.
func (l Location) ToProto() *ttnpb.Location {
	return &ttnpb.Location{
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Altitude:  l.Altitude,
		Accuracy:  l.Accuracy,
		Source:    ttnpb.LocationSource(l.Source),
	}
}

// FromProto converts the Location from a protobuf representation.
func (l *Location) FromProto(pb *ttnpb.Location) error {
	l.Latitude = pb.Latitude
	l.Longitude = pb.Longitude
	l.Altitude = pb.Altitude
	l.Accuracy = pb.Accuracy
	l.Source = int32(pb.Source)
	return nil
}

// RxMetadata contains the RxMetadata stored by the package.
type RxMetadata struct {
	GatewayIDs    *GatewayIDs `json:"gateway_ids"`
	AntennaIndex  uint32      `json:"antenna_index"`
	FineTimestamp uint64      `json:"fine_timestamp"`
	Location      *Location   `json:"location"`
	RSSI          float32     `json:"rssi"`
	SNR           float32     `json:"snr"`
}

// ToProto converts the RxMetadata to a protobuf representation.
func (r RxMetadata) ToProto() *ttnpb.RxMetadata {
	return &ttnpb.RxMetadata{
		GatewayIds:    r.GatewayIDs.ToProto(),
		AntennaIndex:  r.AntennaIndex,
		FineTimestamp: r.FineTimestamp,
		Location:      r.Location.ToProto(),
		Rssi:          r.RSSI,
		Snr:           r.SNR,
	}
}

// FromProto converts the RxMetadata from a protobuf representation.
func (r *RxMetadata) FromProto(pb *ttnpb.RxMetadata) error {
	r.GatewayIDs = &GatewayIDs{}
	if err := r.GatewayIDs.FromProto(pb.GatewayIds); err != nil {
		return err
	}

	r.Location = &Location{}
	if err := r.Location.FromProto(pb.Location); err != nil {
		return err
	}

	r.AntennaIndex = pb.AntennaIndex
	r.FineTimestamp = pb.FineTimestamp
	r.RSSI = pb.Rssi
	r.SNR = pb.Snr

	return nil
}

// UplinkMetadata contains the uplink metadata stored by the package.
type UplinkMetadata struct {
	RxMetadata []*RxMetadata `json:"rx_metadata"`
	ReceivedAt time.Time     `json:"received_at"`
}

// FromApplicationUplink cleans the ApplicationUplink to stored values for the UpLinkMetadata.
func (u *UplinkMetadata) FromApplicationUplink(msg *ttnpb.ApplicationUplink) error {
	u.ReceivedAt = msg.ReceivedAt.AsTime()

	for _, md := range msg.RxMetadata {
		rxmd := &RxMetadata{}
		if err := rxmd.FromProto(md); err != nil {
			return err
		}
		u.RxMetadata = append(u.RxMetadata, rxmd)
	}
	return nil
}

// ToProto converts the UplinkMetadata to a protobuf representation of a clean ApplicationUplink.
func (u *UplinkMetadata) ToProto() *ttnpb.ApplicationUplink {
	msg := &ttnpb.ApplicationUplink{
		ReceivedAt: timestamppb.New(u.ReceivedAt),
	}

	for _, md := range u.RxMetadata {
		msg.RxMetadata = append(msg.RxMetadata, md.ToProto())
	}
	return msg
}
