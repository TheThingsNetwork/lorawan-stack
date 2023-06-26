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
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/url"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	errFieldNotFound = errors.DefineNotFound("field_not_found", "field `{field}` not found")
	errInvalidType   = errors.DefineCorruption("invalid_type", "wrong type `{type}`")
	errInvalidValue  = errors.DefineCorruption("invalid_value", "wrong value `{value}`")
	errEncodingField = errors.DefineCorruption("encoding_field", "encoding field `{field}`")
	errDecodingField = errors.DefineCorruption("decoding_field", "decoding field `{field}`")
)

// UplinkMetadata contains the uplink metadata stored by the package.
type UplinkMetadata struct {
	RxMetadata []*api.RxMetadata `json:"rx_metadata"`
	ReceivedAt time.Time         `json:"received_at"`
}

// FromApplicationUplink cleans the ApplicationUplink to stored values for the UpLinkMetadata.
func (u *UplinkMetadata) FromApplicationUplink(msg *ttnpb.ApplicationUplink) error {
	u.ReceivedAt = msg.ReceivedAt.AsTime()

	for _, md := range msg.RxMetadata {
		rxmd := &api.RxMetadata{}
		if err := rxmd.FromProto(md); err != nil {
			return err
		}
		u.RxMetadata = append(u.RxMetadata, rxmd)
	}
	return nil
}

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
	// RecentMetadata are the metadatas from the recent uplink messages received by the gateway.
	RecentMetadata []*UplinkMetadata
}

const (
	queryField           = "query"
	multiFrameField      = "multi_frame"
	multiFrameWindowSize = "multi_frame_window_size"
	multiFrameWindowAge  = "multi_frame_window_age"
	serverURLField       = "server_url"
	tokenField           = "token"
	recentMDField        = "recent_metadata"
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

func toRecentMD(mds []*UplinkMetadata) (*structpb.Value, error) {
	gobBytes := new(bytes.Buffer)
	gobEncoder := gob.NewEncoder(gobBytes)
	if err := gobEncoder.Encode(mds); err != nil {
		return nil, errEncodingField.WithCause(err)
	}

	base64Bytes := new(bytes.Buffer)
	base64Encoder := base64.NewEncoder(base64.RawStdEncoding, base64Bytes)
	if _, err := base64Encoder.Write(gobBytes.Bytes()); err != nil {
		return nil, errEncodingField.WithCause(err).WithAttributes("field", recentMDField)
	}

	if err := base64Encoder.Close(); err != nil {
		return nil, errEncodingField.WithCause(err).WithAttributes("field", recentMDField)
	}

	return toString(base64Bytes.String()), nil
}

// Struct serializes the configuration to *structpb.Struct.
func (d *Data) Struct() (*structpb.Struct, error) {
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
	if len(d.RecentMetadata) > 0 {
		recentMD, err := toRecentMD(d.RecentMetadata)
		if err != nil {
			return nil, err
		}
		st.Fields[recentMDField] = recentMD
	}
	return st, nil
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

func recentMDFromValue(v *structpb.Value) ([]*UplinkMetadata, error) {
	sv, ok := v.Kind.(*structpb.Value_StringValue)
	if !ok {
		return nil, errInvalidType.WithAttributes("type", fmt.Sprintf("%T", v.Kind))
	}

	base64Bytes := bytes.NewBufferString(sv.StringValue)
	base64Decoder := base64.NewDecoder(base64.RawStdEncoding, base64Bytes)
	gobBytes := new(bytes.Buffer)
	if _, err := gobBytes.ReadFrom(base64Decoder); err != nil {
		return nil, errDecodingField.WithCause(err).WithAttributes("field", recentMDField)
	}

	var mds []*UplinkMetadata
	gobDecoder := gob.NewDecoder(gobBytes)
	if err := gobDecoder.Decode(&mds); err != nil {
		return nil, errDecodingField.WithCause(err).WithAttributes("field", recentMDField)
	}
	return mds, nil
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
		value, ok := fields[recentMDField]
		if ok {
			uplinks, err := recentMDFromValue(value)
			if err != nil {
				return err
			}
			d.RecentMetadata = uplinks
		}
	}
	return nil
}
