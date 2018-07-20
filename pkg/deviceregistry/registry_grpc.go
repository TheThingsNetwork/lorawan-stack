// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package deviceregistry

import (
	"context"
	"fmt"
	"math"
	"path"
	"strconv"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var defaultListCount uint64 = 10

// RegistryRPC implements the device registry gRPC service.
type RegistryRPC struct {
	Interface
	*component.Component

	setDeviceProcessor func(ctx context.Context, create bool, dev *ttnpb.EndDevice, fields ...string) (*ttnpb.EndDevice, []string, error)

	servedComponents []ttnpb.PeerInfo_Role
}

// RPCOption represents RegistryRPC option
type RPCOption func(*RegistryRPC)

// ForComponents takes in parameter the components that this device registry RPC will serve for.
func ForComponents(components ...ttnpb.PeerInfo_Role) RPCOption {
	return func(r *RegistryRPC) { r.servedComponents = append(r.servedComponents, components...) }
}

// WithSetDeviceProcessor sets a function, which checks and processes the device and fields,
// which are about to be passed to SetDevice method of RegistryRPC instance.
// After a successful search, SetDevice passes request context, bool, indicating whether the request will trigger a 'Create' or an 'Update',
// device, which is about to be passed to the underlying registry and converted field paths(if such are specified in the request).
// If nil error is returned by fn, SetDevice passes the device and fields returned to the underlying registry,
// otherwise SetDevice returns the error without modifying the registry.
func WithSetDeviceProcessor(fn func(ctx context.Context, create bool, dev *ttnpb.EndDevice, fields ...string) (*ttnpb.EndDevice, []string, error)) RPCOption {
	return func(r *RegistryRPC) { r.setDeviceProcessor = fn }
}

// NewRPC returns a new instance of RegistryRPC
func NewRPC(c *component.Component, r Interface, opts ...RPCOption) (*RegistryRPC, error) {
	rpc := &RegistryRPC{
		Component: c,
		Interface: r,
	}

	for _, opt := range opts {
		opt(rpc)
	}

	return rpc, nil
}

// setPaginationHeaders sets the pagination headers on the underlying gRPC transport stream present in ctx.
// setPaginationHeaders sets:
// - 'x-total-count' - total count of results.
// - 'link' - as defined in the RFC 5988. It is only set if limit != 0 and url != "".
func setPaginationHeaders(ctx context.Context, url string, limit, page, total uint64) error {
	md := metadata.MD{
		"x-total-count": []string{strconv.FormatUint(total, 10)},
	}
	if limit == 0 || url == "" {
		return grpc.SetHeader(ctx, md)
	}

	linkf := fmt.Sprintf(`<%s?page=%%d&per_page=%d>; rel="%%s"`, url, limit)

	n := uint64(math.Ceil(float64(total) / float64(limit)))
	if n == 0 {
		n = 1
	}

	md["link"] = make([]string, 0, 4)

	md["link"] = append(md["link"], fmt.Sprintf(linkf, 1, "first"))
	if page > 1 && page <= n {
		md["link"] = append(md["link"], fmt.Sprintf(linkf, page-1, "prev"))
	}
	if page < n {
		md["link"] = append(md["link"], fmt.Sprintf(linkf, page+1, "next"))
	}
	md["link"] = append(md["link"], fmt.Sprintf(linkf, n, "last"))

	return grpc.SetHeader(ctx, md)
}

// ListDevices lists devices matching filter in underlying registry.
func (r *RegistryRPC) ListDevices(ctx context.Context, filter *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevices, error) {
	if err := rights.RequireApplication(ctx, filter.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}

	md := rpcmetadata.FromIncomingContext(ctx)

	count := defaultListCount
	if md.Limit > 0 {
		count = md.Limit
	}

	page := md.Page
	if page == 0 {
		page = 1
	}

	offset := uint64(0)
	if page > 1 {
		// Page is 1-indexed
		offset = (page - 1) * count
	}

	eds := make([]*ttnpb.EndDevice, 0, defaultListCount)
	total, err := RangeByIdentifiers(r.Interface, filter, "", count, offset, func(dev *Device) bool {
		eds = append(eds, dev.EndDevice)
		return count == 0 || uint64(len(eds)) < count
	})
	if err != nil {
		return nil, err
	}

	return &ttnpb.EndDevices{EndDevices: eds}, setPaginationHeaders(ctx, path.Join(md.Host, md.URI), count, page, total)
}

// GetDevice returns the device associated with id in underlying registry, if found.
func (r *RegistryRPC) GetDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, id.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}

	dev, err := FindByIdentifiers(r.Interface, id)
	if err != nil {
		return nil, err
	}
	return dev.EndDevice, nil
}

// SetDevice sets the device fields to match those of req.Device in underlying registry.
func (r *RegistryRPC) SetDevice(ctx context.Context, req *ttnpb.SetDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.Device.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}

	var fields []string
	if req.FieldMask != nil {
		fields = gogoproto.GoFieldsPaths(req.FieldMask, req.Device)
	}

	dev, err := FindByIdentifiers(r.Interface, &req.Device.EndDeviceIdentifiers)
	notFound := errors.IsNotFound(err)
	if err != nil && !notFound {
		return nil, err
	}

	setDev := &req.Device
	if r.setDeviceProcessor != nil {
		setDev, fields, err = r.setDeviceProcessor(ctx, notFound, setDev, fields...)
		if err != nil && !errors.IsUnknown(err) {
			return nil, err
		} else if err != nil {
			return nil, errProcessorFailed.WithCause(err)
		}
	}

	if notFound {
		dev, err := r.Interface.Create(setDev, fields...)
		if err != nil {
			return nil, err
		}
		events.Publish(evtCreateDevice(ctx, setDev.EndDeviceIdentifiers, nil))
		return dev.EndDevice, nil
	}

	dev.EndDevice = setDev
	if err = dev.Store(fields...); err != nil {
		return nil, err
	}
	events.Publish(evtUpdateDevice(ctx, setDev.EndDeviceIdentifiers, fields))
	return dev.EndDevice, nil
}

// DeleteDevice deletes the device associated with id from underlying registry.
func (r *RegistryRPC) DeleteDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, id.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}

	dev, err := FindByIdentifiers(r.Interface, id)
	if err != nil {
		return nil, err
	}

	if err = dev.Delete(); err != nil {
		return nil, err
	}
	events.Publish(evtDeleteDevice(ctx, id, nil))
	return ttnpb.Empty, nil
}
