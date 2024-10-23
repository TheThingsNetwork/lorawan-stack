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
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blocklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	evtCreateEndDevice = events.Define(
		"end_device.create", "create end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtUpdateEndDevice = events.Define(
		"end_device.update", "update end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtDeleteEndDevice = events.Define(
		"end_device.delete", "delete end device",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
	evtBatchDeleteEndDevices = events.Define(
		"end_device.batch.delete", "batch delete end devices",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
		events.WithDataType(&ttnpb.EndDeviceIdentifiersList{}),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
		events.WithPropagateToParent(),
	)
)

var errEndDeviceEUIsTaken = errors.DefineAlreadyExists(
	"end_device_euis_taken",
	"an end device with JoinEUI `{join_eui}` and DevEUI `{dev_eui}` is already registered as `{device_id}` in application `{application_id}`",
)

func getHost(address string) string {
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err == nil {
			address = u.Host
		}
	}
	if strings.Contains(address, ":") {
		host, _, err := net.SplitHostPort(address)
		if err == nil {
			return host
		}
	}
	return address
}

var endDeviceAuthenticationCodeSeparator = ":"

var (
	errNetworkServerAddressMismatch = errors.DefineInvalidArgument(
		"network_server_address_mismatch",
		"network server address `{address}` does not match `{expected}`",
	)
	errApplicationServerAddressMismatch = errors.DefineInvalidArgument(
		"application_server_address_mismatch",
		"application server address `{address}` does not match `{expected}`",
	)
	errJoinServerAddressMismatch = errors.DefineInvalidArgument(
		"join_server_address_mismatch",
		"join server address `{address}` does not match `{expected}`",
	)
)

func (is *IdentityServer) validateEndDeviceServerAddressMatch(ctx context.Context, dev *ttnpb.EndDevice) error {
	if dev.NetworkServerAddress == "" && dev.ApplicationServerAddress == "" && dev.JoinServerAddress == "" {
		return nil
	}
	var app *ttnpb.Application
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		app, err = st.GetApplication(ctx, dev.GetIds().GetApplicationIds(), store.FieldMask{
			"network_server_address",
			"application_server_address",
			"join_server_address",
		})
		return err
	})
	if err != nil {
		return err
	}
	if app.NetworkServerAddress != "" && dev.NetworkServerAddress != "" &&
		getHost(app.NetworkServerAddress) != getHost(dev.NetworkServerAddress) {
		return errNetworkServerAddressMismatch.WithAttributes(
			"address", dev.NetworkServerAddress,
			"expected", app.NetworkServerAddress,
		)
	}
	if app.ApplicationServerAddress != "" && dev.ApplicationServerAddress != "" &&
		getHost(app.ApplicationServerAddress) != getHost(dev.ApplicationServerAddress) {
		return errApplicationServerAddressMismatch.WithAttributes(
			"address", dev.ApplicationServerAddress,
			"expected", app.ApplicationServerAddress,
		)
	}
	if app.JoinServerAddress != "" && dev.JoinServerAddress != "" &&
		getHost(app.JoinServerAddress) != getHost(dev.JoinServerAddress) {
		return errJoinServerAddressMismatch.WithAttributes(
			"address", dev.JoinServerAddress,
			"expected", app.JoinServerAddress,
		)
	}
	return nil
}

func (is *IdentityServer) createEndDevice(ctx context.Context, req *ttnpb.CreateEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, req.EndDevice.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if err = blocklist.Check(ctx, req.EndDevice.Ids.DeviceId); err != nil {
		return nil, err
	}

	if err := is.validateEndDeviceServerAddressMatch(ctx, req.EndDevice); err != nil {
		return nil, err
	}

	if req.EndDevice.Picture != nil {
		if err = is.processEndDevicePicture(ctx, req.EndDevice); err != nil {
			return nil, err
		}
	}
	defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()

	// Store plaintext value to return in the response to clients.
	var ptCACSecret string

	// Create the CAC only if the value is set.
	if req.EndDevice.GetClaimAuthenticationCode().GetValue() != "" {
		ptCACSecret = req.EndDevice.ClaimAuthenticationCode.Value
		if err = validateEndDeviceAuthenticationCode(req.EndDevice.ClaimAuthenticationCode); err != nil {
			return nil, err
		}
		if is.config.EndDevices.EncryptionKeyID != "" {
			encrypted, err := is.KeyService().Encrypt(
				ctx,
				[]byte(req.EndDevice.ClaimAuthenticationCode.Value),
				is.config.EndDevices.EncryptionKeyID,
			)
			if err != nil {
				return nil, err
			}
			// Store the encrypted value along with the ID of the key used to encrypt it.
			req.EndDevice.ClaimAuthenticationCode.Value = fmt.Sprintf(
				"%s%s%s",
				is.config.EndDevices.EncryptionKeyID,
				endDeviceAuthenticationCodeSeparator,
				hex.EncodeToString(encrypted),
			)
		} else {
			log.FromContext(ctx).Debug(
				"No encryption key defined, store end device claim authentication code directly in plaintext",
			)
		}
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		dev, err = st.CreateEndDevice(ctx, req.EndDevice)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		joinEUI := types.MustEUI64(req.EndDevice.Ids.JoinEui)
		devEUI := types.MustEUI64(req.EndDevice.Ids.DevEui)
		if errors.IsAlreadyExists(err) && errors.Resemble(err, storeutil.ErrEUITaken) {
			if ids, err := is.getEndDeviceIdentifiersForEUIs(ctx, &ttnpb.GetEndDeviceIdentifiersForEUIsRequest{
				JoinEui: req.EndDevice.Ids.JoinEui,
				DevEui:  req.EndDevice.Ids.DevEui,
			}); err == nil {
				return nil, errEndDeviceEUIsTaken.WithAttributes(
					"join_eui", joinEUI.String(),
					"dev_eui", devEUI.String(),
					"device_id", ids.GetDeviceId(),
					"application_id", ids.GetApplicationIds().GetApplicationId(),
				)
			}
		}
		return nil, err
	}
	if ptCACSecret != "" {
		dev.ClaimAuthenticationCode.Value = ptCACSecret
	}
	events.Publish(evtCreateEndDevice.NewWithIdentifiersAndData(ctx, req.EndDevice.Ids, nil))
	return dev, nil
}

