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

package devicerepository_test

import (
	"fmt"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockStore struct {
	// last requests
	lastGetBrandsRequest store.GetBrandsRequest
	lastGetModelsRequest store.GetModelsRequest
	lastVersionIDs       *ttnpb.EndDeviceVersionIdentifiers
	lastCodecPaths       []string

	// mock responses
	brands   []*ttnpb.EndDeviceBrand
	models   []*ttnpb.EndDeviceModel
	template *ttnpb.EndDeviceTemplate
	uplinkDecoder,
	downlinkDecoder *ttnpb.MessagePayloadDecoder
	downlinkEncoder *ttnpb.MessagePayloadEncoder

	// mock errors
	err error
}

// GetBrands lists available end device vendors.
func (s *mockStore) GetBrands(req store.GetBrandsRequest) (*store.GetBrandsResponse, error) {
	s.lastGetBrandsRequest = req
	if s.err != nil {
		return nil, s.err
	}
	if s.brands == nil {
		s.brands = []*ttnpb.EndDeviceBrand{}
	}
	return &store.GetBrandsResponse{
		Count:  uint32(len(s.brands)),
		Offset: 0,
		Total:  uint32(len(s.brands)),
		Brands: s.brands,
	}, nil
}

// GetModels lists available end device definitions.
func (s *mockStore) GetModels(req store.GetModelsRequest) (*store.GetModelsResponse, error) {
	s.lastGetModelsRequest = req
	if s.err != nil {
		return nil, s.err
	}
	if s.models == nil {
		s.models = []*ttnpb.EndDeviceModel{}
	}
	return &store.GetModelsResponse{
		Count:  uint32(len(s.models)),
		Offset: 0,
		Total:  uint32(len(s.models)),
		Models: s.models,
	}, nil
}

// GetEndDeviceProfiles implements store.
func (*mockStore) GetEndDeviceProfiles(
	store.GetEndDeviceProfilesRequest,
) (*store.GetEndDeviceProfilesResponse, error) {
	return nil, nil //nolint:nilnil
}

// GetTemplate retrieves an end device template for an end device definition.
func (s *mockStore) GetTemplate(
	req *ttnpb.GetTemplateRequest,
	_ *store.EndDeviceProfile,
) (*ttnpb.EndDeviceTemplate, error) {
	s.lastVersionIDs = req.GetVersionIds()
	profileIDs := req.GetEndDeviceProfileIds()
	if profileIDs != nil {
		s.lastVersionIDs = &ttnpb.EndDeviceVersionIdentifiers{
			BrandId: "test-brand",
		}
	}
	return s.template, s.err
}

func (*mockStore) GetEndDeviceProfileIdentifiers(
	_ store.GetEndDeviceProfileIdentifiersRequest,
) (*store.GetEndDeviceProfileIdentifiersResponse, error) {
	return nil, nil //nolint:nilnil
}

// GetUplinkDecoder retrieves the codec for decoding uplink messages.
func (s *mockStore) GetUplinkDecoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	s.lastVersionIDs = req.GetVersionIds()
	s.lastCodecPaths = req.GetFieldMask().GetPaths()
	return s.uplinkDecoder, s.err
}

// GetDownlinkDecoder retrieves the codec for decoding downlink messages.
func (s *mockStore) GetDownlinkDecoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadDecoder, error) {
	s.lastVersionIDs = req.GetVersionIds()
	s.lastCodecPaths = req.GetFieldMask().GetPaths()
	return s.downlinkDecoder, s.err
}

// GetDownlinkEncoder retrieves the codec for encoding downlink messages.
func (s *mockStore) GetDownlinkEncoder(req store.GetCodecRequest) (*ttnpb.MessagePayloadEncoder, error) {
	s.lastVersionIDs = req.GetVersionIds()
	s.lastCodecPaths = req.GetFieldMask().GetPaths()
	return s.downlinkEncoder, s.err
}

// Close closes the store.
func (*mockStore) Close() error {
	return nil
}

