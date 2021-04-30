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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// NoopStore is a no-op.
type NoopStore struct{}

var (
	errNotFound = errors.DefineNotFound("not_found", "not found")
)

// GetBrands gets available end device vendors.
func (*NoopStore) GetBrands(GetBrandsRequest) (*GetBrandsResponse, error) {
	return nil, errNotFound.New()
}

// GetModels gets available end device definitions.
func (*NoopStore) GetModels(GetModelsRequest) (*GetModelsResponse, error) {
	return nil, errNotFound.New()
}

// GetTemplate retrieves an end device template for an end device definition.
func (*NoopStore) GetTemplate(*ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.EndDeviceTemplate, error) {
	return nil, errNotFound.New()
}

// GetUplinkDecoder retrieves the codec for decoding uplink messages.
func (*NoopStore) GetUplinkDecoder(GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	return nil, errNotFound.New()
}

// GetDownlinkDecoder retrieves the codec for decoding downlink messages.
func (*NoopStore) GetDownlinkDecoder(GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	return nil, errNotFound.New()
}

// GetDownlinkEncoder retrieves the codec for encoding downlink messages.
func (*NoopStore) GetDownlinkEncoder(GetCodecRequest) (*ttnpb.MessagePayloadEncoder, error) {
	return nil, errNotFound.New()
}