func (is *IdentityServer) getEndDevice(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (dev *ttnpb.EndDevice, err error) {
	if err = rights.RequireApplication(ctx, req.EndDeviceIds.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}

	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, getPaths, nil)
	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "picture") {
		defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "claim_authentication_code") {
		req.FieldMask.Paths = ttnpb.AddFields(req.FieldMask.GetPaths(), "claim_authentication_code")
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		dev, err = st.GetEndDevice(ctx, req.EndDeviceIds, req.FieldMask.GetPaths())
		return err
	})
	if err != nil {
		return nil, err
	}
	if cacVal := dev.GetClaimAuthenticationCode().GetValue(); cacVal != "" {
		s := strings.Split(cacVal, endDeviceAuthenticationCodeSeparator)
		if len(s) == 2 {
			v, err := hex.DecodeString(s[1])
			if err != nil {
				return nil, err
			}
			value, err := is.KeyService().Decrypt(ctx, v, s[0])
			if err != nil {
				return nil, err
			}
			dev.ClaimAuthenticationCode.Value = string(value)
		} else {
			log.FromContext(ctx).Debug("No encryption key defined, return stored end device claim authentication code")
		}
	}
	return dev, nil
}

func (is *IdentityServer) getEndDeviceIdentifiersForEUIs(ctx context.Context, req *ttnpb.GetEndDeviceIdentifiersForEUIsRequest) (ids *ttnpb.EndDeviceIdentifiers, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		dev, err := st.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			JoinEui: req.JoinEui,
			DevEui:  req.DevEui,
		}, []string{"ids.application_ids.application_id", "ids.device_id", "ids.join_eui", "ids.dev_eui"})
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
	} else if err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, getPaths, nil)

	if req.Filters != nil {
		for _, filter := range req.Filters {
			if _, ok := filter.GetField().(*ttnpb.ListEndDevicesRequest_Filter_UpdatedSince); ok {
				ctx = store.WithFilter(ctx, "updated_at", filter.GetUpdatedSince().AsTime().Format(time.RFC3339Nano))
			}
		}
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	devs = &ttnpb.EndDevices{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		devs.EndDevices, err = st.ListEndDevices(ctx, req.GetApplicationIds(), req.FieldMask.GetPaths())
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
		req.FieldMask = cleanFieldMaskPaths([]string{"activated_at", "locations", "last_seen_at"}, req.FieldMask, nil, getPaths)
	} else if err = rights.RequireApplication(ctx, req.EndDevice.Ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask(updatePaths...)
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "activated_at") && req.EndDevice.ActivatedAt == nil {
		// The end device activation state may not be unset once set.
		req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, nil, []string{"activated_at"})
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "claim_authentication_code") {
		req.FieldMask.Paths = ttnpb.AddFields(req.FieldMask.GetPaths(), "claim_authentication_code")
	}

	if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "picture") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "picture") {
			req.FieldMask.Paths = ttnpb.AddFields(req.FieldMask.GetPaths(), "picture")
		}
		if req.EndDevice.Picture != nil {
			if err = is.processEndDevicePicture(ctx, req.EndDevice); err != nil {
				return nil, err
			}
		}
		defer func() { is.setFullEndDevicePictureURL(ctx, dev) }()
	}

	// Store plaintext value to return in the response to clients.
	var ptCACSecret string

	if ttnpb.HasAnyField(
		req.FieldMask.GetPaths(),
		"claim_authentication_code",
	) && req.EndDevice.ClaimAuthenticationCode != nil {
		if err = validateEndDeviceAuthenticationCode(req.EndDevice.ClaimAuthenticationCode); err != nil {
			return nil, err
		}
		if req.EndDevice.ClaimAuthenticationCode.Value != "" {
			if is.config.EndDevices.EncryptionKeyID != "" {
				ptCACSecret = req.EndDevice.ClaimAuthenticationCode.Value
				encrypted, err := is.KeyService().Encrypt(
					ctx,
					[]byte(req.EndDevice.ClaimAuthenticationCode.Value),
					is.config.EndDevices.EncryptionKeyID,
				)
				if err != nil {
					return nil, err
				}
				// Store the encrypted value along with the ID of the key used to encrypt it.
				req.EndDevice.ClaimAuthenticationCode.Value = fmt.Sprintf(
					"%s%s%s",
					is.config.EndDevices.EncryptionKeyID,
					endDeviceAuthenticationCodeSeparator,
					hex.EncodeToString(encrypted),
				)
			} else {
				log.FromContext(ctx).Debug(
					"No encryption key defined, store end device claim authentication code directly in plaintext",
				)
			}
		}
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		dev, err = st.UpdateEndDevice(ctx, req.EndDevice, req.FieldMask.GetPaths())
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateEndDevice.NewWithIdentifiersAndData(ctx, req.EndDevice.Ids, req.FieldMask.GetPaths()))

	if ptCACSecret != "" {
		dev.ClaimAuthenticationCode.Value = ptCACSecret
	}

	return dev, nil
}

