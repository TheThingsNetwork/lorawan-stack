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
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api/objects"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	urlutil "go.thethings.network/lorawan-stack/v3/pkg/util/url"
	"google.golang.org/grpc"
)

const packageName = "lora-cloud-device-management-v1"

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
		join := m.JoinAccept
		loraUp := &objects.LoRaUplink{
			Type:      objects.JoiningUplinkType,
			Timestamp: float64Ptr(float64(join.ReceivedAt.UTC().Unix())),
		}
		return p.sendUplink(ctx, up, loraUp, data)
	case *ttnpb.ApplicationUp_UplinkMessage:
		msg := m.UplinkMessage
		settings := msg.GetSettings()
		loraUp := &objects.LoRaUplink{
			Type:      objects.UplinkUplinkType,
			FCnt:      uint32Ptr(msg.GetFCnt()),
			Port:      uint8Ptr(uint8(msg.GetFPort())),
			Payload:   hexPtr(objects.Hex(msg.FRMPayload)),
			DR:        uint8Ptr(uint8(settings.DataRateIndex)),
			Freq:      uint32Ptr(uint32(settings.Frequency)),
			Timestamp: float64Ptr(float64(msg.ReceivedAt.UTC().Unix())),
		}
		if fPort != msg.FPort {
			log.FromContext(ctx).Debug("Uplink received on unhandled FPort; drop payload")
			loraUp.Payload = &objects.Hex{}
		}
		return p.sendUplink(ctx, up, loraUp, data)
	default:
		return nil
	}
}

func (p *DeviceManagementPackage) sendUplink(ctx context.Context, up *ttnpb.ApplicationUp, loraUp *objects.LoRaUplink, data *packageData) error {
	logger := log.FromContext(ctx)
	eui := objects.EUI(*up.DevEUI)

	client, err := api.New(http.DefaultClient, api.WithToken(data.token), api.WithBaseURL(data.serverURL))
	if err != nil {
		logger.WithError(err).Debug("Failed to create API client")
		return err
	}
	resp, err := client.Uplinks.Send(objects.DeviceUplinks{
		eui: loraUp,
	})
	if err != nil {
		logger.WithError(err).Debug("Failed to send uplink upstream")
		return err
	}
	logger.Debug("Uplink sent to the Device Management Service")

	response, ok := resp[eui]
	if !ok {
		return errDeviceEUIMissing.WithAttributes("dev_eui", up.DevEUI)
	}
	if response.Error != "" {
		return errUplinkRequestFailed.WithCause(errors.New(response.Error))
	}

	result := response.Result
	resultStruct, err := toStruct(&result)
	if err != nil {
		return err
	}

	ctx = events.ContextWithCorrelationID(ctx, append(up.CorrelationIDs, fmt.Sprintf("as:packages:loradas:%s", events.NewCorrelationID()))...)
	now := time.Now().UTC()
	err = p.server.SendUp(ctx, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: up.EndDeviceIdentifiers,
		CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
		ReceivedAt:           &now,
		Up: &ttnpb.ApplicationUp_ServiceData{
			ServiceData: &ttnpb.ApplicationServiceData{
				Data:    resultStruct,
				Service: packageName,
			},
		},
	})
	if err != nil {
		return err
	}

	downlink := result.Downlink
	if downlink == nil {
		logger.Debug("No downlink to be scheduled from the Device Management Service")
		return nil
	}
	down := &ttnpb.ApplicationDownlink{
		FPort:      uint32(downlink.Port),
		FRMPayload: []byte(downlink.Payload),
	}
	err = p.server.DownlinkQueuePush(ctx, up.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink{down})
	if err != nil {
		logger.WithError(err).Debug("Failed to push downlink to device")
		return err
	}
	logger.Debug("Device Management Service downlink scheduled")

	return nil
}

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
		if data.serverURL != nil {
			merged.serverURL = urlutil.CloneURL(data.serverURL)
		}
		if data.token != "" {
			merged.token = data.token
		}
	}
	if merged.serverURL == nil {
		merged.serverURL = urlutil.CloneURL(api.DefaultServerURL)
	}
	return &merged, fPort, nil
}

func init() {
	packages.RegisterPackage(ttnpb.ApplicationPackage{
		Name:         packageName,
		DefaultFPort: 200,
	}, packages.CreateApplicationPackage(
		func(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
			return &DeviceManagementPackage{server, registry}
		},
	))
}

func uint8Ptr(x uint8) *uint8 {
	return &x
}

func uint32Ptr(x uint32) *uint32 {
	return &x
}

func float64Ptr(x float64) *float64 {
	return &x
}

func hexPtr(x objects.Hex) *objects.Hex {
	return &x
}

func toStruct(i interface{}) (*types.Struct, error) {
	b, err := jsonpb.TTN().Marshal(i)
	if err != nil {
		return nil, err
	}
	var st types.Struct
	err = jsonpb.TTN().Unmarshal(b, &st)
	if err != nil {
		return nil, err
	}
	return &st, nil
}
