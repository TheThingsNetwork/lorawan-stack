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
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	dr_processor "go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockEncoderDecoder struct {
	encodeDownlink func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error
	decodeUplink   func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error
	decodeDownlink func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error
}

func (m mockEncoderDecoder) EncodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
	return m.encodeDownlink(ctx, ids, version, message, parameter)
}

func (m mockEncoderDecoder) DecodeUplink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error {
	return m.decodeUplink(ctx, ids, version, message, parameter)
}

func (m mockEncoderDecoder) DecodeDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
	return m.decodeDownlink(ctx, ids, version, message, parameter)
}

type mockCompilableEncoderDecoder struct {
	mockEncoderDecoder

	compileDownlinkEncoder func(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error)
	compileUplinkDecoder   func(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationUplink) error, error)
	compileDownlinkDecoder func(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error)
}

func (m mockCompilableEncoderDecoder) CompileDownlinkEncoder(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error) {
	return m.compileDownlinkEncoder(ctx, parameter)
}

func (m mockCompilableEncoderDecoder) CompileUplinkDecoder(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationUplink) error, error) {
	return m.compileUplinkDecoder(ctx, parameter)
}

func (m mockCompilableEncoderDecoder) CompileDownlinkDecoder(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error) {
	return m.compileDownlinkDecoder(ctx, parameter)
}

type mockProvider struct {
	processor messageprocessors.PayloadEncoderDecoder

	err error
}

func (m mockProvider) GetPayloadEncoderDecoder(ctx context.Context, formatter ttnpb.PayloadFormatter) (messageprocessors.PayloadEncoderDecoder, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.processor, nil
}

type mockDR struct {
	ttnpb.UnimplementedDeviceRepositoryServer

	uplinkDecoders,
	downlinkDecoders map[string]*ttnpb.MessagePayloadDecoder
	downlinkEncoders map[string]*ttnpb.MessagePayloadEncoder
}

func (dr *mockDR) ListBrands(_ context.Context, _ *ttnpb.ListEndDeviceBrandsRequest) (*ttnpb.ListEndDeviceBrandsResponse, error) {
	panic("not implemented")
}

func (dr *mockDR) GetBrand(_ context.Context, _ *ttnpb.GetEndDeviceBrandRequest) (*ttnpb.EndDeviceBrand, error) {
	panic("not implemented")
}

func (dr *mockDR) ListModels(_ context.Context, _ *ttnpb.ListEndDeviceModelsRequest) (*ttnpb.ListEndDeviceModelsResponse, error) {
	panic("not implemented")
}

func (dr *mockDR) GetModel(_ context.Context, _ *ttnpb.GetEndDeviceModelRequest) (*ttnpb.EndDeviceModel, error) {
	panic("not implemented")
}

func (dr *mockDR) GetTemplate(_ context.Context, _ *ttnpb.GetTemplateRequest) (*ttnpb.EndDeviceTemplate, error) {
	panic("not implemented")
}

func (dr *mockDR) key(ids *ttnpb.EndDeviceVersionIdentifiers) string {
	return fmt.Sprintf("%s:%s:%s:%s", ids.BrandId, ids.ModelId, ids.FirmwareVersion, ids.BandId)
}

var errMock = fmt.Errorf("mock_error")

func (dr *mockDR) GetUplinkDecoder(_ context.Context, req *ttnpb.GetPayloadFormatterRequest) (*ttnpb.MessagePayloadDecoder, error) {
	f, ok := dr.uplinkDecoders[dr.key(req.VersionIds)]
	if !ok {
		return nil, errMock
	}
	return f, nil
}

func (dr *mockDR) GetDownlinkDecoder(_ context.Context, req *ttnpb.GetPayloadFormatterRequest) (*ttnpb.MessagePayloadDecoder, error) {
	f, ok := dr.downlinkDecoders[dr.key(req.VersionIds)]
	if !ok {
		return nil, errMock
	}
	return f, nil
}

func (dr *mockDR) GetDownlinkEncoder(_ context.Context, req *ttnpb.GetPayloadFormatterRequest) (*ttnpb.MessagePayloadEncoder, error) {
	f, ok := dr.downlinkEncoders[dr.key(req.VersionIds)]
	if !ok {
		return nil, errMock
	}
	return f, nil
}

