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

// PackageName defines the package name.
const PackageName = "lora-cloud-device-management-v1"

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
func (p *DeviceManagementPackage) HandleUp(ctx context.Context, def *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation, up *ttnpb.ApplicationUp) (err error) {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/packages/loradms/v1")
	logger := log.FromContext(ctx)

	if def == nil && assoc == nil {
		return errNoAssociation.New()
	}

	if up.EndDeviceIds.DevEui == nil || up.EndDeviceIds.DevEui.IsZero() {
		logger.Debug("Package configured for end device with no device EUI")
		return nil
	}

	defer func() {
		if err != nil {
			registerPackageFail(ctx, *up.EndDeviceIds, err)
		}
	}()

	data, fPort, err := p.mergePackageData(def, assoc)
	if err != nil {
		return err
	}

	switch m := up.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		join := m.JoinAccept
		loraUp := &objects.LoRaUplink{
			Type:      objects.JoiningUplinkType,
			Timestamp: float64PtrOfTimestamp(join.ReceivedAt),
		}
		return p.sendUplink(ctx, up, loraUp, data)
	case *ttnpb.ApplicationUp_UplinkMessage:
		msg := m.UplinkMessage
		settings := msg.GetSettings()
		loraUp := &objects.LoRaUplink{
			Type:      objects.UplinkUplinkType,
			FCnt:      uint32Ptr(msg.GetFCnt()),
			Port:      uint8Ptr(uint8(msg.GetFPort())),
			Payload:   hexPtr(objects.Hex(msg.FrmPayload)),
			Freq:      uint32Ptr(uint32(settings.Frequency)),
			Timestamp: float64PtrOfTimestamp(msg.ReceivedAt),
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
	ctx = events.ContextWithCorrelationID(ctx, append(up.CorrelationIds, fmt.Sprintf("as:packages:loraclouddmsv1:%s", events.NewCorrelationID()))...)
	logger := log.FromContext(ctx)
	eui := objects.EUI(*up.EndDeviceIds.DevEui)

	httpClient, err := p.server.HTTPClient(ctx)
	if err != nil {
		logger.WithError(err).Debug("Failed to create HTTP client")
		return err
	}
	client, err := api.New(httpClient, api.WithToken(data.token), api.WithBaseURL(data.serverURL))
	if err != nil {
		logger.WithError(err).Debug("Failed to create API client")
		return err
	}
	resp, err := client.Uplinks.Send(ctx, objects.DeviceUplinks{
		eui: loraUp,
	})
	if err != nil {
		logger.WithError(err).Debug("Failed to send uplink upstream")
		return err
	}
	logger.Debug("Uplink sent to the Device Management Service")

	response, ok := resp[eui]
	if !ok {
		return errDeviceEUIMissing.WithAttributes("dev_eui", up.EndDeviceIds.DevEui)
	}
	if response.Error != "" {
		return errUplinkRequestFailed.WithCause(errors.New(response.Error))
	}

	result := response.Result
	resultStruct, err := toStruct(result.Raw)
	if err != nil {
		return err
	}

	if err := p.sendDownlink(ctx, up.EndDeviceIds, result.Downlink, data); err != nil {
		return err
	}

	if err := p.sendServiceData(ctx, up.EndDeviceIds, resultStruct); err != nil {
		return err
	}

	if err := p.sendLocationSolved(ctx, up.EndDeviceIds, result.Position); err != nil {
		return err
	}

	if err := p.parseStreamRecords(ctx, result.StreamRecords, up, data, loraUp.Timestamp); err != nil {
		return err
	}

	return nil
}

func (p *DeviceManagementPackage) sendDownlink(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, downlink *objects.LoRaDnlink, data *packageData) error {
	if downlink == nil {
		return nil
	}
	// Downlinks that are the result of a location solving query will erroneously arrive
	// on FPort 0. If we know that the device uses the TLV encoding, we can translate the
	// FPort to 150 in order to fix this.
	if downlink.Port == 0 && data.GetUseTLVEncoding() {
		downlink.Port = 150
	}
	return p.server.DownlinkQueuePush(ctx, ids, []*ttnpb.ApplicationDownlink{{
		FPort:      uint32(downlink.Port),
		FrmPayload: []byte(downlink.Payload),
	}})
}

func (p *DeviceManagementPackage) sendServiceData(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, data *types.Struct) error {
	return p.server.Publish(ctx, &ttnpb.ApplicationUp{
		EndDeviceIds:   ids,
		CorrelationIds: events.CorrelationIDsFromContext(ctx),
		ReceivedAt:     ttnpb.ProtoTimePtr(time.Now()),
		Up: &ttnpb.ApplicationUp_ServiceData{
			ServiceData: &ttnpb.ApplicationServiceData{
				Data:    data,
				Service: PackageName,
			},
		},
	})
}

func (p *DeviceManagementPackage) sendLocationSolved(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, position *objects.PositionSolution) error {
	if position == nil {
		return nil
	}
	if len(position.LLH) != 3 {
		log.FromContext(ctx).WithField("len", len(position.LLH)).Warn("Invalid LLH length")
		return nil
	}
	source := ttnpb.LocationSource_SOURCE_UNKNOWN
	switch position.Algorithm {
	case objects.GNSSPositionSolutionType:
		source = ttnpb.LocationSource_SOURCE_GPS
	case objects.WiFiPositionSolutionType:
		source = ttnpb.LocationSource_SOURCE_WIFI_RSSI_GEOLOCATION
	}
	return p.server.Publish(ctx, &ttnpb.ApplicationUp{
		EndDeviceIds:   ids,
		CorrelationIds: events.CorrelationIDsFromContext(ctx),
		ReceivedAt:     ttnpb.ProtoTimePtr(time.Now()),
		Up: &ttnpb.ApplicationUp_LocationSolved{
			LocationSolved: &ttnpb.ApplicationLocation{
				Service: fmt.Sprintf("%v-%s", PackageName, position.Algorithm),
				Location: &ttnpb.Location{
					Latitude:  position.LLH[0],
					Longitude: position.LLH[1],
					Altitude:  int32(position.LLH[2]),
					Accuracy:  int32(position.Accuracy),
					Source:    source,
				},
			},
		},
	})
}

const tlvWiFiHeaderLength = 5

func (p *DeviceManagementPackage) parseStreamRecords(ctx context.Context, records []objects.StreamRecord, up *ttnpb.ApplicationUp, data *packageData, originalTimestamp *float64) error {
	if records == nil || !data.GetUseTLVEncoding() {
		return nil
	}
	f := func(tag uint8, bytes []byte) error {
		loraUp := &objects.LoRaUplink{
			Timestamp: originalTimestamp,
		}
		switch tag {
		case 0x05, 0x06, 0x07: // GNSS data
			payload := objects.Hex(bytes)
			loraUp.Type = objects.GNSSUplinkType
			loraUp.Payload = &payload
		case 0x0E: // WiFi data
			if len(bytes) < tlvWiFiHeaderLength {
				return nil
			}
			bytes = bytes[tlvWiFiHeaderLength:]
			fallthrough
		case 0x08: // Legacy WiFi data
			payload := append(objects.Hex{0x01}, bytes...)
			loraUp.Type = objects.WiFiUplinkType
			loraUp.Payload = &payload
		default:
			return nil
		}
		return p.sendUplink(ctx, up, loraUp, data)
	}
	for _, record := range records {
		if err := parseTLVPayload(record.Data, f); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to parse TLV record")
			continue
		}
	}
	return nil
}

