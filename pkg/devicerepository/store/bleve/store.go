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

package bleve

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errCorruptedIndex = errors.DefineCorruption("corrupted_index", "corrupted index file")
)

// retrieve returns the resulting document from the cache, if available. Otherwise,
// it extracts it from the appropriate field of the document match result and stores
// in the cache for future use.
func (s *bleveStore) retrieve(hit *search.DocumentMatch, fieldName string, newValue func() interface{}) (interface{}, error) {
	cached, err := s.cache.Get(hit.ID)
	if err != nil {
		jsonString, ok := hit.Fields[fieldName].(string)
		if !ok {
			return nil, errCorruptedIndex.New()
		}
		v := newValue()
		if err := jsonpb.TTN().Unmarshal([]byte(jsonString), v); err != nil {
			return nil, err
		}
		if err := s.cache.Set(hit.ID, v); err != nil {
			log.FromContext(s.ctx).WithError(err).WithField("key", hit.ID).Warn("Failed to cache")
		}
		return v, nil
	}
	return cached, nil
}

// GetBrands lists available end device vendors.
func (s *bleveStore) GetBrands(req store.GetBrandsRequest) (*store.GetBrandsResponse, error) {
	documentTypeQuery := bleve.NewTermQuery(brandDocumentType)
	documentTypeQuery.SetField("Type")
	queries := []query.Query{documentTypeQuery}

	if q := req.Search; q != "" {
		queries = append(queries, bleve.NewQueryStringQuery(q))
	}
	if q := req.BrandID; q != "" {
		query := bleve.NewTermQuery(q)
		query.SetField("BrandID")
		queries = append(queries, query)
	}

	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(queries...))
	if req.Limit > 0 {
		searchRequest.Size = int(req.Limit)
	}
	if req.Page == 0 {
		req.Page = 1
	}
	searchRequest.From = int((req.Page - 1) * req.Limit)

	searchRequest.Fields = []string{"BrandJSON"}
	switch req.OrderBy {
	case "brand_id":
		searchRequest.SortBy([]string{"BrandID"})
	case "-brand_id":
		searchRequest.SortBy([]string{"-BrandID"})
	case "name":
		searchRequest.SortBy([]string{"BrandName"})
	case "-name":
		searchRequest.SortBy([]string{"-BrandName"})
	}

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	brands := make([]*ttnpb.EndDeviceBrand, 0, len(result.Hits))
	for _, hit := range result.Hits {
		brand, err := s.retrieve(hit, "BrandJSON", func() interface{} { return &ttnpb.EndDeviceBrand{} })
		if err != nil {
			return nil, err
		}
		pb := &ttnpb.EndDeviceBrand{}
		if err := pb.SetFields(brand.(*ttnpb.EndDeviceBrand), req.Paths...); err != nil {
			return nil, err
		}
		brands = append(brands, pb)
	}
	return &store.GetBrandsResponse{
		Count:  uint32(len(result.Hits)),
		Total:  uint32(result.Total),
		Offset: uint32(searchRequest.From),
		Brands: brands,
	}, nil
}

// GetModels lists available end device definitions.
func (s *bleveStore) GetModels(req store.GetModelsRequest) (*store.GetModelsResponse, error) {
	documentTypeQuery := bleve.NewTermQuery(modelDocumentType)
	documentTypeQuery.SetField("Type")
	queries := []query.Query{documentTypeQuery}

	if q := req.Search; q != "" {
		queries = append(queries, bleve.NewQueryStringQuery(q))
	}
	if q := req.BrandID; q != "" {
		query := bleve.NewTermQuery(q)
		query.SetField("BrandID")
		queries = append(queries, query)
	}
	if q := req.ModelID; q != "" {
		query := bleve.NewTermQuery(q)
		query.SetField("ModelID")
		queries = append(queries, query)
	}

	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(queries...))
	if req.Limit > 0 {
		searchRequest.Size = int(req.Limit)
	}
	if req.Page == 0 {
		req.Page = 1
	}
	searchRequest.From = int((req.Page - 1) * req.Limit)
	searchRequest.Fields = []string{"ModelJSON"}

	switch req.OrderBy {
	case "brand_id":
		searchRequest.SortBy([]string{"BrandID"})
	case "-brand_id":
		searchRequest.SortBy([]string{"-BrandID"})
	case "model_id":
		searchRequest.SortBy([]string{"ModelID"})
	case "-model_id":
		searchRequest.SortBy([]string{"-ModelID"})
	}

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	models := make([]*ttnpb.EndDeviceModel, 0, len(result.Hits))
	for _, hit := range result.Hits {
		model, err := s.retrieve(hit, "ModelJSON", func() interface{} { return &ttnpb.EndDeviceModel{} })
		if err != nil {
			return nil, err
		}
		pb := &ttnpb.EndDeviceModel{}
		if err := pb.SetFields(model.(*ttnpb.EndDeviceModel), req.Paths...); err != nil {
			return nil, err
		}
		models = append(models, pb)
	}
	return &store.GetModelsResponse{
		Count:  uint32(len(result.Hits)),
		Total:  uint32(result.Total),
		Offset: uint32(searchRequest.From),
		Models: models,
	}, nil
}

// GetTemplate retrieves an end device template for an end device definition.
func (s *bleveStore) GetTemplate(ids *ttnpb.EndDeviceVersionIdentifiers) (*ttnpb.EndDeviceTemplate, error) {
	return s.store.GetTemplate(ids)
}

// GetUplinkDecoder retrieves the codec for decoding uplink messages.
func (s *bleveStore) GetUplinkDecoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	return s.store.GetUplinkDecoder(req)
}

// GetDownlinkDecoder retrieves the codec for decoding downlink messages.
func (s *bleveStore) GetDownlinkDecoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	return s.store.GetDownlinkDecoder(req)
}

// GetDownlinkEncoder retrieves the codec for encoding downlink messages.
func (s *bleveStore) GetDownlinkEncoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadEncoder, error) {
	return s.store.GetDownlinkEncoder(req)
}

// Close closes the store.
func (s *bleveStore) Close() error {
	return s.index.Close()
}
