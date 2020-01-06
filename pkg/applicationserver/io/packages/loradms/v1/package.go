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

package loraclouddevicemanagementv1

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/loradms/v1/api"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/loradms/v1/api/objects"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// DeviceManagementPackage is the LoRa Cloud Device Management application package.
type DeviceManagementPackage struct {
	server   io.Server
	registry packages.Registry
}

// RegisterServices implements packages.ApplicationPackageHandler.
func (p *DeviceManagementPackage) RegisterServices(s *grpc.Server) {}

// RegisterHandlers implements packages.ApplicationPackageHandler.
func (p *DeviceManagementPackage) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {}

var (
	errDeviceEUIMissing    = errors.DefineNotFound("device_eui_missing", "device EUI `{devEUI}` not found")
	errUplinkRequestFailed = errors.DefineInternal("uplink_request_failed", "uplink request failed")
)

// HandleUp implements packages.ApplicationPackageHandler.
func (p *DeviceManagementPackage) HandleUp(ctx context.Context, assoc *ttnpb.ApplicationPackageAssociation, up *ttnpb.ApplicationUp) error {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/packages/loradms/v1")
	logger := log.FromContext(ctx)

	eui := objects.EUI(*up.DevEUI)
	message := up.GetUplinkMessage()
	settings := message.GetSettings()
	uplink := objects.LoRaUplink{
		FCnt:      message.GetFCnt(),
		Port:      uint8(message.GetFPort()),
		Payload:   objects.Hex(message.FRMPayload),
		DR:        uint8(settings.DataRateIndex),
		Freq:      uint32(settings.GetFrequency()),
		Timestamp: float64(settings.Timestamp),
	}

	var data packageData
	err := data.fromStruct(assoc.Data)
	if errors.IsNotFound(err) {
		// If there is no package data available, just reset the data.
		data = packageData{}
		assoc, err = p.registry.Set(ctx, assoc.ApplicationPackageAssociationIdentifiers, []string{"data"},
			func(assoc *ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error) {
				assoc.Data = data.toStruct()
				return assoc, []string{"data"}, nil
			},
		)
		if err != nil {
			logger.WithError(err).Debug("Failed to update package data")
			return err
		}
	} else if err != nil {
		logger.WithError(err).Debug("Failed to parse package data")
		return err
	}

	client, err := api.New(http.DefaultClient, api.WithToken(data.token))
	if err != nil {
		logger.WithError(err).Debug("Failed to create API client")
		return err
	}
	resp, err := client.Uplinks.Send(objects.DeviceUplinks{
		eui: uplink,
	})
	if err != nil {
		logger.WithError(err).Debug("Failed to send uplink upstream")
		return err
	}
	logger.Debug("Uplink sent to the Device Management Service")

	response, ok := resp[eui]
	if !ok {
		return errDeviceEUIMissing.WithAttributes("devEUI", up.DevEUI)
	}
	if response.Error != "" {
		return errUplinkRequestFailed.WithCause(errors.New(response.Error))
	}

	downlink := response.Result.Downlink
	if downlink == nil {
		logger.Debug("No downlink to be scheduled from the Device Management Service")
		return nil
	}
	down := &ttnpb.ApplicationDownlink{
		FPort:      uint32(downlink.Port),
		FRMPayload: []byte(downlink.Payload),
	}
	err = p.server.DownlinkQueuePush(ctx, assoc.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink{down})
	if err != nil {
		logger.WithError(err).Debug("Failed to push downlink to device")
		return err
	}
	logger.Debug("Device Management Service downlink scheduled")

	return nil
}

func init() {
	p := ttnpb.ApplicationPackage{
		Name:         "lora-cloud-device-management-v1",
		DefaultFPort: 200,
	}
	packages.RegisterPackage(p, packages.CreateApplicationPackage(
		func(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
			return &DeviceManagementPackage{server, registry}
		},
	))
}
