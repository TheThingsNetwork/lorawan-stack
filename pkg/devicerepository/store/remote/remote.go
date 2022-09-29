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
	"go.thethings.network/lorawan-stack/v3/pkg/goproto"
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

var errBrandNotFound = errors.DefineNotFound("brand_not_found", "brand `{brand_id}` not found")

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
			BrandID: brand.BrandId,
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

// getEndDeviceProfilesByBrand lists the available LoRaWAN end device profiles by a single brand.
func (s *remoteStore) getEndDeviceProfilesByBrand(req store.GetEndDeviceProfilesRequest) (*store.GetEndDeviceProfilesResponse, error) {
	b, err := s.fetcher.File("vendor", req.BrandID, "index.yaml")
	if err != nil {
		return nil, errBrandNotFound.WithAttributes("brand_id", req.BrandID)
	}
	index := VendorEndDevicesIndex{}
	if err := yaml.Unmarshal(b, &index); err != nil {
		return nil, err
	}
	start, end := paginate(len(index.EndDevices), req.Limit, req.Page)

	profiles := make([]*store.EndDeviceProfile, 0, end-start)

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

		// For each profile ID, get the profile.
		for _, fwVersion := range model.FirmwareVersions {
			for _, profile := range fwVersion.Profiles {
				p, err := s.fetcher.File("vendor", req.BrandID, profile.ID+".yaml")
				if err != nil {
					return nil, err
				}
				profile := store.EndDeviceProfile{}
				if err := yaml.Unmarshal(p, &profile); err != nil {
					return nil, err
				}
				profiles = append(profiles, &profile)
			}
		}
	}
	return &store.GetEndDeviceProfilesResponse{
		Count:    end - start,
		Offset:   start,
		Total:    uint32(len(index.EndDevices)),
		Profiles: profiles,
	}, nil
}

// GetEndDeviceProfiles lists available LoRaWAN end device profiles per brand.
// Note that this can be very slow, and does not support searching/sorting.
// This function is primarily intended to be used for creating the bleve index.
func (s *remoteStore) GetEndDeviceProfiles(req store.GetEndDeviceProfilesRequest) (*store.GetEndDeviceProfilesResponse, error) {
	if req.BrandID != "" {
		return s.getEndDeviceProfilesByBrand(req)
	}
	all := []*store.EndDeviceProfile{}
	brands, err := s.GetBrands(store.GetBrandsRequest{
		Paths: []string{"brand_id"},
	})
	if err != nil {
		return nil, err
	}
	for _, brand := range brands.Brands {
		profiles, err := s.GetEndDeviceProfiles(store.GetEndDeviceProfilesRequest{
			BrandID: brand.BrandId,
		})
		if errors.IsNotFound(err) {
			// Skip vendors without any profiles.
			continue
		} else if err != nil {
			return nil, err
		}
		all = append(all, profiles.Profiles...)
	}

	start, end := paginate(len(all), req.Limit, req.Page)
	return &store.GetEndDeviceProfilesResponse{
		Count:    end - start,
		Offset:   start,
		Total:    uint32(len(all)),
		Profiles: all[start:end],
	}, nil
}

var (
	errModelNotFound           = errors.DefineNotFound("model_not_found", "model `{brand_id}/{model_id}` not found")
	errBandNotFound            = errors.DefineNotFound("band_not_found", "band `{band_id}` not found")
	errNoProfileForBand        = errors.DefineNotFound("no_profile_for_band", "device does not have a profile for band `{band_id}`")
	errFirmwareVersionNotFound = errors.DefineNotFound("firmware_version_not_found", "firmware version `{firmware_version}` for model `{brand_id}/{model_id}` not found")
)

