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

package identityserver

import (
	"context"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blacklist"
	gormstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateEndDevice = events.Define(
		"end_device.create", "create end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateEndDevice = events.Define(
		"end_device.update", "update end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteEndDevice = events.Define(
		"end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var errEndDeviceEUIsTaken = errors.DefineAlreadyExists(
	"end_device_euis_taken",
	"an end device with JoinEUI `{join_eui}` and DevEUI `{dev_eui}` is already registered as `{device_id}` in application `{application_id}`",
)

func (is *IdentityServer) createEndDevice(ctx context.Context, req *ttnpb.CreateEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, *req.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if err = blacklist.Check(ctx, req.Ids.DeviceId); err != nil {
		return nil, err
	}

	if req.EndDevice.Picture != nil {
		if err = is.processEndDevicePicture(ctx, &req.EndDevice); err != nil {
			return nil, err
		}
	}
	defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = gormstore.GetEndDeviceStore(db).CreateEndDevice(ctx, &req.EndDevice)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.IsAlreadyExists(err) && errors.Resemble(err, gormstore.ErrEUITaken) {
			if ids, err := is.getEndDeviceIdentifiersForEUIs(ctx, &ttnpb.GetEndDeviceIdentifiersForEUIsRequest{
				JoinEui: *req.Ids.JoinEui,
				DevEui:  *req.Ids.DevEui,
			}); err == nil {
				return nil, errEndDeviceEUIsTaken.WithAttributes(
					"join_eui", req.Ids.JoinEui.String(),
					"dev_eui", req.Ids.DevEui.String(),
					"device_id", ids.GetDeviceId(),
					"application_id", ids.GetApplicationIds().GetApplicationId(),
				)
			}
		}
		return nil, err
	}
	events.Publish(evtCreateEndDevice.NewWithIdentifiersAndData(ctx, req.Ids, nil))
	return dev, nil
}

func (is *IdentityServer) getEndDevice(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, *req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}

	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, getPaths, nil)
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "picture") {
		defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = gormstore.GetEndDeviceStore(db).GetEndDevice(ctx, req.EndDeviceIds, req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func (is *IdentityServer) getEndDeviceIdentifiersForEUIs(ctx context.Context, req *ttnpb.GetEndDeviceIdentifiersForEUIsRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err := gormstore.GetEndDeviceStore(db).GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			JoinEui: &req.JoinEui,
			DevEui:  &req.DevEui,
		}, &pbtypes.FieldMask{Paths: []string{"ids.application_ids.application_id", "ids.device_id", "ids.join_eui", "ids.dev_eui"}})
		if err != nil {
			return err
		}
		ids = dev.Ids
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) listEndDevices(ctx context.Context, req *ttnpb.ListEndDevicesRequest) (devs *ttnpb.EndDevices, err error) {
	// If nil identifiers passed, check that the request came from the cluster.
	if req.GetApplicationIds() == nil {
		if err = clusterauth.Authorized(ctx); err != nil {
			return nil, err
		}
		req.FieldMask = cleanFieldMaskPaths([]string{"ids"}, req.FieldMask, nil, []string{"created_at", "updated_at"})
	} else if err = rights.RequireApplication(ctx, *req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, getPaths, nil)
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	devs = &ttnpb.EndDevices{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		devs.EndDevices, err = gormstore.GetEndDeviceStore(db).ListEndDevices(ctx, req.GetApplicationIds(), req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return devs, nil
}

func (is *IdentityServer) setFullEndDevicePictureURL(ctx context.Context, dev *ttnpb.EndDevice) {
	bucketURL := is.configFromContext(ctx).EndDevicePicture.BucketURL
	if bucketURL == "" {
		return
	}
	bucketURL = strings.TrimSuffix(bucketURL, "/") + "/"
	if dev != nil && dev.Picture != nil {
		for size, file := range dev.Picture.Sizes {
			if !strings.Contains(file, "://") {
				dev.Picture.Sizes[size] = bucketURL + strings.TrimPrefix(file, "/")
			}
		}
	}
}

func (is *IdentityServer) updateEndDevice(ctx context.Context, req *ttnpb.UpdateEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if clusterauth.Authorized(ctx) == nil {
		req.FieldMask = cleanFieldMaskPaths([]string{"activated_at", "locations"}, req.FieldMask, nil, getPaths)
	} else if err = rights.RequireApplication(ctx, *req.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "activated_at") && req.ActivatedAt == nil {
		// The end device activation state may not be unset once set.
		req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, nil, []string{"activated_at"})
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "picture") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "picture") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "picture")
		}
		if req.EndDevice.Picture != nil {
			if err = is.processEndDevicePicture(ctx, &req.EndDevice); err != nil {
				return nil, err
			}
		}
		defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = gormstore.GetEndDeviceStore(db).UpdateEndDevice(ctx, &req.EndDevice, req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, req.Ids, req.FieldMask.GetPaths()))
	return dev, nil
}

func (is *IdentityServer) deleteEndDevice(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return gormstore.GetEndDeviceStore(db).DeleteEndDevice(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteEndDevice.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

type endDeviceRegistry struct {
	*IdentityServer
}

func (dr *endDeviceRegistry) Create(ctx context.Context, req *ttnpb.CreateEndDeviceRequest) (*ttnpb.EndDevice, error) {
	return dr.createEndDevice(ctx, req)
}

func (dr *endDeviceRegistry) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	return dr.getEndDevice(ctx, req)
}

func (dr *endDeviceRegistry) GetIdentifiersForEUIs(ctx context.Context, req *ttnpb.GetEndDeviceIdentifiersForEUIsRequest) (*ttnpb.EndDeviceIdentifiers, error) {
	return dr.getEndDeviceIdentifiersForEUIs(ctx, req)
}

func (dr *endDeviceRegistry) List(ctx context.Context, req *ttnpb.ListEndDevicesRequest) (*ttnpb.EndDevices, error) {
	return dr.listEndDevices(ctx, req)
}

func (dr *endDeviceRegistry) Update(ctx context.Context, req *ttnpb.UpdateEndDeviceRequest) (*ttnpb.EndDevice, error) {
	return dr.updateEndDevice(ctx, req)
}

func (dr *endDeviceRegistry) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	return dr.deleteEndDevice(ctx, req)
}
