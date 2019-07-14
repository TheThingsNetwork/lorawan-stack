// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// Package devicetemplateconverter provides device template services.
package devicetemplateconverter

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config represents the DeviceTemplateConverter configuration.
type Config struct {
}

// DeviceTemplateConverter implements the Device Template Converter component.
//
// The Device Template Converter exposes the EndDeviceTemplateConverter service.
type DeviceTemplateConverter struct {
	*component.Component
	ctx context.Context

	grpc struct {
		endDeviceTemplateConverter *endDeviceTemplateConverterServer
	}
}

// New returns a new *DeviceTemplateConverter.
func New(c *component.Component, conf *Config) (*DeviceTemplateConverter, error) {
	dtc := &DeviceTemplateConverter{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "devicetemplateconverter"),
	}
	dtc.grpc.endDeviceTemplateConverter = &endDeviceTemplateConverterServer{DTC: dtc}

	c.RegisterGRPC(dtc)
	return dtc, nil
}

// Context returns the context of the Device Template Converter.
func (dtc *DeviceTemplateConverter) Context() context.Context {
	return dtc.ctx
}

// Roles returns the roles that the Device Template Converter fulfills.
func (dtc *DeviceTemplateConverter) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_DEVICE_TEMPLATE_CONVERTER}
}

// RegisterServices registers services provided by dtc at s.
func (dtc *DeviceTemplateConverter) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEndDeviceTemplateConverterServer(s, dtc.grpc.endDeviceTemplateConverter)
}

// RegisterHandlers registers gRPC handlers.
func (dtc *DeviceTemplateConverter) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEndDeviceTemplateConverterHandler(dtc.Context(), s, conn)
}