// GetTemplate retrieves an end device template for an end device definition.
func (s *remoteStore) GetTemplate(req *ttnpb.GetTemplateRequest, profile *store.EndDeviceProfile) (*ttnpb.EndDeviceTemplate, error) {
	ids := req.GetVersionIds()
	if profile != nil {
		return profile.ToTemplatePB(ids, nil)
	}

	// Parse the models and return the End Device Profile that corresponds to the Band ID.
	models, err := s.GetModels(store.GetModelsRequest{
		BrandID: ids.BrandId,
		ModelID: ids.ModelId,
		Paths: []string{
			"firmware_versions",
		},
	})
	if err != nil {
		return nil, err
	}
	if len(models.Models) == 0 {
		return nil, errModelNotFound.WithAttributes("brand_id", ids.BrandId, "model_id", ids.ModelId)
	}
	model := models.Models[0]
	for _, ver := range model.FirmwareVersions {
		if ver.Version != ids.FirmwareVersion {
			continue
		}

		if _, ok := bandIDToRegion[ids.BandId]; !ok {
			return nil, errBandNotFound.WithAttributes("unknown_band", ids.BandId)
		}
		profileInfo, ok := ver.Profiles[ids.BandId]
		if !ok {
			return nil, errNoProfileForBand.WithAttributes(
				"band_id", ids.BandId,
			)
		}

		profileVendorID := ids.BrandId
		if id := profileInfo.VendorId; id != "" {
			profileVendorID = id
		}
		b, err := s.fetcher.File("vendor", profileVendorID, profileInfo.ProfileId+".yaml")
		if err != nil {
			return nil, err
		}
		profile := store.EndDeviceProfile{}
		if err := yaml.Unmarshal(b, &profile); err != nil {
			return nil, err
		}

		return profile.ToTemplatePB(ids, profileInfo)
	}
	return nil, errFirmwareVersionNotFound.WithAttributes(
		"brand_id", ids.BrandId,
		"model_id", ids.ModelId,
		"firmware_version", ids.FirmwareVersion,
	)
}

var errNoCodec = errors.DefineNotFound("no_codec", "no codec defined for firmware version `{firmware_version}` and band `{band_id}`")

func (s *remoteStore) getCodecs(ids *ttnpb.EndDeviceVersionIdentifiers) (*EndDeviceCodecs, error) {
	models, err := s.GetModels(store.GetModelsRequest{
		BrandID: ids.BrandId,
		ModelID: ids.ModelId,
		Paths: []string{
			"firmware_versions",
		},
	})
	if err != nil {
		return nil, err
	}
	if len(models.Models) == 0 {
		return nil, errModelNotFound.WithAttributes("brand_id", ids.BrandId, "model_id", ids.ModelId)
	}
	model := models.Models[0]
	var version *ttnpb.EndDeviceModel_FirmwareVersion = nil
	for _, ver := range model.FirmwareVersions {
		if ver.Version == ids.FirmwareVersion {
			version = ver
			break
		}
	}

	if version == nil {
		return nil, errFirmwareVersionNotFound.WithAttributes(
			"brand_id", ids.BrandId,
			"model_id", ids.ModelId,
			"firmware_version", ids.FirmwareVersion,
		)
	}

	if _, ok := bandIDToRegion[ids.BandId]; !ok {
		return nil, errBandNotFound.WithAttributes("band_id", ids.BandId)
	}
	profileInfo, ok := version.Profiles[ids.BandId]
	if !ok {
		return nil, errNoProfileForBand.WithAttributes(
			"band_id", ids.BandId,
		)
	}
	if profileInfo.CodecId == "" {
		return nil, errNoCodec.WithAttributes("firmware_version", ids.FirmwareVersion, "band_id", ids.BandId)
	}

	codecs := &EndDeviceCodecs{
		CodecID: profileInfo.CodecId,
	}
	b, err := s.fetcher.File("vendor", ids.BrandId, codecs.CodecID+".yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, codecs); err != nil {
		return nil, err
	}
	return codecs, nil
}

var (
	errNoDecoder = errors.DefineNotFound("no_decoder", "no decoder defined for codec `{codec_id}`")
	errNoEncoder = errors.DefineNotFound("no_encoder", "no encoder defined for codec `{codec_id}`")
)