// start mock device repository and return listen address.
func (dr *mockDR) start(ctx context.Context) string {
	srv := rpcserver.New(ctx)
	ttnpb.RegisterDeviceRepositoryServer(srv.Server, dr)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return lis.Addr().String()
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

func TestDeviceRepository(t *testing.T) {
	versionIDs := &ttnpb.EndDeviceVersionIdentifiers{
		BrandId:         "brand",
		ModelId:         "model",
		FirmwareVersion: "1.0",
		HardwareVersion: "1.1",
		BandId:          "band",
	}
	idsNotFound := &ttnpb.EndDeviceVersionIdentifiers{
		BrandId:         "brand2",
		ModelId:         "model1",
		FirmwareVersion: "1.0",
		HardwareVersion: "1.1",
		BandId:          "band",
	}
	devIDs := &ttnpb.EndDeviceIdentifiers{
		DeviceId: "dev1",
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "app1",
		},
	}

	dr := &mockDR{
		uplinkDecoders:   make(map[string]*ttnpb.MessagePayloadDecoder),
		downlinkDecoders: make(map[string]*ttnpb.MessagePayloadDecoder),
		downlinkEncoders: make(map[string]*ttnpb.MessagePayloadEncoder),
	}
	dr.uplinkDecoders[dr.key(versionIDs)] = &ttnpb.MessagePayloadDecoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: "uplink decoder",
	}
	dr.downlinkDecoders[dr.key(versionIDs)] = &ttnpb.MessagePayloadDecoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: "downlink decoder",
	}
	dr.downlinkEncoders[dr.key(versionIDs)] = &ttnpb.MessagePayloadEncoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: "downlink encoder",
	}
	drAddr := dr.start(test.Context())

	ctx := test.Context()

	// start mock device repository
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				DeviceRepository: drAddr,
			},
		},
	})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_REPOSITORY)

	t.Run("NilDeviceIdentifiers", func(t *testing.T) {
		p := dr_processor.New(&mockProvider{}, c)

		err := p.DecodeDownlink(test.Context(), devIDs, nil, nil, "")
		assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
	})

	t.Run("DeviceNotFound", func(t *testing.T) {
		p := dr_processor.New(&mockProvider{}, c)

		err := p.DecodeDownlink(test.Context(), devIDs, idsNotFound, nil, "")
		a := assertions.New(t)
		a.So(err.Error(), should.ContainSubstring, errMock.Error())
	})

	t.Run("UplinkDecoder-Simple", func(t *testing.T) {
		a := assertions.New(t)

		called := false
		mockProvider := &mockProvider{
			processor: mockEncoderDecoder{
				decodeUplink: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error {
					a.So(ids, should.Resemble, devIDs)
					a.So(version, should.Resemble, versionIDs)
					a.So(message, should.BeNil)
					a.So(parameter, should.Equal, "uplink decoder")

					called = true

					return nil
				},
			},
		}
		p := dr_processor.New(mockProvider, c)

		err := p.DecodeUplink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err, should.BeNil)

		a.So(called, should.BeTrue)
	})
	t.Run("UplinkDecoder-Compile", func(t *testing.T) {
		a := assertions.New(t)

		calledCompile := false
		calledRun := false
		mockProvider := mockProvider{
			processor: mockCompilableEncoderDecoder{
				mockEncoderDecoder: mockEncoderDecoder{
					decodeUplink: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, parameter string) error {
						t.Error("Direct uplink decoder should not be called")
						return nil
					},
				},
				compileUplinkDecoder: func(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationUplink) error, error) {
					a.So(parameter, should.Equal, "uplink decoder")

					calledCompile = true

					return func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink) error {
						a.So(ids, should.Resemble, devIDs)
						a.So(version, should.Resemble, versionIDs)
						a.So(message, should.BeNil)

						calledRun = true

						return nil
					}, nil
				},
			},
		}
		p := dr_processor.New(mockProvider, c)

		err := p.DecodeUplink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err, should.BeNil)

		a.So(calledCompile, should.BeTrue)
		a.So(calledRun, should.BeTrue)
	})

	t.Run("DownlinkDecoder-Simple", func(t *testing.T) {
		a := assertions.New(t)

		called := false
		mockProvider := &mockProvider{
			processor: mockEncoderDecoder{
				decodeDownlink: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
					a.So(ids, should.Resemble, devIDs)
					a.So(version, should.Resemble, versionIDs)
					a.So(message, should.BeNil)
					a.So(parameter, should.Equal, "downlink decoder")

					called = true

					return nil
				},
			},
		}
		p := dr_processor.New(mockProvider, c)

		err := p.DecodeDownlink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err, should.BeNil)

		a.So(called, should.BeTrue)
	})
	t.Run("DownlinkDecoder-Compile", func(t *testing.T) {
		a := assertions.New(t)

		calledCompile := false
		calledRun := false
		mockProvider := mockProvider{
			processor: mockCompilableEncoderDecoder{
				mockEncoderDecoder: mockEncoderDecoder{
					decodeDownlink: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
						t.Error("Direct downlink decoder should not be called")
						return nil
					},
				},
				compileDownlinkDecoder: func(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error) {
					a.So(parameter, should.Equal, "downlink decoder")

					calledCompile = true

					return func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink) error {
						a.So(ids, should.Resemble, devIDs)
						a.So(version, should.Resemble, versionIDs)
						a.So(message, should.BeNil)

						calledRun = true

						return nil
					}, nil
				},
			},
		}
		p := dr_processor.New(mockProvider, c)

		err := p.DecodeDownlink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err, should.BeNil)

		a.So(calledCompile, should.BeTrue)
		a.So(calledRun, should.BeTrue)
	})

	t.Run("DownlinkEncoder-Simple", func(t *testing.T) {
		a := assertions.New(t)

		called := false
		mockProvider := &mockProvider{
			processor: mockEncoderDecoder{
				encodeDownlink: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
					a.So(ids, should.Resemble, devIDs)
					a.So(version, should.Resemble, versionIDs)
					a.So(message, should.BeNil)
					a.So(parameter, should.Equal, "downlink encoder")

					called = true

					return nil
				},
			},
		}
		p := dr_processor.New(mockProvider, c)

		err := p.EncodeDownlink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err, should.BeNil)

		a.So(called, should.BeTrue)
	})
	t.Run("DownlinkEncoder-Compile", func(t *testing.T) {
		a := assertions.New(t)

		calledCompile := false
		calledRun := false
		mockProvider := mockProvider{
			processor: mockCompilableEncoderDecoder{
				mockEncoderDecoder: mockEncoderDecoder{
					encodeDownlink: func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, parameter string) error {
						t.Error("Direct downlink encoder should not be called")
						return nil
					},
				},
				compileDownlinkEncoder: func(ctx context.Context, parameter string) (func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDeviceVersionIdentifiers, *ttnpb.ApplicationDownlink) error, error) {
					a.So(parameter, should.Equal, "downlink encoder")

					calledCompile = true

					return func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink) error {
						a.So(ids, should.Resemble, devIDs)
						a.So(version, should.Resemble, versionIDs)
						a.So(message, should.BeNil)

						calledRun = true

						return nil
					}, nil
				},
			},
		}
		p := dr_processor.New(mockProvider, c)

		err := p.EncodeDownlink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err, should.BeNil)

		a.So(calledCompile, should.BeTrue)
		a.So(calledRun, should.BeTrue)
	})

	t.Run("ProviderError", func(t *testing.T) {
		mockProvider := mockProvider{
			err: errMock,
		}
		p := dr_processor.New(mockProvider, c)

		a := assertions.New(t)

		err := p.DecodeDownlink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err.Error(), should.ContainSubstring, errMock.Error())
		err = p.DecodeUplink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err.Error(), should.ContainSubstring, errMock.Error())
		err = p.EncodeDownlink(test.Context(), devIDs, versionIDs, nil, "")
		a.So(err.Error(), should.ContainSubstring, errMock.Error())
	})
}
