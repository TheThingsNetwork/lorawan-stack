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
	"fmt"
	"net/url"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	errFieldNotFound = errors.DefineNotFound("field_not_found", "field `{field}` not found")
	errInvalidType   = errors.DefineCorruption("invalid_type", "wrong type `{type}`")
	errInvalidValue  = errors.DefineCorruption("invalid_value", "wrong value `{value}`")
)

// QueryType enum defines the location query types of the package.
type QueryType uint8

// Value returns the protobuf value for the query type.
func (t QueryType) Value() *structpb.Value {
	var s string
	switch t {
	case QUERY_TOARSSI:
		s = "TOARSSI"
	case QUERY_GNSS:
		s = "GNSS"
	case QUERY_TOAWIFI:
		s = "TOAWIFI"
	default:
		panic("invalid query type")
	}
	return &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: s,
		},
	}
}

// FromValue sets the query type from a protobuf value.
func (t *QueryType) FromValue(v *structpb.Value) error {
	sv, ok := v.Kind.(*structpb.Value_StringValue)
	if !ok {
		return errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	switch sv.StringValue {
	case "TOARSSI":
		*t = QUERY_TOARSSI
	case "GNSS":
		*t = QUERY_GNSS
	case "TOAWIFI":
		*t = QUERY_TOAWIFI
	default:
		return errInvalidValue.WithAttributes("value", sv.StringValue)
	}
	return nil
}

const (
	// QUERY_TOARSSI uses the TOA and RSSI information from the gateway metadata to compute the location of the end device.
	QUERY_TOARSSI QueryType = iota
	// QUERY_GNSS uses the GNSS scan operations payload of the LR1110 transceiver.
	QUERY_GNSS
	// QUERY_TOAWIFI uses the TOA and RSSI information, in addition to nearby WiFi access points.
	QUERY_TOAWIFI
)

// UplinkMD contains the uplink metadata stored by the package.
type UplinkMD struct {
	RxMetadata []*ttnpb.RxMetadata
	ReceivedAt time.Time
}

// Data contains the package configuration.
type Data struct {
	// Query is the query type used by the package.
	Query QueryType
	// MultiFrame enables multi frame requests for TOARSSI queries.
	MultiFrame bool
	// MultiFrameWindowSize represents the number of historical frames to consider for the query.
	// A window size of 0 automatically determines the number of frames based on the first byte
	// of the uplink message.
	MultiFrameWindowSize int
	// MultiFrameWindowAge limits the maximum age of the historical frames considered for the query.
	MultiFrameWindowAge time.Duration
	// ServerURL represents the remote server to which the GLS queries are sent.
	ServerURL *url.URL
	// Token is the API token to be used when comunicating with the GLS server.
	Token string
	// UplinkMDs are the metadatas from the recent uplink messages received by the gateway.
	UplinkMDs []*UplinkMD
}

const (
	queryField           = "query"
	multiFrameField      = "multi_frame"
	multiFrameWindowSize = "multi_frame_window_size"
	multiFrameWindowAge  = "multi_frame_window_age"
	serverURLField       = "server_url"
	tokenField           = "token"
	uplinksField         = "uplinks"
)

func toString(s string) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: s,
		},
	}
}

func toBool(b bool) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_BoolValue{
			BoolValue: b,
		},
	}
}

func toFloat64(f float64) *structpb.Value {
	return &structpb.Value{
		Kind: &structpb.Value_NumberValue{
			NumberValue: f,
		},
	}
}

func toGatewayIDsStruct(gatewayIDs *ttnpb.GatewayIdentifiers) *structpb.Value {
	gatewayIDsMap := make(map[string]*structpb.Value)
	gatewayIDsMap["gateway_id"] = toString(gatewayIDs.GatewayId)

	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: gatewayIDsMap,
			},
		},
	}
}

func toLocationStruct(location *ttnpb.Location) *structpb.Value {
	locationMap := make(map[string]*structpb.Value)
	locationMap["latitude"] = toFloat64(location.Latitude)
	locationMap["longitude"] = toFloat64(location.Longitude)
	locationMap["altitude"] = toFloat64(float64(location.Altitude))
	locationMap["accuracy"] = toFloat64(float64(location.Accuracy))
	locationMap["source"] = toFloat64(float64(location.Source))

	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: locationMap,
			},
		},
	}
}

