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

package bleve_test

import (
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func brandsResponse(brandIDs ...string) *store.GetBrandsResponse {
	if brandIDs == nil {
		return &store.GetBrandsResponse{Brands: []*ttnpb.EndDeviceBrand{}}
	}
	brands := make([]*ttnpb.EndDeviceBrand, 0, len(brandIDs))
	for _, brandID := range brandIDs {
		brands = append(brands, &ttnpb.EndDeviceBrand{BrandId: brandID})
	}
	return &store.GetBrandsResponse{
		Count:  uint32(len(brandIDs)),
		Total:  uint32(len(brandIDs)),
		Offset: 0,
		Brands: brands,
	}
}

func modelsResponse(modelIDs ...string) *store.GetModelsResponse {
	if modelIDs == nil {
		return &store.GetModelsResponse{Models: []*ttnpb.EndDeviceModel{}}
	}
	models := make([]*ttnpb.EndDeviceModel, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, &ttnpb.EndDeviceModel{ModelId: modelID})
	}
	return &store.GetModelsResponse{
		Count:  uint32(len(modelIDs)),
		Total:  uint32(len(modelIDs)),
		Offset: 0,
		Models: models,
	}
}
