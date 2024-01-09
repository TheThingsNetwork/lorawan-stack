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

package devicerepository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// withDefaultModelFields appends default ttnpb.EndDeviceModel fields.
func withDefaultModelFields(paths []string) []string {
	return ttnpb.AddFields(paths, "brand_id", "model_id")
}

// withDefaultBrandFields appends default ttnpb.EndDeviceBrand fields.
func withDefaultBrandFields(paths []string) []string {
	return ttnpb.AddFields(paths, "brand_id")
}

func (dr *DeviceRepository) assetURL(brandID, path string) string {
	if path == "" || dr.config.AssetsBaseURL == "" || strings.HasPrefix(path, dr.config.AssetsBaseURL) {
		return path
	}
	return fmt.Sprintf("%s/vendor/%s/%s", strings.TrimRight(dr.config.AssetsBaseURL, "/"), brandID, path)
}

// ensureBaseAssetURLs prepends the BaseAssetURL to model assets.
func (dr *DeviceRepository) ensureBaseAssetURLs(models []*ttnpb.EndDeviceModel) {
	for _, model := range models {
		if photos := model.Photos; photos != nil {
			photos.Main = dr.assetURL(model.BrandId, photos.Main)
			for idx, photo := range photos.Other {
				photos.Other[idx] = dr.assetURL(model.BrandId, photo)
			}
		}
	}
}

const defaultLimit = 1000

// ListBrands implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) ListBrands(
	ctx context.Context,
	req *ttnpb.ListEndDeviceBrandsRequest,
) (*ttnpb.ListEndDeviceBrandsResponse, error) {
	if err := rights.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	if req.Limit > defaultLimit || req.Limit == 0 {
		req.Limit = defaultLimit
	}
	response, err := dr.store.GetBrands(store.GetBrandsRequest{
		Limit:   req.Limit,
		Page:    req.Page,
		OrderBy: req.OrderBy,
		Paths:   withDefaultBrandFields(req.FieldMask.GetPaths()),
		Search:  req.Search,
	})
	if err != nil {
		return nil, err
	}
	for _, brand := range response.Brands {
		brand.Logo = dr.assetURL(brand.BrandId, brand.Logo)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(uint64(response.Total), 10)))
	return &ttnpb.ListEndDeviceBrandsResponse{
		Brands: response.Brands,
	}, nil
}

var errBrandNotFound = errors.DefineNotFound("brand_not_found", "brand `{brand_id}` not found")

// GetBrand implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) GetBrand(
	ctx context.Context,
	req *ttnpb.GetEndDeviceBrandRequest,
) (*ttnpb.EndDeviceBrand, error) {
	if err := rights.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	response, err := dr.store.GetBrands(store.GetBrandsRequest{
		BrandID: req.BrandId,
		Paths:   withDefaultBrandFields(req.FieldMask.GetPaths()),
		Limit:   1,
	})
	if err != nil {
		return nil, err
	}
	if len(response.Brands) == 0 {
		return nil, errBrandNotFound.WithAttributes("brand_id", req.BrandId)
	}
	brand := response.Brands[0]
	brand.Logo = dr.assetURL(brand.BrandId, brand.Logo)
	return brand, nil
}

// ListModels implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) ListModels(
	ctx context.Context,
	req *ttnpb.ListEndDeviceModelsRequest,
) (*ttnpb.ListEndDeviceModelsResponse, error) {
	if err := rights.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	if req.Limit > defaultLimit || req.Limit == 0 {
		req.Limit = defaultLimit
	}
	response, err := dr.store.GetModels(store.GetModelsRequest{
		BrandID: req.BrandId,
		Limit:   req.Limit,
		Page:    req.Page,
		Paths:   withDefaultModelFields(req.FieldMask.GetPaths()),
		Search:  req.Search,
		OrderBy: req.OrderBy,
	})
	if err != nil {
		return nil, err
	}
	dr.ensureBaseAssetURLs(response.Models)
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(uint64(response.Total), 10)))
	return &ttnpb.ListEndDeviceModelsResponse{
		Models: response.Models,
	}, nil
}

var errModelNotFound = errors.DefineNotFound("model_not_found", "model `{brand_id}/{model_id}` not found")

// GetModel implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) GetModel(
	ctx context.Context,
	req *ttnpb.GetEndDeviceModelRequest,
) (*ttnpb.EndDeviceModel, error) {
	if err := rights.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	response, err := dr.store.GetModels(store.GetModelsRequest{
		BrandID: req.BrandId,
		ModelID: req.ModelId,
		Limit:   1,
		Paths:   withDefaultModelFields(req.FieldMask.GetPaths()),
	})
	if err != nil {
		return nil, err
	}
	if len(response.Models) == 0 {
		return nil, errModelNotFound.WithAttributes("brand_id", req.BrandId, "model_id", req.ModelId)
	}
	dr.ensureBaseAssetURLs(response.Models)

	return response.Models[0], nil
}

// GetTemplate implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) GetTemplate(
	ctx context.Context,
	req *ttnpb.GetTemplateRequest,
) (*ttnpb.EndDeviceTemplate, error) {
	if err := rights.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return dr.store.GetTemplate(req, nil)
}

func getDecoder(
	ctx context.Context,
	req *ttnpb.GetPayloadFormatterRequest,
	f func(store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error),
) (*ttnpb.MessagePayloadDecoder, error) {
	if clusterauth.Authorized(ctx) != nil {
		if err := rights.RequireAuthenticated(ctx); err != nil {
			return nil, err
		}
	}
	return f(req)
}

// GetUplinkDecoder implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) GetUplinkDecoder(
	ctx context.Context,
	req *ttnpb.GetPayloadFormatterRequest,
) (*ttnpb.MessagePayloadDecoder, error) {
	return getDecoder(ctx, req, dr.store.GetUplinkDecoder)
}

// GetDownlinkDecoder implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) GetDownlinkDecoder(
	ctx context.Context,
	req *ttnpb.GetPayloadFormatterRequest,
) (*ttnpb.MessagePayloadDecoder, error) {
	return getDecoder(ctx, req, dr.store.GetDownlinkDecoder)
}

// GetDownlinkEncoder implements the ttnpb.DeviceRepositoryServer interface.
func (dr *DeviceRepository) GetDownlinkEncoder(
	ctx context.Context,
	req *ttnpb.GetPayloadFormatterRequest,
) (*ttnpb.MessagePayloadEncoder, error) {
	if clusterauth.Authorized(ctx) != nil {
		if err := rights.RequireAuthenticated(ctx); err != nil {
			return nil, err
		}
	}
	return dr.store.GetDownlinkEncoder(req)
}
