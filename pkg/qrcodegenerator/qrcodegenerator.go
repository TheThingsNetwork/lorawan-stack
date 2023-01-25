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

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode/enddevices"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// QRCodeGenerator implements the QR Code Generator component.
//
// The QR Code Generator exposes the EndDeviceQRCodeGenerator service.
type QRCodeGenerator struct {
	*component.Component
	ctx context.Context

	endDevices *enddevices.Server

	grpc struct {
		endDeviceQRCodeGenerator *endDeviceQRCodeGeneratorServer
	}
}

var errFormatNotFound = errors.DefineNotFound("format_not_found", "format `{id}` not found")

// New returns a new *QRCodeGenerator.
func New(c *component.Component, conf *Config, opts ...Option) (*QRCodeGenerator, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "qrcodegenerator")
	qrg := &QRCodeGenerator{
		Component: c,
		ctx:       ctx,
	}
	qrg.grpc.endDeviceQRCodeGenerator = &endDeviceQRCodeGeneratorServer{QRG: qrg}
	qrg.endDevices = enddevices.New(ctx)

	c.RegisterGRPC(qrg)

	for _, opt := range opts {
		opt(qrg)
	}

	return qrg, nil
}

// Option configures QRCodeGenerator.
type Option func(*QRCodeGenerator)

// WithEndDeviceFormat configures QRCodeGenerator with an EndDeviceFormat.
func WithEndDeviceFormat(id string, f enddevices.Format) Option {
	return func(qrg *QRCodeGenerator) {
		qrg.endDevices.RegisterEndDeviceFormat(id, f)
	}
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