func (s *remoteStore) getDecoder(req store.GetCodecRequest, choose func(codecs *EndDeviceCodecs) *EndDeviceDecoderCodec) (*ttnpb.MessagePayloadDecoder, error) {
	codecs, err := s.getCodecs(req.GetVersionIds())
	if err != nil {
		return nil, err
	}
	codec := choose(codecs)
	if codec.FileName == "" {
		return nil, errNoDecoder.WithAttributes("codec_id", codecs.CodecID)
	}

	b, err := s.fetcher.File("vendor", req.GetVersionIds().BrandId, codec.FileName)
	if err != nil {
		return nil, err
	}

	paths := ttnpb.AddFields(req.GetFieldMask().GetPaths(), "formatter", "formatter_parameter")
	var examples []*ttnpb.MessagePayloadDecoder_Example
	if ttnpb.HasAnyField([]string{"examples"}, paths...) && len(codec.Examples) > 0 {
		examples = make([]*ttnpb.MessagePayloadDecoder_Example, 0, len(codec.Examples))
		for _, e := range codec.Examples {
			pb := &ttnpb.MessagePayloadDecoder_Example{
				Description: e.Description,
				Input: &ttnpb.EncodedMessagePayload{
					FPort:      e.Input.FPort,
					FrmPayload: e.Input.Bytes,
				},
				Output: &ttnpb.DecodedMessagePayload{
					Warnings: e.Output.Warnings,
					Errors:   e.Output.Errors,
				},
			}
			if pb.Output.Data, err = goproto.Struct(e.Output.Data); err != nil {
				return nil, err
			}
			examples = append(examples, pb)
		}
	}
	formatter := &ttnpb.MessagePayloadDecoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: string(b),
		Examples:           examples,
		CodecId:            codecs.CodecID,
	}
	pb := &ttnpb.MessagePayloadDecoder{}
	if err := pb.SetFields(formatter, paths...); err != nil {
		return nil, err
	}
	return pb, nil
}

// GetUplinkDecoder retrieves the codec for decoding uplink messages.
func (s *remoteStore) GetUplinkDecoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	return s.getDecoder(req, func(codecs *EndDeviceCodecs) *EndDeviceDecoderCodec { return &codecs.UplinkDecoder })
}

// GetDownlinkDecoder retrieves the codec for decoding downlink messages.
func (s *remoteStore) GetDownlinkDecoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	return s.getDecoder(req, func(codecs *EndDeviceCodecs) *EndDeviceDecoderCodec { return &codecs.DownlinkDecoder })
}

// GetDownlinkEncoder retrieves the codec for encoding downlink messages.
func (s *remoteStore) GetDownlinkEncoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadEncoder, error) {
	codecs, err := s.getCodecs(req.GetVersionIds())
	if err != nil {
		return nil, err
	}
	codec := codecs.DownlinkEncoder

	if codec.FileName == "" {
		return nil, errNoEncoder.WithAttributes("firmware_version", req.GetVersionIds().FirmwareVersion, "band_id", req.GetVersionIds().BandId)
	}

	b, err := s.fetcher.File("vendor", req.GetVersionIds().BrandId, codec.FileName)
	if err != nil {
		return nil, err
	}
	paths := ttnpb.AddFields(req.GetFieldMask().GetPaths(), "formatter", "formatter_parameter")
	var examples []*ttnpb.MessagePayloadEncoder_Example
	if ttnpb.HasAnyField([]string{"examples"}, paths...) && len(codec.Examples) > 0 {
		examples = make([]*ttnpb.MessagePayloadEncoder_Example, 0, len(codec.Examples))
		for _, e := range codec.Examples {
			pb := &ttnpb.MessagePayloadEncoder_Example{
				Description: e.Description,
				Input:       &ttnpb.DecodedMessagePayload{},
				Output: &ttnpb.EncodedMessagePayload{
					FPort:      e.Output.FPort,
					FrmPayload: e.Output.Bytes,
					Warnings:   e.Output.Warnings,
					Errors:     e.Output.Errors,
				},
			}
			if pb.Input.Data, err = goproto.Struct(e.Input.Data); err != nil {
				return nil, err
			}
			examples = append(examples, pb)
		}
	}
	formatter := &ttnpb.MessagePayloadEncoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: string(b),
		Examples:           examples,
		CodecId:            codecs.CodecID,
	}
	pb := &ttnpb.MessagePayloadEncoder{}
	if err := pb.SetFields(formatter, paths...); err != nil {
		return nil, err
	}
	return pb, nil
}

// Close closes the store.
func (s *remoteStore) Close() error {
	return nil
}
