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

package remote

import (
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"gopkg.in/yaml.v2"
)

// remoteStore implements the Store interface using a fetcher.
type remoteStore struct {
	fetcher fetch.Interface
}

// NewRemoteStore initializes a new Store using a fetcher. Avoid using directly,
// use bleve.NewIndexedStore instead. Searching and ordering are not supported,
// and some operations can be very slow.
func NewRemoteStore(fetcher fetch.Interface) store.Store {
	return &remoteStore{fetcher}
}

// paginate returns page start and end indices, and false if the page is invalid.
func paginate(size int, limit, page uint32) (uint32, uint32) {
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * limit
	start, end := offset, uint32(size)
	if start >= end {
		return 0, 0
	}
	if limit > 0 && start+limit < end {
		end = start + limit
	}
	return start, end
}

// GetBrands gets available end device vendors from the vendor/index.yaml file.
func (s *remoteStore) GetBrands(req store.GetBrandsRequest) (*store.GetBrandsResponse, error) {
	b, err := s.fetcher.File("vendor", "index.yaml")
	if err != nil {
		return nil, err
	}
	rawVendors := VendorsIndex{}
	if err := yaml.Unmarshal(b, &rawVendors); err != nil {
		return nil, err
	}

	brands := make([]*ttnpb.EndDeviceBrand, 0, len(rawVendors.Vendors))
	for _, vendor := range rawVendors.Vendors {
		// Skip draft vendors
		if vendor.Draft {
			continue
		}
		pb, err := vendor.ToPB(req.Paths...)
		if err != nil {
			return nil, err
		}
		brands = append(brands, pb)
	}

	start, end := paginate(len(brands), req.Limit, req.Page)
	return &store.GetBrandsResponse{
		Count:  end - start,
		Offset: start,
		Total:  uint32(len(brands)),
		Brands: brands[start:end],
	}, nil
}

var (
	errBrandNotFound = errors.DefineNotFound("brand_not_found", "brand `{brand_id}` not found")
)

// listModelsByBrand gets available end device models by a single brand.
func (s *remoteStore) listModelsByBrand(req store.GetModelsRequest) (*store.GetModelsResponse, error) {
	b, err := s.fetcher.File("vendor", req.BrandID, "index.yaml")
	if err != nil {
		return nil, errBrandNotFound.WithAttributes("brand_id", req.BrandID)
	}
	index := VendorEndDevicesIndex{}
	if err := yaml.Unmarshal(b, &index); err != nil {
		return nil, err
	}
	start, end := paginate(len(index.EndDevices), req.Limit, req.Page)

	models := make([]*ttnpb.EndDeviceModel, 0, end-start)
	for idx := start; idx < end; idx++ {
		modelID := index.EndDevices[idx]
		if req.ModelID != "" && modelID != req.ModelID {
			continue
		}
		b, err := s.fetcher.File("vendor", req.BrandID, modelID+".yaml")
		if err != nil {
			return nil, err
		}
		model := EndDeviceModel{}
		if err := yaml.Unmarshal(b, &model); err != nil {
			return nil, err
		}
		pb, err := model.ToPB(req.BrandID, modelID, req.Paths...)
		if err != nil {
			return nil, err
		}
		models = append(models, pb)
	}
	return &store.GetModelsResponse{
		Count:  end - start,
		Offset: start,
		Total:  uint32(len(index.EndDevices)),
		Models: models,
	}, nil
}

// GetModels gets available end device models. Note that this can be very slow, and does not support searching/sorting.
func (s *remoteStore) GetModels(req store.GetModelsRequest) (*store.GetModelsResponse, error) {
	if req.BrandID != "" {
		return s.listModelsByBrand(req)
	}
	all := []*ttnpb.EndDeviceModel{}
	brands, err := s.GetBrands(store.GetBrandsRequest{
		Paths: []string{"brand_id"},
	})
	if err != nil {
		return nil, err
	}
	for _, brand := range brands.Brands {
		models, err := s.GetModels(store.GetModelsRequest{
			Paths:   req.Paths,
			BrandID: brand.BrandID,
			Limit:   req.Limit,
		})
		if errors.IsNotFound(err) {
			// Skip vendors without any models
			continue
		} else if err != nil {
			return nil, err
		}
		all = append(all, models.Models...)
	}

	start, end := paginate(len(all), req.Limit, req.Page)
	return &store.GetModelsResponse{
		Count:  end - start,
		Offset: start,
		Total:  uint32(len(all)),
		Models: all[start:end],
	}, nil
}

var (
	errModelNotFound           = errors.DefineNotFound("model_not_found", "model `{brand_id}/{model_id}` not found")
	errBandNotFound            = errors.DefineNotFound("band_not_found", "band `{band_id}` not found")
	errNoProfileForBand        = errors.DefineNotFound("no_profile_for_band", "device does not have a profile for band `{band_id}`")
	errFirmwareVersionNotFound = errors.DefineNotFound("firmware_version_not_found", "firmware version `{firmware_version}` for model `{brand_id}/{model_id}` not found")
)