func toRxMetadataStruct(rx *ttnpb.RxMetadata) *structpb.Value {
	rxMetadataMap := make(map[string]*structpb.Value)
	rxMetadataMap["gateway_ids"] = toGatewayIDsStruct(rx.GatewayIds)
	rxMetadataMap["antenna_index"] = toFloat64(float64(rx.AntennaIndex))
	rxMetadataMap["fine_timestamp"] = toFloat64(float64(rx.FineTimestamp))
	rxMetadataMap["location"] = toLocationStruct(rx.Location)
	rxMetadataMap["rssi"] = toFloat64(float64(rx.Rssi))
	rxMetadataMap["snr"] = toFloat64(float64(rx.Snr))

	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: rxMetadataMap,
			},
		},
	}
}

func toRxMetadataStructs(rxs []*ttnpb.RxMetadata) *structpb.Value {
	rxMetadataList := make([]*structpb.Value, 0, len(rxs))
	for _, rx := range rxs {
		rxMetadataList = append(rxMetadataList, toRxMetadataStruct(rx))
	}
	return &structpb.Value{
		Kind: &structpb.Value_ListValue{
			ListValue: &structpb.ListValue{
				Values: rxMetadataList,
			},
		},
	}
}

func toUplinkStruct(up *UplinkMD) *structpb.Value {
	uplinkMap := make(map[string]*structpb.Value)
	uplinkMap["received_at"] = toString(up.ReceivedAt.Format(time.RFC3339Nano))
	uplinkMap["rx_metadata"] = toRxMetadataStructs(up.RxMetadata)
	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: uplinkMap,
			},
		},
	}
}

func toUplinkStructs(ups []*UplinkMD) *structpb.Value {
	uplinkList := make([]*structpb.Value, len(ups))
	for i, up := range ups {
		uplinkList[i] = toUplinkStruct(up)
	}
	return &structpb.Value{
		Kind: &structpb.Value_ListValue{
			ListValue: &structpb.ListValue{
				Values: uplinkList,
			},
		},
	}
}

// Struct serializes the configuration to *structpb.Struct.
func (d *Data) Struct() *structpb.Struct {
	st := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			queryField: d.Query.Value(),
			tokenField: toString(d.Token),
		},
	}
	if d.ServerURL != nil {
		st.Fields[serverURLField] = toString(d.ServerURL.String())
	}
	if d.MultiFrame {
		st.Fields[multiFrameField] = toBool(d.MultiFrame)
	}
	if d.MultiFrameWindowSize > 0 {
		st.Fields[multiFrameWindowSize] = toFloat64(float64(d.MultiFrameWindowSize))
	}
	if d.MultiFrameWindowAge > 0 {
		st.Fields[multiFrameWindowAge] = toFloat64(float64(d.MultiFrameWindowAge / time.Minute))
	}
	if len(d.UplinkMDs) > 0 {
		st.Fields[uplinksField] = toUplinkStructs(d.UplinkMDs)
	}
	return st
}