func TestGRPC(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	mockIS, mockISAddr, closeIS := mockis.New(ctx)
	defer closeIS()

	componentConfig := &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: mockISAddr,
			},
		},
	}
	c := componenttest.NewComponent(t, componentConfig)

	st := &mockStore{}
	conf := &devicerepository.Config{
		Store: devicerepository.StoreConfig{
			Store: st,
		},
		AssetsBaseURL: "https://assets/",
	}
	dr, err := devicerepository.New(c, conf)
	test.Must(dr, err)

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)
	cc := dr.LoopbackConn()
	cl := ttnpb.NewDeviceRepositoryClient(cc)

	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     "MockValue",
		AllowInsecure: true,
	})

	t.Run("Auth", func(t *testing.T) {
		for _, atc := range []struct {
			name    string
			execute func(opts ...grpc.CallOption) (any, error)
		}{
			{
				name: "ListBrands",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.ListBrands(ctx, &ttnpb.ListEndDeviceBrandsRequest{}, opts...)
					return response, err
				},
			},
			{
				name: "GetBrand",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.GetBrand(ctx, &ttnpb.GetEndDeviceBrandRequest{
						BrandId: "brand1",
					}, opts...)
					return response, err
				},
			},
			{
				name: "ListModels",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.ListModels(ctx, &ttnpb.ListEndDeviceModelsRequest{
						BrandId: "brand1",
					}, opts...)
					return response, err
				},
			},
			{
				name: "GetModel",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.GetModel(ctx, &ttnpb.GetEndDeviceModelRequest{
						BrandId: "brand1",
						ModelId: "model1",
					}, opts...)
					return response, err
				},
			},
			{
				name: "GetTemplate",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
						VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
							BrandId: "brand1",
							ModelId: "model1",
						},
					}, opts...)
					return response, err
				},
			},
			{
				name: "GetUplinkDecoder",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.GetUplinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
						VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
							BrandId: "brand1",
							ModelId: "model1",
						},
					}, opts...)
					return response, err
				},
			},
			{
				name: "GetDownlinkDecoder",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.GetDownlinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
						VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
							BrandId: "brand1",
							ModelId: "model1",
						},
					}, opts...)
					return response, err
				},
			},
			{
				name: "GetDownlinkEncoder",
				execute: func(opts ...grpc.CallOption) (any, error) {
					response, err := cl.GetDownlinkEncoder(ctx, &ttnpb.GetPayloadFormatterRequest{
						VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
							BrandId: "brand1",
							ModelId: "model1",
						},
					}, opts...)
					return response, err
				},
			},
		} {
			t.Run(atc.name, func(t *testing.T) {
				for _, tc := range []struct {
					name               string
					opts               []grpc.CallOption
					injectEntityAccess func()
					assertion          func(err error) bool
				}{
					{
						name:      "no_creds",
						assertion: errors.IsUnauthenticated,
					},
					{
						name:               "no_rights",
						opts:               []grpc.CallOption{creds},
						injectEntityAccess: func() { mockIS.EntityAccess().SetAuthInfoResponse(&ttnpb.AuthInfoResponse{}) },
						assertion:          errors.IsUnauthenticated,
					},
				} {
					t.Run(tc.name, func(t *testing.T) {
						a := assertions.New(t)
						defer func() {
							mockIS.EntityAccess().SetAuthInfoResponse(nil)
							mockIS.EntityAccess().SetErr(nil)
						}()

						if tc.injectEntityAccess != nil {
							tc.injectEntityAccess()
						}

						response, err := atc.execute(tc.opts...)
						a.So(response, should.BeNil)
						t.Log(err)
						a.So(tc.assertion(err), should.BeTrue)
					})
				}
			})
		}
	})

	ids := &ttnpb.EndDeviceVersionIdentifiers{
		BrandId:         "brand",
		ModelId:         "model",
		FirmwareVersion: "1.0",
		HardwareVersion: "1.0",
		BandId:          "band",
	}

	// Add rights to entity access mock.
	mockIS.EntityAccess().SetAuthInfoResponse(&ttnpb.AuthInfoResponse{
		AccessMethod: &ttnpb.AuthInfoResponse_ApiKey{
			ApiKey: &ttnpb.AuthInfoResponse_APIKeyAccess{
				ApiKey: &ttnpb.APIKey{
					// The specific right here doesn't have an affect since the
					// validation that happens only checks if its nil.
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
				},
			},
		},
	})

	t.Run("ListBrands", func(t *testing.T) {
		t.Run("Request", func(t *testing.T) {
			a := assertions.New(t)

			_, err := cl.ListBrands(ctx, &ttnpb.ListEndDeviceBrandsRequest{
				Limit:     100,
				Page:      2,
				OrderBy:   "brand_id",
				Search:    "query string",
				FieldMask: ttnpb.FieldMask("lora_alliance_vendor_id"),
			}, creds)
			a.So(err, should.BeNil)
			a.So(st.lastGetBrandsRequest, should.Resemble, store.GetBrandsRequest{
				BrandID: "",
				Limit:   100,
				Page:    2,
				OrderBy: "brand_id",
				Paths:   []string{"lora_alliance_vendor_id", "brand_id"},
				Search:  "query string",
			})
		})

		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.err = fmt.Errorf("store error")
			brands, err := cl.ListBrands(ctx, &ttnpb.ListEndDeviceBrandsRequest{}, creds)
			a.So(brands, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.brands = []*ttnpb.EndDeviceBrand{
				{
					BrandId:                 "brand1",
					PrivateEnterpriseNumber: 100,
					Logo:                    "item.png",
				},
				{BrandId: "brand2"},
			}
			st.err = nil

			responseHeaders := metadata.MD{}
			brands, err := cl.ListBrands(ctx, &ttnpb.ListEndDeviceBrandsRequest{}, grpc.Header(&responseHeaders), creds)
			a.So(err, should.BeNil)
			a.So(brands, should.Resemble, &ttnpb.ListEndDeviceBrandsResponse{
				Brands: []*ttnpb.EndDeviceBrand{
					{
						BrandId:                 "brand1",
						PrivateEnterpriseNumber: 100,
						Logo:                    "https://assets/vendor/brand1/item.png",
					},
					{
						BrandId: "brand2",
					},
				},
			})

			s := responseHeaders.Get("x-total-count")
			a.So(s, should.Resemble, []string{"2"})
		})
	})

	t.Run("GetBrand", func(t *testing.T) {
		t.Run("Request", func(t *testing.T) {
			a := assertions.New(t)

			_, err := cl.GetBrand(ctx, &ttnpb.GetEndDeviceBrandRequest{
				BrandId:   "brand1",
				FieldMask: ttnpb.FieldMask("lora_alliance_vendor_id"),
			}, creds)
			a.So(err, should.BeNil)
			a.So(st.lastGetBrandsRequest, should.Resemble, store.GetBrandsRequest{
				Limit:   1,
				BrandID: "brand1",
				Paths:   []string{"lora_alliance_vendor_id", "brand_id"},
			})
		})

		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.err = fmt.Errorf("store error")
			brands, err := cl.ListBrands(ctx, &ttnpb.ListEndDeviceBrandsRequest{}, creds)
			a.So(brands, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.brands = []*ttnpb.EndDeviceBrand{
				{
					BrandId:                 "brand1",
					PrivateEnterpriseNumber: 100,
					Logo:                    "item.png",
				},
			}
			st.err = nil

			brand, err := cl.GetBrand(ctx, &ttnpb.GetEndDeviceBrandRequest{
				BrandId: "brand1",
			}, creds)
			a.So(err, should.BeNil)
			a.So(brand, should.Resemble, &ttnpb.EndDeviceBrand{
				BrandId:                 "brand1",
				PrivateEnterpriseNumber: 100,
				Logo:                    "https://assets/vendor/brand1/item.png",
			})
		})
	})

	t.Run("ListModels", func(t *testing.T) {
		t.Run("Request", func(t *testing.T) {
			a := assertions.New(t)

			_, err := cl.ListModels(ctx, &ttnpb.ListEndDeviceModelsRequest{
				BrandId:   "brand1",
				Limit:     100,
				Page:      2,
				OrderBy:   "brand_id",
				Search:    "query string",
				FieldMask: ttnpb.FieldMask("firmware_versions"),
			}, creds)
			a.So(err, should.BeNil)
			a.So(st.lastGetModelsRequest, should.Resemble, store.GetModelsRequest{
				ModelID: "",
				BrandID: "brand1",
				Limit:   100,
				Page:    2,
				OrderBy: "brand_id",
				Paths:   []string{"firmware_versions", "brand_id", "model_id"},
				Search:  "query string",
			})
		})

		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.err = fmt.Errorf("store error")
			res, err := cl.ListModels(ctx, &ttnpb.ListEndDeviceModelsRequest{}, creds)
			a.So(res, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.models = []*ttnpb.EndDeviceModel{
				{
					BrandId: "brand1",
					ModelId: "model1",
					Photos: &ttnpb.EndDeviceModel_Photos{
						Main:  "a.png",
						Other: []string{"b.png"},
					},
				},
				{
					BrandId: "brand2",
					ModelId: "model2",
				},
			}
			st.err = nil

			responseHeaders := metadata.MD{}
			brands, err := cl.ListModels(ctx, &ttnpb.ListEndDeviceModelsRequest{}, grpc.Header(&responseHeaders), creds)
			a.So(err, should.BeNil)
			a.So(brands, should.Resemble, &ttnpb.ListEndDeviceModelsResponse{
				Models: []*ttnpb.EndDeviceModel{
					{
						BrandId: "brand1",
						ModelId: "model1",
						Photos: &ttnpb.EndDeviceModel_Photos{
							Main:  "https://assets/vendor/brand1/a.png",
							Other: []string{"https://assets/vendor/brand1/b.png"},
						},
					},
					{
						BrandId: "brand2",
						ModelId: "model2",
					},
				},
			})

			s := responseHeaders.Get("x-total-count")
			a.So(s, should.Resemble, []string{"2"})
		})
	})

	t.Run("GetModel", func(t *testing.T) {
		t.Run("Request", func(t *testing.T) {
			a := assertions.New(t)

			_, err := cl.GetModel(ctx, &ttnpb.GetEndDeviceModelRequest{
				BrandId:   "brand1",
				ModelId:   "model1",
				FieldMask: ttnpb.FieldMask("firmware_versions"),
			}, creds)
			a.So(err, should.BeNil)
			a.So(st.lastGetModelsRequest, should.Resemble, store.GetModelsRequest{
				Limit:   1,
				BrandID: "brand1",
				ModelID: "model1",
				Paths:   []string{"firmware_versions", "brand_id", "model_id"},
			})
		})

		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.err = fmt.Errorf("store error")
			models, err := cl.ListModels(ctx, &ttnpb.ListEndDeviceModelsRequest{}, creds)
			a.So(models, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.models = []*ttnpb.EndDeviceModel{
				{
					BrandId: "brand1",
					ModelId: "model1",
					Photos: &ttnpb.EndDeviceModel_Photos{
						Main:  "a.png",
						Other: []string{"b.png"},
					},
				},
			}
			st.err = nil

			model, err := cl.GetModel(ctx, &ttnpb.GetEndDeviceModelRequest{
				BrandId: "brand1",
				ModelId: "model1",
			}, creds)
			a.So(err, should.BeNil)
			a.So(model, should.Resemble, &ttnpb.EndDeviceModel{
				BrandId: "brand1",
				ModelId: "model1",
				Photos: &ttnpb.EndDeviceModel_Photos{
					Main:  "https://assets/vendor/brand1/a.png",
					Other: []string{"https://assets/vendor/brand1/b.png"},
				},
			})
		})
	})

	t.Run("GetTemplate", func(t *testing.T) {
		st.template = &ttnpb.EndDeviceTemplate{
			EndDevice: &ttnpb.EndDevice{
				VersionIds: ids,
			},
			FieldMask: ttnpb.FieldMask("version_ids"),
		}

		t.Run("Request", func(t *testing.T) {
			a := assertions.New(t)
			_, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
				VersionIds: ids,
			}, creds)
			a.So(err, should.BeNil)
			a.So(st.lastVersionIDs, should.Resemble, ids)
		})

		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.err = fmt.Errorf("store error")
			models, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
				VersionIds: ids,
			}, creds)
			a.So(models, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.err = nil

			template, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
				VersionIds: ids,
			}, creds)
			a.So(err, should.BeNil)
			a.So(template, should.Resemble, st.template)
		})
	})

	t.Run("GetTemplateByNumericIDs", func(t *testing.T) {
		profileIDs := &ttnpb.GetTemplateRequest_EndDeviceProfileIdentifiers{
			VendorId:        1,
			VendorProfileId: 0,
		}
		st.template = &ttnpb.EndDeviceTemplate{
			EndDevice: &ttnpb.EndDevice{
				VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
					BrandId: "test-brand",
				},
			},
			FieldMask: ttnpb.FieldMask("version_ids"),
		}

		t.Run("Request", func(t *testing.T) {
			a := assertions.New(t)
			_, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
				EndDeviceProfileIds: profileIDs,
			}, creds)
			a.So(err, should.BeNil)
			a.So(st.lastVersionIDs, should.Resemble, &ttnpb.EndDeviceVersionIdentifiers{
				BrandId: "test-brand",
			})
		})

		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.err = fmt.Errorf("store error")
			models, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
				EndDeviceProfileIds: profileIDs,
			}, creds)
			a.So(models, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.err = nil

			template, err := cl.GetTemplate(ctx, &ttnpb.GetTemplateRequest{
				EndDeviceProfileIds: profileIDs,
			}, creds)
			a.So(err, should.BeNil)
			a.So(template, should.Resemble, st.template)
		})
	})

	t.Run("GetUplinkDecoder", func(t *testing.T) {
		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.lastVersionIDs = nil
			st.lastCodecPaths = nil
			st.err = fmt.Errorf("store error")
			c, err := cl.GetUplinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, creds)
			a.So(c, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
			a.So(st.lastVersionIDs, should.Resemble, ids)
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.uplinkDecoder = &ttnpb.MessagePayloadDecoder{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: "uplink decoder",
			}
			st.err = nil

			c, err := cl.GetUplinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, creds)
			a.So(err, should.BeNil)
			a.So(c, should.Resemble, st.uplinkDecoder)
		})
		t.Run("ClusterAuth", func(t *testing.T) {
			codec, err := cl.GetUplinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			})
			a := assertions.New(t)
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
			a.So(codec, should.BeNil)

			_, err = cl.GetUplinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, c.WithClusterAuth())
			a.So(err, should.BeNil)
		})
	})

	t.Run("GetDownlinkDecoder", func(t *testing.T) {
		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.lastVersionIDs = nil
			st.lastCodecPaths = nil
			st.err = fmt.Errorf("store error")
			c, err := cl.GetDownlinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, creds)
			a.So(c, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
			a.So(st.lastVersionIDs, should.Resemble, ids)
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.downlinkDecoder = &ttnpb.MessagePayloadDecoder{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: "downlink decoder script",
			}
			st.err = nil

			c, err := cl.GetDownlinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, creds)
			a.So(err, should.BeNil)
			a.So(c, should.Resemble, st.downlinkDecoder)
		})
		t.Run("ClusterAuth", func(t *testing.T) {
			codec, err := cl.GetDownlinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			})
			a := assertions.New(t)
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
			a.So(codec, should.BeNil)

			_, err = cl.GetDownlinkDecoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, c.WithClusterAuth())
			a.So(err, should.BeNil)
		})
	})

	t.Run("GetDownlinkEncoder", func(t *testing.T) {
		t.Run("StoreError", func(t *testing.T) {
			a := assertions.New(t)
			st.lastVersionIDs = nil
			st.lastCodecPaths = nil
			st.err = fmt.Errorf("store error")
			c, err := cl.GetDownlinkEncoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, creds)
			a.So(c, should.BeNil)
			a.So(err.Error(), should.ContainSubstring, st.err.Error())
			a.So(st.lastVersionIDs, should.Resemble, ids)
		})

		t.Run("Success", func(t *testing.T) {
			a := assertions.New(t)
			st.downlinkEncoder = &ttnpb.MessagePayloadEncoder{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: "downlink encoder script",
			}
			st.err = nil

			c, err := cl.GetDownlinkEncoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, creds)
			a.So(err, should.BeNil)
			a.So(c, should.Resemble, st.downlinkEncoder)
		})
		t.Run("ClusterAuth", func(t *testing.T) {
			codec, err := cl.GetDownlinkEncoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			})
			a := assertions.New(t)
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
			a.So(codec, should.BeNil)

			_, err = cl.GetDownlinkEncoder(ctx, &ttnpb.GetPayloadFormatterRequest{
				VersionIds: ids,
			}, c.WithClusterAuth())
			a.So(err, should.BeNil)
		})
	})
}
