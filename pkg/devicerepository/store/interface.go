// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetBrandsRequest is a request to list available end device vendors, with pagination and sorting.
type GetBrandsRequest struct {
	BrandID string
	Limit,
	Page uint32
	OrderBy string
	Paths   []string
	Search  string
}

// GetBrandsResponse is a list of brands, along with pagination information.
type GetBrandsResponse struct {
	Count,
	Offset,
	Total uint32
	Brands []*ttnpb.EndDeviceBrand
}

// GetModelsRequest is a request to list available end device model definitions.
type GetModelsRequest struct {
	BrandID,
	ModelID string
	Limit,
	Page uint32
	OrderBy string
	Paths   []string
	Search  string
}

// GetEndDeviceProfilesRequest is a request to list available LoRaWAN end device profile definitions.
type GetEndDeviceProfilesRequest struct {
	BrandID,
	ModelID string
	Limit,
	Page uint32
	OrderBy string
	Paths   []string
	Search  string
}

// GetEndDeviceProfilesResponse returns available LoRaWAN end device profile definitions.
type GetEndDeviceProfilesResponse struct {
	Count,
	Offset,
	Total uint32
	Profiles []*EndDeviceProfile
}

// GetModelsResponse is a list of models, along with model information
type GetModelsResponse struct {
	Count,
	Offset,
	Total uint32
	Models []*ttnpb.EndDeviceModel
}

// GetCodecRequest is a request to retrieve the codec of
type GetCodecRequest interface {
	GetVersionIds() *ttnpb.EndDeviceVersionIdentifiers
	GetFieldMask() *fieldmaskpb.FieldMask
}

// Store contains end device definitions.
type Store interface {
	// GetBrands lists available end device vendors.
	GetBrands(GetBrandsRequest) (*GetBrandsResponse, error)
	// GetModels lists available end device definitions.
	GetModels(GetModelsRequest) (*GetModelsResponse, error)
	// GetEndDeviceProfiles lists available LoRaWAN end device profile definitions.
	GetEndDeviceProfiles(GetEndDeviceProfilesRequest) (*GetEndDeviceProfilesResponse, error)
	// GetTemplate retrieves an end device template for an end device definition.
	GetTemplate(*ttnpb.GetTemplateRequest, *EndDeviceProfile) (*ttnpb.EndDeviceTemplate, error)
	// GetUplinkDecoder retrieves the codec for decoding uplink messages.
	GetUplinkDecoder(GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error)
	// GetDownlinkDecoder retrieves the codec for decoding downlink messages.
	GetDownlinkDecoder(GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error)
	// GetDownlinkEncoder retrieves the codec for encoding downlink messages.
	GetDownlinkEncoder(GetCodecRequest) (*ttnpb.MessagePayloadEncoder, error)
	// Close closes the store.
	Close() error
}