func stringFromValue(v *structpb.Value) (string, error) {
	sv, ok := v.Kind.(*structpb.Value_StringValue)
	if !ok {
		return "", errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	return sv.StringValue, nil
}

func boolFromValue(v *structpb.Value) (bool, error) {
	bv, ok := v.Kind.(*structpb.Value_BoolValue)
	if !ok {
		return false, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	return bv.BoolValue, nil
}

func float64FromValue(v *structpb.Value) (float64, error) {
	fv, ok := v.Kind.(*structpb.Value_NumberValue)
	if !ok {
		return 0.0, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	return fv.NumberValue, nil
}

func gatewayIDsFromValue(v *structpb.Value) (*ttnpb.GatewayIdentifiers, error) {
	sv, ok := v.Kind.(*structpb.Value_StructValue)
	if !ok {
		return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	gatewayIDs := &ttnpb.GatewayIdentifiers{}
	if v, ok := sv.StructValue.Fields["gateway_id"]; ok {
		sv, ok := v.Kind.(*structpb.Value_StringValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		gatewayIDs.GatewayId = sv.StringValue
	}
	return gatewayIDs, nil
}

func locationFromValue(v *structpb.Value) (*ttnpb.Location, error) {
	sv, ok := v.Kind.(*structpb.Value_StructValue)
	if !ok {
		return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	location := &ttnpb.Location{}
	if v, ok := sv.StructValue.Fields["latitude"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		location.Latitude = sv.NumberValue
	}
	if v, ok := sv.StructValue.Fields["longitude"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		location.Longitude = sv.NumberValue
	}
	if v, ok := sv.StructValue.Fields["altitude"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		location.Altitude = int32(sv.NumberValue)
	}
	if v, ok := sv.StructValue.Fields["accuracy"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		location.Accuracy = int32(sv.NumberValue)
	}
	if v, ok := sv.StructValue.Fields["source"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		location.Source = ttnpb.LocationSource(sv.NumberValue)
	}
	return location, nil
}

func rxMetadataFromValue(v *structpb.Value) (*ttnpb.RxMetadata, error) {
	sv, ok := v.Kind.(*structpb.Value_StructValue)
	if !ok {
		return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	rxMetadata := &ttnpb.RxMetadata{}
	if v, ok := sv.StructValue.Fields["gateway_ids"]; ok {
		gatewayIDs, err := gatewayIDsFromValue(v)
		if err != nil {
			return nil, err
		}
		rxMetadata.GatewayIds = gatewayIDs
	}
	if v, ok := sv.StructValue.Fields["antenna_index"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		rxMetadata.AntennaIndex = uint32(sv.NumberValue)
	}
	if v, ok := sv.StructValue.Fields["fine_timestamp"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		rxMetadata.FineTimestamp = uint64(sv.NumberValue)
	}
	if v, ok := sv.StructValue.Fields["location"]; ok {
		location, err := locationFromValue(v)
		if err != nil {
			return nil, err
		}
		rxMetadata.Location = location
	}
	if v, ok := sv.StructValue.Fields["rssi"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		rxMetadata.Rssi = float32(sv.NumberValue)
	}
	if v, ok := sv.StructValue.Fields["snr"]; ok {
		sv, ok := v.Kind.(*structpb.Value_NumberValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		rxMetadata.Snr = float32(sv.NumberValue)
	}
	return rxMetadata, nil
}

func rxMetadataFromValues(vs []*structpb.Value) ([]*ttnpb.RxMetadata, error) {
	rxs := make([]*ttnpb.RxMetadata, 0, len(vs))
	for _, v := range vs {
		rx, err := rxMetadataFromValue(v)
		if err != nil {
			return nil, err
		}
		rxs = append(rxs, rx)
	}
	return rxs, nil
}

func uplinkFromValue(v *structpb.Value) (*UplinkMD, error) {
	sv, ok := v.Kind.(*structpb.Value_StructValue)
	if !ok {
		return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}
	uplink := &UplinkMD{}
	if v, ok := sv.StructValue.Fields["received_at"]; ok {
		sv, ok := v.Kind.(*structpb.Value_StringValue)
		if !ok {
			return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
		}
		receivedAt, err := time.Parse(time.RFC3339Nano, sv.StringValue)
		if err != nil {
			return nil, err
		}
		uplink.ReceivedAt = receivedAt
	}
	if v, ok := sv.StructValue.Fields["rx_metadata"]; ok {
		rxMetadata, err := rxMetadataFromValues(v.GetListValue().GetValues())
		if err != nil {
			return nil, err
		}
		uplink.RxMetadata = rxMetadata
	}
	return uplink, nil
}

func uplinkFromValues(vs []*structpb.Value) ([]*UplinkMD, error) {
	ups := make([]*UplinkMD, 0, len(vs))
	for _, v := range vs {
		up, err := uplinkFromValue(v)
		if err != nil {
			return nil, err
		}
		ups = append(ups, up)
	}
	return ups, nil
}

// FromStruct deserializes the configuration from *structpb.Struct.
func (d *Data) FromStruct(st *structpb.Struct) error {
	fields := st.GetFields()
	{
		value, ok := fields[queryField]
		if !ok {
			return errFieldNotFound.WithAttributes("field", queryField)
		}
		if err := d.Query.FromValue(value); err != nil {
			return err
		}
	}
	{
		value, ok := fields[multiFrameField]
		if ok {
			multiFrame, err := boolFromValue(value)
			if err != nil {
				return err
			}
			d.MultiFrame = multiFrame
		}
	}
	{
		value, ok := fields[multiFrameWindowSize]
		if ok {
			windowSize, err := float64FromValue(value)
			if err != nil {
				return err
			}
			d.MultiFrameWindowSize = int(windowSize)
		}
	}
	{
		value, ok := fields[multiFrameWindowAge]
		if ok {
			windowAge, err := float64FromValue(value)
			if err != nil {
				return err
			}
			d.MultiFrameWindowAge = time.Duration(windowAge) * time.Minute
		}
	}
	{
		value, ok := fields[tokenField]
		if !ok {
			return errFieldNotFound.WithAttributes("field", tokenField)
		}
		token, err := stringFromValue(value)
		if err != nil {
			return err
		}
		d.Token = token
	}
	{
		value, ok := fields[serverURLField]
		if ok {
			serverURL, err := stringFromValue(value)
			if err != nil {
				return err
			}
			u, err := url.Parse(serverURL)
			if err != nil {
				return err
			}
			d.ServerURL = u
		}
	}
	{
		value, ok := fields[uplinksField]
		if ok {
			uplinks, err := uplinkFromValues(value.GetListValue().GetValues())
			if err != nil {
				return err
			}
			d.UplinkMDs = uplinks
		}
	}
	return nil
}
