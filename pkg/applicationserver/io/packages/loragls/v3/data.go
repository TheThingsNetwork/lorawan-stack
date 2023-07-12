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
	urlutil "go.thethings.network/lorawan-stack/v3/pkg/util/url"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	errFieldRequired = errors.DefineNotFound("field_required", "field `{field}` is required")
	errInvalidType   = errors.DefineCorruption("invalid_type", "wrong type `{type}`")
	errInvalidValue  = errors.DefineCorruption("invalid_value", "wrong value `{value}`")
	errEncodingField = errors.DefineCorruption("encoding_field", "encoding field `{field}`")
	errDecodingField = errors.DefineCorruption("decoding_field", "decoding field `{field}`")

	// ErrEmptyData is a sentinel error that indicates that the data is empty.
	ErrEmptyData = errors.DefineInvalidArgument("empty_data", "empty data")
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
	return structpb.NewStringValue(s)
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
	Query *QueryType
	// MultiFrame enables multi frame requests for TOARSSI queries.
	MultiFrame *bool
	// MultiFrameWindowSize represents the number of historical frames to consider for the query.
	MultiFrameWindowSize *int
	// MultiFrameWindowAge limits the maximum age of the historical frames considered for the query.
	MultiFrameWindowAge *time.Duration
	// ServerURL represents the remote server to which the GLS queries are sent.
	ServerURL *url.URL
	// Token is the API token to be used when comunicating with the GLS server.
	Token *string
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

	return structpb.NewStringValue(base64Bytes.String()), nil
}

// GetMultiFrameWindowAge returns the value of the MultiFrameWindowAge field.
func (d *Data) GetMultiFrameWindowAge() time.Duration {
	if d.MultiFrameWindowAge == nil {
		return 0
	}
	return *d.MultiFrameWindowAge
}

// GetMultiFrameWindowSize returns the value of the MultiFrameWindowSize field.
func (d *Data) GetMultiFrameWindowSize() int {
	if d.MultiFrameWindowSize == nil {
		return 0
	}
	return *d.MultiFrameWindowSize
}

// GetMultiFrame returns the value of the MultiFrame field.
func (d *Data) GetMultiFrame() bool {
	if d.MultiFrame == nil {
		return false
	}
	return *d.MultiFrame
}

// Struct serializes the configuration to *structpb.Struct.
func (d *Data) Struct() (*structpb.Struct, error) {
	fields := map[string]*structpb.Value{}

	if d.Token != nil && *d.Token != "" {
		fields[tokenField] = structpb.NewStringValue(*d.Token)
	}

	if d.Query != nil {
		fields[queryField] = d.Query.Value()
	}

	if d.MultiFrame != nil {
		fields[multiFrameField] = structpb.NewBoolValue(*d.MultiFrame)
	}

	if d.ServerURL != nil && d.ServerURL.String() != "" {
		fields[serverURLField] = structpb.NewStringValue(d.ServerURL.String())
	}

	if d.MultiFrame != nil {
		fields[multiFrameField] = structpb.NewBoolValue(*d.MultiFrame)
	}

	if d.MultiFrameWindowSize != nil {
		fields[multiFrameWindowSize] = structpb.NewNumberValue(float64(*d.MultiFrameWindowSize))
	}

	if d.MultiFrameWindowAge != nil {
		windowAge := d.MultiFrameWindowAge.Minutes()
		fields[multiFrameWindowAge] = structpb.NewNumberValue(windowAge)
	}

	if len(d.RecentMetadata) > 0 {
		recentMD, err := toRecentMD(d.RecentMetadata)
		if err != nil {
			return nil, err
		}
		fields[recentMDField] = recentMD
	}

	if len(fields) == 0 {
		return nil, ErrEmptyData.New()
	}

	return &structpb.Struct{
		Fields: fields,
	}, nil
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
	if value, ok := fields[queryField]; ok {
		query := new(QueryType)
		if err := query.FromValue(value); err != nil {
			return err
		}
		d.Query = query
	}
	if value, ok := fields[multiFrameField]; ok {
		multiFrame, err := boolFromValue(value)
		if err != nil {
			return err
		}
		d.MultiFrame = &multiFrame
	}
	if value, ok := fields[multiFrameWindowSize]; ok {
		floatValue, err := float64FromValue(value)
		if err != nil {
			return err
		}
		windowSize := int(floatValue)
		d.MultiFrameWindowSize = &windowSize
	}
	if value, ok := fields[multiFrameWindowAge]; ok {
		windowAge, err := float64FromValue(value)
		if err != nil {
			return err
		}
		windowAgeDuration := time.Duration(windowAge) * time.Minute
		d.MultiFrameWindowAge = &windowAgeDuration
	}
	if value, ok := fields[tokenField]; ok {
		token, err := stringFromValue(value)
		if err != nil {
			return err
		}
		d.Token = &token
	}
	if value, ok := fields[serverURLField]; ok {
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
	if value, ok := fields[recentMDField]; ok {
		uplinks, err := recentMDFromValue(value)
		if err != nil {
			return err
		}
		d.RecentMetadata = uplinks
	}
	return nil
}

func mergeData(defaultData, associationData Data) *Data {
	merged := Data{
		Query:                defaultData.Query,
		MultiFrame:           defaultData.MultiFrame,
		MultiFrameWindowSize: defaultData.MultiFrameWindowSize,
		MultiFrameWindowAge:  defaultData.MultiFrameWindowAge,
		ServerURL:            defaultData.ServerURL,
		Token:                defaultData.Token,
		RecentMetadata:       defaultData.RecentMetadata,
	}

	if associationData.Query != nil {
		merged.Query = associationData.Query
	}
	if associationData.MultiFrame != nil {
		merged.MultiFrame = associationData.MultiFrame
	}
	if associationData.MultiFrameWindowSize != nil {
		merged.MultiFrameWindowSize = associationData.MultiFrameWindowSize
	}
	if associationData.MultiFrameWindowAge != nil {
		merged.MultiFrameWindowAge = associationData.MultiFrameWindowAge
	}
	if associationData.ServerURL != nil {
		merged.ServerURL = urlutil.CloneURL(associationData.ServerURL)
	}
	if associationData.Token != nil {
		merged.Token = associationData.Token
	}
	if len(associationData.RecentMetadata) > 0 {
		merged.RecentMetadata = associationData.RecentMetadata
	}
	if merged.ServerURL == nil {
		merged.ServerURL = urlutil.CloneURL(api.DefaultServerURL)
	}
	return &merged
}

// validateData validates the package configuration.
func validateData(data *Data) error {
	if data.Query == nil {
		return errFieldRequired.WithAttributes("field", queryField)
	}
	if data.ServerURL == nil {
		return errFieldRequired.WithAttributes("field", serverURLField)
	}
	if data.Token == nil {
		return errFieldRequired.WithAttributes("field", tokenField)
	}
	return nil
}