// GetTemplate retrieves an end device template for an end device definition.
func (s *remoteStore) GetTemplate(ids *ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.EndDeviceTemplate, error) {
	models, err := s.GetModels(store.GetModelsRequest{
		BrandID: ids.BrandID,
		ModelID: ids.ModelID,
		Paths: []string{
			"firmware_versions",
		},
	})
	if err != nil {
		return nil, err
	}
	if len(models.Models) == 0 {
		return nil, errModelNotFound.WithAttributes("brand_id", ids.BrandID, "model_id", ids.ModelID)
	}
	model := models.Models[0]
	for _, ver := range model.FirmwareVersions {
		if ver.Version != ids.FirmwareVersion {
			continue
		}

		if _, ok := bandIDToRegion[ids.BandID]; !ok {
			return nil, errBandNotFound.WithAttributes("unknown_band", ids.BandID)
		}
		profileInfo, ok := ver.Profiles[ids.BandID]
		if !ok {
			return nil, errNoProfileForBand.WithAttributes(
				"band_id", ids.BandID,
			)
		}

		b, err := s.fetcher.File("vendor", ids.BrandID, profileInfo.ProfileID+".yaml")
		if err != nil {
			return nil, err
		}
		profile := EndDeviceProfile{}
		if err := yaml.Unmarshal(b, &profile); err != nil {
			return nil, err
		}

		return profile.ToTemplatePB(ids, profileInfo)
	}
	return nil, errFirmwareVersionNotFound.WithAttributes(
		"brand_id", ids.BrandID,
		"model_id", ids.ModelID,
		"firmware_version", ids.FirmwareVersion,
	)
}

var (
	errNoCodec = errors.DefineNotFound("no_codec", "no codec defined for firmware version `{firmware_version}` and band `{band_id}`")
)

// getCodec retrieves codec information for a specific model and returns.
func (s *remoteStore) getCodec(ids *ttnpb.EndDeviceVersionIdentifiers, chooseFile func(EndDeviceCodec) string) (*ttnpb.MessagePayloadFormatter, error) {
	models, err := s.GetModels(store.GetModelsRequest{
		BrandID: ids.BrandID,
		ModelID: ids.ModelID,
		Paths: []string{
			"firmware_versions",
		},
	})
	if err != nil {
		return nil, err
	}
	if len(models.Models) == 0 {
		return nil, errModelNotFound.WithAttributes("brand_id", ids.BrandID, "model_id", ids.ModelID)
	}
	model := models.Models[0]
	for _, ver := range model.FirmwareVersions {
		if ver.Version != ids.FirmwareVersion {
			continue
		}

		if _, ok := bandIDToRegion[ids.BandID]; !ok {
			return nil, errBandNotFound.WithAttributes("unknown_band", ids.BandID)
		}
		profileInfo, ok := ver.Profiles[ids.BandID]
		if !ok {
			return nil, errNoProfileForBand.WithAttributes(
				"band_id", ids.BandID,
			)
		}

		if profileInfo.CodecID == "" {
			return nil, errNoCodec.WithAttributes("firmware_version", ids.FirmwareVersion, "band_id", ids.BandID)
		}

		codec := EndDeviceCodec{}
		b, err := s.fetcher.File("vendor", ids.BrandID, profileInfo.CodecID+".yaml")
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(b, &codec); err != nil {
			return nil, err
		}
		if file := chooseFile(codec); file != "" {
			b, err := s.fetcher.File("vendor", ids.BrandID, file)
			if err != nil {
				return nil, err
			}
			return &ttnpb.MessagePayloadFormatter{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: string(b),
			}, nil
		}
	}

	return nil, errFirmwareVersionNotFound.WithAttributes(
		"brand_id", ids.BrandID,
		"model_id", ids.ModelID,
		"firmware_version", ids.FirmwareVersion,
	)
}

// GetUplinkDecoder retrieves the codec for decoding uplink messages.
func (s *remoteStore) GetUplinkDecoder(ids *ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.MessagePayloadFormatter, error) {
	return s.getCodec(ids, func(c EndDeviceCodec) string { return c.UplinkDecoder.FileName })
}

// GetDownlinkDecoder retrieves the codec for decoding downlink messages.
func (s *remoteStore) GetDownlinkDecoder(ids *ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.MessagePayloadFormatter, error) {
	return s.getCodec(ids, func(c EndDeviceCodec) string { return c.DownlinkDecoder.FileName })
}

// GetDownlinkEncoder retrieves the codec for encoding downlink messages.
func (s *remoteStore) GetDownlinkEncoder(ids *ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.MessagePayloadFormatter, error) {
	return s.getCodec(ids, func(c EndDeviceCodec) string { return c.DownlinkEncoder.FileName })
}

// Close closes the store.
func (s *remoteStore) Close() error {
	return nil
}