func (p *DeviceManagementPackage) mergePackageData(def *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation) (*packageData, uint32, error) {
	var defaultData, associationData packageData
	var fPort uint32
	if def != nil {
		if err := defaultData.fromStruct(def.Data); err != nil {
			return nil, 0, err
		}
		fPort = def.Ids.FPort
	}
	if assoc != nil {
		if err := associationData.fromStruct(assoc.Data); err != nil {
			return nil, 0, err
		}
		fPort = assoc.Ids.FPort
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
		if data.useTLVEncoding != nil {
			merged.useTLVEncoding = data.useTLVEncoding
		}
	}
	if merged.serverURL == nil {
		merged.serverURL = urlutil.CloneURL(api.DefaultServerURL)
	}
	return &merged, fPort, nil
}

// Package implements packages.ApplicationPackageHandler.
func (p *DeviceManagementPackage) Package() *ttnpb.ApplicationPackage {
	return &ttnpb.ApplicationPackage{
		Name:         PackageName,
		DefaultFPort: 199,
	}
}

// New instantiates the LoRa Cloud Device Management package.
func New(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
	return &DeviceManagementPackage{
		server:   server,
		registry: registry,
	}
}

func uint8Ptr(x uint8) *uint8 {
	return &x
}

func uint32Ptr(x uint32) *uint32 {
	return &x
}

func float64PtrOfTimestamp(x *types.Timestamp) *float64 {
	if x == nil {
		return nil
	}
	f := float64(ttnpb.StdTime(x).Unix())
	return &f
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
