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
	"fmt"
	"net/http"
	"net/url"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api/objects"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	urlutil "go.thethings.network/lorawan-stack/v3/pkg/util/url"
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
	errDeviceEUIMissing    = errors.DefineNotFound("device_eui_missing", "device EUI `{dev_eui}` not found")
	errUplinkRequestFailed = errors.DefineInternal("uplink_request_failed", "uplink request failed")
	errNoAssociation       = errors.DefineInternal("no_association", "no association available")
)

// HandleUp implements packages.ApplicationPackageHandler.
func (p *DeviceManagementPackage) HandleUp(ctx context.Context, def *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation, up *ttnpb.ApplicationUp) error {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/packages/loradms/v1")
	logger := log.FromContext(ctx)

	if def == nil && assoc == nil {
		return errNoAssociation.New()
	}

	if up.DevEUI == nil || up.DevEUI.IsZero() {
		logger.Debug("Package configured for end device with no device EUI")
		return nil
	}

	data, fPort, err := p.mergePackageData(def, assoc)
	if err != nil {
		return err
	}

	switch m := up.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		return p.handleJoinAccept(ctx, data, up.EndDeviceIdentifiers, m.JoinAccept)
	case *ttnpb.ApplicationUp_UplinkMessage:
		return p.handleUplinkMessage(ctx, data, fPort, up.EndDeviceIdentifiers, m.UplinkMessage)
	default:
		return nil
	}
}

func (p *DeviceManagementPackage) handleJoinAccept(ctx context.Context, data *packageData, ids ttnpb.EndDeviceIdentifiers, join *ttnpb.ApplicationJoinAccept) error {
	return nil
}

func (p *DeviceManagementPackage) handleUplinkMessage(ctx context.Context, data *packageData, fPort uint32, ids ttnpb.EndDeviceIdentifiers, up *ttnpb.ApplicationUplink) error {
	eui := objects.EUI(*ids.DevEUI)
	uplink := objects.LoRaUplink{
		FCnt:      up.GetFCnt(),
		Port:      uint8(up.GetFPort()),
		Payload:   objects.Hex(up.FRMPayload),
		DR:        uint8(up.GetSettings().DataRateIndex),
		Freq:      uint32(up.GetSettings().Frequency),
		Timestamp: float64(up.ReceivedAt.UTC().Unix()),
	}

	logger := log.FromContext(ctx)
	if fPort != up.FPort {
		logger.Debug("Uplink received on unhandled FPort; drop payload")
		uplink.Payload = nil
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
		return errDeviceEUIMissing.WithAttributes("dev_eui", ids.DevEUI)
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
	err = p.server.DownlinkQueuePush(ctx, ids, []*ttnpb.ApplicationDownlink{down})
	if err != nil {
		logger.WithError(err).Debug("Failed to push downlink to device")
		return err
	}
	logger.Debug("Device Management Service downlink scheduled")

	return nil
}

const defaultServerURL = "https://das.loracloud.com/api/v1"

var parsedDefaultServerURL *url.URL

func (p *DeviceManagementPackage) mergePackageData(def *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation) (*packageData, uint32, error) {
	var defaultData, associationData packageData
	var fPort uint32
	if def != nil {
		if err := defaultData.fromStruct(def.Data); err != nil {
			return nil, 0, err
		}
		fPort = def.FPort
	}
	if assoc != nil {
		if err := associationData.fromStruct(assoc.Data); err != nil {
			return nil, 0, err
		}
		fPort = assoc.FPort
	}
	var merged packageData
	for _, data := range []*packageData{
		&defaultData,
		&associationData,
	} {
		if merged.serverURL == nil {
			merged.serverURL = urlutil.CloneURL(data.serverURL)
		}
		if merged.token == "" {
			merged.token = data.token
		}
	}
	if merged.serverURL == nil {
		merged.serverURL = urlutil.CloneURL(parsedDefaultServerURL)
	}
	return &merged, fPort, nil
}

func init() {
	packages.RegisterPackage(ttnpb.ApplicationPackage{
		Name:         "lora-cloud-device-management-v1",
		DefaultFPort: 200,
	}, packages.CreateApplicationPackage(
		func(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
			return &DeviceManagementPackage{server, registry}
		},
	))

	var err error
	parsedDefaultServerURL, err = url.Parse(defaultServerURL)
	if err != nil {
		panic(fmt.Sprintf("loradms: failed to parse base URL: %v", err))
	}
}
