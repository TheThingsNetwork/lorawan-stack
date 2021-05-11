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
	dr_processor "go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// mockProcessor is a mock messageprocessors.PayloadEncodeDecoder
type mockProcessor struct {
	ch chan dr_processor.PayloadFormatter

	err error
}

func (p *mockProcessor) EncodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	if p.err == nil {
		p.ch <- &ttnpb.MessagePayloadEncoder{
			Formatter:          formatter,
			FormatterParameter: parameter,
		}
	}
	return p.err
}

func (p *mockProcessor) DecodeUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationUplink, formatter ttnpb.PayloadFormatter, parameter string) error {
	if p.err == nil {
		p.ch <- &ttnpb.MessagePayloadDecoder{
			Formatter:          formatter,
			FormatterParameter: parameter,
		}
	}
	return p.err
}

func (p *mockProcessor) DecodeDownlink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, version *ttnpb.EndDeviceVersionIdentifiers, message *ttnpb.ApplicationDownlink, formatter ttnpb.PayloadFormatter, parameter string) error {
	if p.err == nil {
		p.ch <- &ttnpb.MessagePayloadDecoder{
			Formatter:          formatter,
			FormatterParameter: parameter,
		}
	}
	return p.err
}

type mockDR struct {
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
	return fmt.Sprintf("%s:%s:%s:%s", ids.BrandID, ids.ModelID, ids.FirmwareVersion, ids.BandID)
}

var errMock = fmt.Errorf("mock_error")

func (dr *mockDR) GetUplinkDecoder(_ context.Context, req *ttnpb.GetPayloadFormatterRequest) (*ttnpb.MessagePayloadDecoder, error) {
	f, ok := dr.uplinkDecoders[dr.key(req.VersionIDs)]
	if !ok {
		return nil, errMock
	}
	return f, nil
}

func (dr *mockDR) GetDownlinkDecoder(_ context.Context, req *ttnpb.GetPayloadFormatterRequest) (*ttnpb.MessagePayloadDecoder, error) {
	f, ok := dr.downlinkDecoders[dr.key(req.VersionIDs)]
	if !ok {
		return nil, errMock
	}
	return f, nil
}

func (dr *mockDR) GetDownlinkEncoder(_ context.Context, req *ttnpb.GetPayloadFormatterRequest) (*ttnpb.MessagePayloadEncoder, error) {
	f, ok := dr.downlinkEncoders[dr.key(req.VersionIDs)]
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
	ids := &ttnpb.EndDeviceVersionIdentifiers{
		BrandID:         "brand",
		ModelID:         "model",
		FirmwareVersion: "1.0",
		HardwareVersion: "1.1",
		BandID:          "band",
	}
	idsNotFound := &ttnpb.EndDeviceVersionIdentifiers{
		BrandID:         "brand2",
		ModelID:         "model1",
		FirmwareVersion: "1.0",
		HardwareVersion: "1.1",
		BandID:          "band",
	}
	devID := ttnpb.EndDeviceIdentifiers{
		DeviceId: "dev1",
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationId: "app1",
		},
	}

	dr := &mockDR{
		uplinkDecoders:   make(map[string]*ttnpb.MessagePayloadDecoder),
		downlinkDecoders: make(map[string]*ttnpb.MessagePayloadDecoder),
		downlinkEncoders: make(map[string]*ttnpb.MessagePayloadEncoder),
	}
	dr.uplinkDecoders[dr.key(ids)] = &ttnpb.MessagePayloadDecoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: "uplink decoder",
	}
	dr.downlinkDecoders[dr.key(ids)] = &ttnpb.MessagePayloadDecoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: "downlink decoder",
	}
	dr.downlinkEncoders[dr.key(ids)] = &ttnpb.MessagePayloadEncoder{
		Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
		FormatterParameter: "downlink encoder",
	}
	drAddr := dr.start(test.Context())

	ctx := test.Context()
	mockProcessor := &mockProcessor{
		ch:  make(chan dr_processor.PayloadFormatter, 1),
		err: nil,
	}

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

	p := dr_processor.New(mockProcessor, c)

	t.Run("NilDeviceIdentifiers", func(t *testing.T) {
		err := p.DecodeDownlink(test.Context(), devID, nil, nil, "")
		assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)

		select {
		case <-mockProcessor.ch:
			t.Error("Expected timeout but processor was called instead")
			t.FailNow()
		case <-time.After(30 * time.Millisecond):
		}
	})

	t.Run("DeviceNotFound", func(t *testing.T) {
		err := p.DecodeDownlink(test.Context(), devID, idsNotFound, nil, "")
		a := assertions.New(t)
		a.So(err.Error(), should.ContainSubstring, errMock.Error())

		select {
		case <-mockProcessor.ch:
			t.Error("Expected timeout but processor was called instead")
			t.FailNow()
		case <-time.After(30 * time.Millisecond):
		}
	})

	t.Run("UplinkDecoder", func(t *testing.T) {
		err := p.DecodeUplink(test.Context(), devID, ids, nil, "")
		a := assertions.New(t)
		a.So(err, should.BeNil)

		select {
		case f := <-mockProcessor.ch:
			a.So(f, should.Resemble, &ttnpb.MessagePayloadDecoder{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: "uplink decoder",
			})
		case <-time.After(time.Second):
			t.Error("Timed out waiting for message processor")
			t.FailNow()
		}
	})

	t.Run("DownlinkDecoder", func(t *testing.T) {
		err := p.DecodeDownlink(test.Context(), devID, ids, nil, "")
		a := assertions.New(t)
		a.So(err, should.BeNil)

		select {
		case f := <-mockProcessor.ch:
			a.So(f, should.Resemble, &ttnpb.MessagePayloadDecoder{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: "downlink decoder",
			})
		case <-time.After(time.Second):
			t.Error("Timed out waiting for message processor")
			t.FailNow()
		}
	})
	t.Run("DownlinkEncoder", func(t *testing.T) {
		err := p.EncodeDownlink(test.Context(), devID, ids, nil, "")
		a := assertions.New(t)
		a.So(err, should.BeNil)

		select {
		case f := <-mockProcessor.ch:
			a.So(f, should.Resemble, &ttnpb.MessagePayloadEncoder{
				Formatter:          ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
				FormatterParameter: "downlink encoder",
			})
		case <-time.After(time.Second):
			t.Error("Timed out waiting for message processor")
			t.FailNow()
		}
	})

	t.Run("ProcessorError", func(t *testing.T) {
		mockProcessor.err = errMock
		a := assertions.New(t)

		err := p.DecodeDownlink(test.Context(), devID, ids, nil, "")
		a.So(err.Error(), should.ContainSubstring, errMock.Error())
		err = p.DecodeUplink(test.Context(), devID, ids, nil, "")
		a.So(err.Error(), should.ContainSubstring, errMock.Error())
		err = p.EncodeDownlink(test.Context(), devID, ids, nil, "")
		a.So(err.Error(), should.ContainSubstring, errMock.Error())

		mockProcessor.err = nil
	})
}
