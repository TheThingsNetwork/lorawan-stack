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

//go:build slowtests
// +build slowtests

package remote_test

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store/remote"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

const (
	brandWorkers     = 16
	modelWorkers     = 32
	githubRepository = "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-devices/master"
)

// TestGithubDeviceRepository tests that the lorawan-devices GitHub repository
// is fetched and can be used without errors from the DeviceRepository store.
//
// TestGithubDeviceRepository is a slow integration test, and thus disabled by default.
//
// Run test with `$ go test -tags=slowtests ./pkg/devicerepository/store/remote`.
func TestGithubDeviceRepository(t *testing.T) {
	a := assertions.New(t)
	f, err := fetch.FromHTTP(http.DefaultClient, githubRepository)
	a.So(err, should.BeNil)
	s := remote.NewRemoteStore(f)
	brands, err := s.GetBrands(store.GetBrandsRequest{Limit: 0, Paths: []string{"brand_id"}})
	if err != nil {
		t.Fatal(err)
	}

	var (
		totalBrands, totalModels, failures uint32
		brandWg, modelWg                   sync.WaitGroup
	)
	modelCh := make(chan *ttnpb.EndDeviceModel)
	brandCh := make(chan *ttnpb.EndDeviceBrand)

	for i := 0; i < modelWorkers; i++ {
		modelWg.Add(1)
		go func() {
			for model := range modelCh {
				atomic.AddUint32(&totalModels, 1)
				if err := model.ValidateFields(); err != nil {
					fmt.Printf("Failed to validate model %s/%s: %v -- %v\n", model.BrandId, model.ModelId, err, model.DatasheetUrl)
					atomic.AddUint32(&failures, 1)
				}
				for _, fwVer := range model.FirmwareVersions {
					for bandID, profile := range fwVer.Profiles {
						ids := &ttnpb.EndDeviceVersionIdentifiers{
							ModelId:         model.ModelId,
							BrandId:         model.BrandId,
							FirmwareVersion: fwVer.Version,
							BandId:          bandID,
						}
						_, err := s.GetTemplate(ids)
						if err != nil {
							fmt.Printf("Failed to retrieve template for %v: %v\n", ids, err)
							atomic.AddUint32(&failures, 1)
						}

						if profile.CodecId != "" {
							d, err := s.GetUplinkDecoder(&ttnpb.GetPayloadFormatterRequest{VersionIds: ids})
							if err != nil && !strings.Contains(err.Error(), "no_decoder") {
								fmt.Printf("Failed to retrieve uplink decoder for %v: %v\n", ids, err)
								atomic.AddUint32(&failures, 1)
							}
							if err := d.ValidateFields(); err != nil {
								fmt.Printf("Failed to validate uplink encoder for %v: %v\n", ids, err)
								atomic.AddUint32(&failures, 1)
							}
						}
						if profile.CodecId != "" {
							d, err := s.GetDownlinkDecoder(&ttnpb.GetPayloadFormatterRequest{VersionIds: ids})
							if err != nil && !strings.Contains(err.Error(), "no_decoder") {
								fmt.Printf("Failed to retrieve downlink decoder for %v: %v\n", ids, err)
								atomic.AddUint32(&failures, 1)
							}
							if err := d.ValidateFields(); err != nil {
								fmt.Printf("Failed to validate downlink decoder for %v: %v\n", ids, err)
								atomic.AddUint32(&failures, 1)
							}
						}
						if profile.CodecId != "" {
							d, err := s.GetDownlinkEncoder(&ttnpb.GetPayloadFormatterRequest{VersionIds: ids})
							if err != nil && !strings.Contains(err.Error(), "no_encoder") {
								fmt.Printf("Failed to retrieve downlink encoder for %v: %v\n", ids, err)
								atomic.AddUint32(&failures, 1)
							}
							if err := d.ValidateFields(); err != nil {
								fmt.Printf("Failed to validate downlink encoder for %v: %v\n", ids, err)
								atomic.AddUint32(&failures, 1)
							}
						}
					}
				}
			}
			modelWg.Done()
		}()
	}

	for i := 0; i < brandWorkers; i++ {
		brandWg.Add(1)
		go func() {
			for brand := range brandCh {
				if err := brand.ValidateFields(); err != nil {
					fmt.Printf("Failed to validate brand %s: %v\n", brand.BrandId, err)
					atomic.AddUint32(&failures, 1)
				}
				models, err := s.GetModels(store.GetModelsRequest{BrandID: brand.BrandId, Limit: 0, Paths: ttnpb.EndDeviceModelFieldPathsNested})
				if err != nil {
					if !strings.Contains(err.Error(), "not found") {
						fmt.Printf("Failed fetching brand %s\n", brand.BrandId)
						atomic.AddUint32(&failures, 1)
					}
					continue
				}
				fmt.Printf("%v %d\n", brand.BrandId, len(models.Models))
				atomic.AddUint32(&totalBrands, 1)
				for _, model := range models.Models {
					modelCh <- model
				}
			}
			brandWg.Done()
		}()
	}

	for _, brand := range brands.Brands {
		brandCh <- brand
	}
	close(brandCh)
	brandWg.Wait()
	close(modelCh)
	modelWg.Wait()

	if totalModels == 0 {
		t.Fatal("No models found")
	}
	if failures > 0 {
		t.Fatalf("%d validation failures", failures)
	}
}