func (is *IdentityServer) batchUpdateEndDeviceLastSeen(ctx context.Context, req *ttnpb.BatchUpdateEndDeviceLastSeenRequest) (*emptypb.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	if len(req.Updates) == 0 {
		return ttnpb.Empty, nil
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		return st.BatchUpdateEndDeviceLastSeen(ctx, req.Updates)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) deleteEndDevice(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*emptypb.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		if err := st.PurgeEntityBookmarks(ctx, ids.GetEntityIdentifiers()); err != nil {
			return err
		}
		return st.DeleteEndDevice(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteEndDevice.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func validateEndDeviceAuthenticationCode(authCode *ttnpb.EndDeviceAuthenticationCode) error {
	if validFrom, validTo := ttnpb.StdTime(authCode.ValidFrom), ttnpb.StdTime(authCode.ValidTo); validFrom != nil &&
		validTo != nil {
		if validTo.Before(*validFrom) || authCode.Value == "" {
			return errClaimAuthenticationCode.New()
		}
	}
	return nil
}

type endDeviceRegistry struct {
	ttnpb.UnimplementedEndDeviceRegistryServer

	*IdentityServer
}

type endDeviceBatchRegistry struct {
	ttnpb.UnimplementedEndDeviceBatchRegistryServer

	*IdentityServer
}

func (is *IdentityServer) batchDeleteEndDevice(
	ctx context.Context,
	req *ttnpb.BatchDeleteEndDevicesRequest,
) (*emptypb.Empty, error) {
	if err := rights.RequireApplication(ctx,
		req.ApplicationIds,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	var (
		err     error
		deleted = []*ttnpb.EndDeviceIdentifiers{}
	)
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		deleted, err = st.BatchDeleteEndDevices(ctx, req.ApplicationIds, req.DeviceIds)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(deleted) != 0 {
		events.Publish(
			evtBatchDeleteEndDevices.NewWithIdentifiersAndData(
				ctx, req.ApplicationIds, &ttnpb.EndDeviceIdentifiersList{
					EndDeviceIds: deleted,
				},
			),
		)
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) batchGetEndDevices(
	ctx context.Context,
	req *ttnpb.BatchGetEndDevicesRequest,
) (*ttnpb.EndDevices, error) {
	err := rights.RequireApplication(ctx,
		req.ApplicationIds,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	)
	if err != nil {
		return nil, err
	}

	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, getPaths, nil)

	res := &ttnpb.EndDevices{}
	ids := make([]*ttnpb.EndDeviceIdentifiers, 0, len(req.DeviceIds))
	for _, id := range req.DeviceIds {
		ids = append(ids, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: req.ApplicationIds,
			DeviceId:       id,
		})
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		res.EndDevices, err = st.FindEndDevices(ctx, ids, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, dev := range res.EndDevices {
		if ttnpb.HasAnyField(ttnpb.TopLevelFields(req.FieldMask.GetPaths()), "picture") {
			is.setFullEndDevicePictureURL(ctx, dev)
		}
	}
	return res, nil
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

func (dr *endDeviceRegistry) BatchUpdateLastSeen(ctx context.Context, req *ttnpb.BatchUpdateEndDeviceLastSeenRequest) (*emptypb.Empty, error) {
	return dr.batchUpdateEndDeviceLastSeen(ctx, req)
}

func (dr *endDeviceRegistry) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*emptypb.Empty, error) {
	return dr.deleteEndDevice(ctx, req)
}

func (reg *endDeviceBatchRegistry) Delete(
	ctx context.Context,
	req *ttnpb.BatchDeleteEndDevicesRequest,
) (*emptypb.Empty, error) {
	return reg.batchDeleteEndDevice(ctx, req)
}

func (reg *endDeviceBatchRegistry) Get(
	ctx context.Context,
	req *ttnpb.BatchGetEndDevicesRequest,
) (*ttnpb.EndDevices, error) {
	return reg.batchGetEndDevices(ctx, req)
}
