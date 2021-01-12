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
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store/remote"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	indexPath         = "index.bleve"
	brandDocumentType = "brand"
	modelDocumentType = "model"
)

type indexableBrand struct {
	Brand  *ttnpb.EndDeviceBrand
	Models []*ttnpb.EndDeviceModel

	BrandJSON          string // *ttnpb.EndDeviceBrand marshaled into string
	BrandID, BrandName string // stored separately to support queries

	Type string // Index document type, always brandDocumentType
}

type indexableModel struct {
	Brand *ttnpb.EndDeviceBrand
	Model *ttnpb.EndDeviceModel

	ModelJSON        string // *ttnpb.EndDeviceModel marshaled into string.
	BrandID, ModelID string // stored separately to support queries.

	Type string // Index document type, always modelDocumentType
}

func newIndex(path string, overwrite bool, keywords ...string) (bleve.Index, error) {
	mapping := bleve.NewIndexMapping()
	if st, err := os.Stat(path); err == nil && st.IsDir() && overwrite {
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
	}
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name
	for _, keyword := range keywords {
		mapping.DefaultMapping.AddFieldMappingsAt(keyword, keywordFieldMapping)
	}
	return bleve.New(path, mapping)
}

func getWorkingDirectory(paths []string) (string, error) {
	for _, path := range paths {
		if s, err := os.Stat(path); err == nil && s.IsDir() {
			return path, nil
		}
	}

	// Default to working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, "device-repository"), nil
}

// packageFileName is the name of the Device Repository package.
const packageFileName = "master.zip"

// Initialize fetches the Device Repository package file and generates index files.
func (c Config) Initialize(ctx context.Context, fetcher fetch.Interface, overwrite bool) error {
	wd, err := getWorkingDirectory(c.SearchPaths)
	if err != nil {
		return err
	}

	b, err := fetcher.File(packageFileName)
	if err != nil {
		return err
	}

	if err := unarchive(b, wd, func(path string) (string, bool) {
		path = strings.TrimPrefix(path, "lorawan-devices-master/")
		if !strings.HasPrefix(path, "vendor/") || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".js")) {
			return "", true
		}
		return path, false
	}); err != nil {
		return err
	}
	s := remote.NewRemoteStore(fetch.FromFilesystem(wd))
	index, err := newIndex(path.Join(wd, indexPath), overwrite, "BrandID", "ModelID", "Type")
	if err != nil {
		return err
	}
	defer index.Close()

	brands, err := s.GetBrands(store.GetBrandsRequest{
		Paths: ttnpb.EndDeviceBrandFieldPathsNested,
	})
	if err != nil {
		return err
	}

	batch := index.NewBatch()
	for _, brand := range brands.Brands {
		models, err := s.GetModels(store.GetModelsRequest{
			Paths:   ttnpb.EndDeviceModelFieldPathsNested,
			BrandID: brand.BrandID,
		})
		if err != nil {
			if errors.IsNotFound(err) {
				// Skip vendors without any models
				continue
			} else {
				return err
			}
		}
		brandJSON, err := jsonpb.TTN().Marshal(brand)
		if err != nil {
			return err
		}
		if err := batch.Index(brand.BrandID, indexableBrand{
			Type:      brandDocumentType,
			BrandJSON: string(brandJSON),
			Brand:     brand,
			Models:    models.Models,
			BrandID:   brand.BrandID,
			BrandName: brand.Name,
		}); err != nil {
			return err
		}
		for _, model := range models.Models {
			modelJSON, err := jsonpb.TTN().Marshal(model)
			if err != nil {
				return err
			}
			if err := batch.Index(fmt.Sprintf("%s:%s", model.BrandID, model.ModelID), indexableModel{
				Type:      modelDocumentType,
				ModelJSON: string(modelJSON),
				Brand:     brand,
				Model:     model,
				BrandID:   model.BrandID,
				ModelID:   model.ModelID,
			}); err != nil {
				return err
			}
		}
	}
	return index.Batch(batch)
}
