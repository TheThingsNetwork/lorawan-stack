// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package loracloudgeolocationv3

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loragls/v3/api"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	urlutil "go.thethings.network/lorawan-stack/v3/pkg/util/url"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// PackageName defines the package name.
const PackageName = "lora-cloud-geolocation-v3"

// GeolocationPackage is the LoRa Cloud Geolocation application package.
type GeolocationPackage struct {
	server   io.Server
	registry packages.Registry
}

// RegisterServices implements packages.ApplicationPackageHandler.
func (p *GeolocationPackage) RegisterServices(s *grpc.Server) {}

// RegisterHandlers implements packages.ApplicationPackageHandler.
func (p *GeolocationPackage) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {}

var (
	errLocationQuery = errors.DefineInternal("location_query", "location query")
	errNoResult      = errors.DefineNotFound("no_result", "no location query result")
	errNoAssociation = errors.DefineInternal("no_association", "no association available")
)

// HandleUp implements packages.ApplicationPackageHandler.
func (p *GeolocationPackage) HandleUp(ctx context.Context, def *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation, up *ttnpb.ApplicationUp) (err error) {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/packages/loragls/v1")
	ctx = events.ContextWithCorrelationID(ctx, append(up.CorrelationIDs, fmt.Sprintf("as:packages:loracloudglsv3:%s", events.NewCorrelationID()))...)

	if def == nil && assoc == nil {
		return errNoAssociation.New()
	}

	defer func() {
		if err != nil {
			registerPackageFail(ctx, up.EndDeviceIdentifiers, err)
		}
	}()

	data, err := p.mergePackageData(def, assoc)
	if err != nil {
		return err
	}

	switch m := up.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		return p.sendQuery(ctx, up.EndDeviceIdentifiers, m.UplinkMessage, data)
	default:
		return nil
	}
}

// Package implements packages.ApplicationPackageHandler.
func (p *GeolocationPackage) Package() *ttnpb.ApplicationPackage {
	return &ttnpb.ApplicationPackage{
		Name:         PackageName,
		DefaultFPort: 197,
	}
}

// New instantiates the LoRa Cloud Geolocation package.
func New(server io.Server, registry packages.Registry) packages.ApplicationPackageHandler {
	return &GeolocationPackage{
		server:   server,
		registry: registry,
	}
}

func (p *GeolocationPackage) sendQuery(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, up *ttnpb.ApplicationUplink, data *Data) error {
	logger := log.FromContext(ctx)

	httpClient, err := p.server.HTTPClient(ctx)
	if err != nil {
		return err
	}

	req := api.BuildSingleFrameRequest(up.RxMetadata)
	if len(req.Gateways) < 3 {
		logger.Debug("Not enough gateways available")
		return nil
	}
	client, err := api.New(httpClient, api.WithToken(data.Token), api.WithBaseURL(data.ServerURL))
	if err != nil {
		logger.WithError(err).Debug("Failed to create API client")
		return err
	}
	resp, err := client.SolveSingleFrame(ctx, req)
	if err != nil {
		logger.WithError(err).Debug("Query failed")
		return err
	}
	logger.Debug("Query sent to Geolocation services")

	resultStruct, err := toStruct(resp.Raw)
	if err != nil {
		return err
	}

	if err := p.sendServiceData(ctx, ids, resultStruct); err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		var details []proto.Message
		for _, message := range resp.Errors {
			details = append(details, &ttnpb.ErrorDetails{
				Code:          uint32(codes.Unknown),
				MessageFormat: message,
			})
		}
		return errLocationQuery.WithDetails(details...)
	}

	if resp.Result == nil {
		return errNoResult.New()
	}

	if err := p.sendLocationSolved(ctx, ids, resp.Result); err != nil {
		return err
	}

	return nil
}

func (p *GeolocationPackage) sendServiceData(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, data *types.Struct) error {
	return p.server.Publish(ctx, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: ids,
		CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
		ReceivedAt:           timePtr(time.Now().UTC()),
		Up: &ttnpb.ApplicationUp_ServiceData{
			ServiceData: &ttnpb.ApplicationServiceData{
				Data:    data,
				Service: PackageName,
			},
		},
	})
}

func (p *GeolocationPackage) sendLocationSolved(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, position *api.LocationSolverResult) error {
	if position == nil {
		return nil
	}
	source := ttnpb.SOURCE_UNKNOWN
	switch position.Algorithm {
	case api.Algorithm_RSSI:
		source = ttnpb.SOURCE_LORA_RSSI_GEOLOCATION
	case api.Algorithm_TDOA, api.Algorithm_RSSITDOA:
		source = ttnpb.SOURCE_LORA_TDOA_GEOLOCATION
	}
	location := position.Location
	return p.server.Publish(ctx, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: ids,
		CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
		ReceivedAt:           timePtr(time.Now().UTC()),
		Up: &ttnpb.ApplicationUp_LocationSolved{
			LocationSolved: &ttnpb.ApplicationLocation{
				Service: fmt.Sprintf("%v-%s", PackageName, position.Algorithm),
				Location: ttnpb.Location{
					Latitude:  location.Latitude,
					Longitude: location.Longitude,
					Accuracy:  int32(location.Tolerance),
					Source:    source,
				},
			},
		},
	})
}

func (p *GeolocationPackage) mergePackageData(def *ttnpb.ApplicationPackageDefaultAssociation, assoc *ttnpb.ApplicationPackageAssociation) (*Data, error) {
	var defaultData, associationData Data
	if def != nil {
		if err := defaultData.FromStruct(def.Data); err != nil {
			return nil, err
		}
	}
	if assoc != nil {
		if err := associationData.FromStruct(assoc.Data); err != nil {
			return nil, err
		}
	}
	var merged Data
	for _, data := range []*Data{
		&defaultData,
		&associationData,
	} {
		if data.Query != 0 {
			merged.Query = data.Query
		}
		if data.ServerURL != nil {
			merged.ServerURL = urlutil.CloneURL(data.ServerURL)
		}
		if data.Token != "" {
			merged.Token = data.Token
		}
	}
	if merged.ServerURL == nil {
		merged.ServerURL = urlutil.CloneURL(api.DefaultServerURL)
	}
	return &merged, nil
}

func timePtr(x time.Time) *time.Time {
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
