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

// Package qrcodegenerator provides QR code generation services.
package qrcodegenerator

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config represents the QRCodeGenerator configuration.
type Config struct {
}

// QRCodeGenerator implements the QR Code Generator component.
//
// The QR Code Generator exposes the EndDeviceQRCodeGenerator service.
type QRCodeGenerator struct {
	*component.Component
	ctx context.Context

	grpc struct {
		endDeviceQRCodeGenerator *endDeviceQRCodeGeneratorServer
	}
}

var errFormatNotFound = errors.DefineNotFound("format_not_found", "format `{id}` not found")

// New returns a new *QRCodeGenerator.
func New(c *component.Component, conf *Config) (*QRCodeGenerator, error) {
	qrg := &QRCodeGenerator{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "qrcodegenerator"),
	}
	qrg.grpc.endDeviceQRCodeGenerator = &endDeviceQRCodeGeneratorServer{QRG: qrg}

	c.RegisterGRPC(qrg)
	return qrg, nil
}

// Context returns the context of the QR Code Generator.
func (qrg *QRCodeGenerator) Context() context.Context {
	return qrg.ctx
}

// Roles returns the roles that the QR Code Generator fulfills.
func (qrg *QRCodeGenerator) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_QR_CODE_GENERATOR}
}

// RegisterServices registers services provided by qrg at s.
func (qrg *QRCodeGenerator) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEndDeviceQRCodeGeneratorServer(s, qrg.grpc.endDeviceQRCodeGenerator)
}

// RegisterHandlers registers gRPC handlers.
func (qrg *QRCodeGenerator) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEndDeviceQRCodeGeneratorHandler(qrg.Context(), s, conn)
}
