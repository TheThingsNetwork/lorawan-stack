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
	"path"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtCreateEndDevice = events.Define(
		"end_device.create", "create end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtUpdateEndDevice = events.Define(
		"end_device.update", "update end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtDeleteEndDevice = events.Define(
		"end_device.delete", "delete end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
)

func (is *IdentityServer) createEndDevice(ctx context.Context, req *ttnpb.CreateEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if err = blacklist.Check(ctx, req.DeviceID); err != nil {
		return nil, err
	}

	if req.EndDevice.Picture != nil {
		if err = is.processEndDevicePicture(ctx, &req.EndDevice); err != nil {
			return nil, err
		}
	}
	defer func() { is.setFullEndDevicePictureURL(ctx, &req.EndDevice) }()

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = store.GetEndDeviceStore(db).CreateEndDevice(ctx, &req.EndDevice)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateEndDevice(ctx, req.EndDeviceIdentifiers, nil))
	return dev, nil
}

func (is *IdentityServer) getEndDevice(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = store.GetEndDeviceStore(db).GetEndDevice(ctx, &req.EndDeviceIdentifiers, &req.FieldMask)
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
		dev, err := store.GetEndDeviceStore(db).GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			JoinEUI: &req.JoinEUI,
			DevEUI:  &req.DevEUI,
		}, &types.FieldMask{Paths: []string{"ids.application_ids.application_id", "ids.device_id", "ids.join_eui", "ids.dev_eui"}})
		if err != nil {
			return err
		}
		ids = &dev.EndDeviceIdentifiers
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) listEndDevices(ctx context.Context, req *ttnpb.ListEndDevicesRequest) (devs *ttnpb.EndDevices, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	devs = &ttnpb.EndDevices{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		devs.EndDevices, err = store.GetEndDeviceStore(db).ListEndDevices(ctx, &req.ApplicationIdentifiers, &req.FieldMask)
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
	if dev != nil && dev.Picture != nil {
		for size, file := range dev.Picture.Sizes {
			if !strings.Contains(file, "://") {
				dev.Picture.Sizes[size] = path.Join(bucketURL, file)
			}
		}
	}
}

func (is *IdentityServer) updateEndDevice(ctx context.Context, req *ttnpb.UpdateEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask.Paths, nil, getPaths)
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = updatePaths
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.Paths), "picture") {
		if !ttnpb.HasAnyField(req.FieldMask.Paths, "picture") {
			req.FieldMask.Paths = append(req.FieldMask.Paths, "picture")
		}
		if req.EndDevice.Picture != nil {
			if err = is.processEndDevicePicture(ctx, &req.EndDevice); err != nil {
				return nil, err
			}
		}
		defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		dev, err = store.GetEndDeviceStore(db).UpdateEndDevice(ctx, &req.EndDevice, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateEndDevice(ctx, req.EndDeviceIdentifiers, req.FieldMask.Paths))
	return dev, nil
}

func (is *IdentityServer) deleteEndDevice(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*types.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetEndDeviceStore(db).DeleteEndDevice(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteEndDevice(ctx, ids, nil))
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

func (dr *endDeviceRegistry) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*types.Empty, error) {
	return dr.deleteEndDevice(ctx, req)
}
