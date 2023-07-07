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

package applicationserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/internal/registry"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// DeviceRegistry is a store for end devices.
type DeviceRegistry interface {
	// Get returns the end device by its identifiers.
	Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error)
	// Set creates, updates or deletes the end device by its identifiers.
	Set(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
	// Range ranges over the end devices and calls the callback function, until false is returned.
	Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error
	// BatchDelete deletes a batch of end devices.
	BatchDelete(
		ctx context.Context,
		appIDs *ttnpb.ApplicationIdentifiers,
		deviceIDs []string,
	) ([]*ttnpb.EndDeviceIdentifiers, error)
}

type replacedEndDeviceFieldRegistryWrapper struct {
	fields   []registry.ReplacedEndDeviceField
	registry DeviceRegistry
}

func (w replacedEndDeviceFieldRegistryWrapper) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	dev, err := w.registry.Get(ctx, ids, paths)
	if err != nil || dev == nil {
		return dev, err
	}
	for _, d := range replaced {
		d.GetTransform(dev)
	}
	return dev, nil
}

func (w replacedEndDeviceFieldRegistryWrapper) Set(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	dev, err := w.registry.Set(ctx, ids, paths, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			for _, d := range replaced {
				d.GetTransform(dev)
			}
		}
		dev, paths, err := f(dev)
		if err != nil || dev == nil {
			return dev, paths, err
		}
		for _, d := range replaced {
			if ttnpb.HasAnyField(paths, d.Old) {
				paths = ttnpb.AddFields(paths, d.New)
			}
			d.SetTransform(dev, d.MatchedOld, d.MatchedNew)
		}
		return dev, paths, nil
	})
	if err != nil || dev == nil {
		return dev, err
	}
	for _, d := range replaced {
		d.GetTransform(dev)
	}
	return dev, nil
}

func (w replacedEndDeviceFieldRegistryWrapper) BatchDelete(
	ctx context.Context,
	appIDs *ttnpb.ApplicationIdentifiers,
	deviceIDs []string,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	return w.registry.BatchDelete(ctx, appIDs, deviceIDs)
}

func (w replacedEndDeviceFieldRegistryWrapper) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	return w.registry.Range(ctx, paths, func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, dev *ttnpb.EndDevice) bool {
		if dev != nil {
			for _, d := range replaced {
				d.GetTransform(dev)
			}
		}
		return f(ctx, ids, dev)
	})
}

func wrapEndDeviceRegistryWithReplacedFields(r DeviceRegistry, fields ...registry.ReplacedEndDeviceField) DeviceRegistry {
	return replacedEndDeviceFieldRegistryWrapper{
		fields:   fields,
		registry: r,
	}
}

var errInvalidFieldValue = errors.DefineInvalidArgument("field_value", "invalid value of field `{field}`")

var replacedEndDeviceFields = []registry.ReplacedEndDeviceField{
	{
		Old: "skip_payload_crypto",
		New: "skip_payload_crypto_override",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.SkipPayloadCryptoOverride == nil && dev.SkipPayloadCrypto {
				dev.SkipPayloadCryptoOverride = &wrapperspb.BoolValue{Value: true}
			} else {
				dev.SkipPayloadCrypto = dev.SkipPayloadCryptoOverride.GetValue()
			}
		},
		SetTransform: func(dev *ttnpb.EndDevice, useOld, useNew bool) error {
			if useOld {
				if useNew {
					if dev.SkipPayloadCrypto != dev.SkipPayloadCryptoOverride.GetValue() {
						return errInvalidFieldValue.WithAttributes("field", "skip_payload_crypto")
					}
				} else {
					dev.SkipPayloadCryptoOverride = &wrapperspb.BoolValue{Value: dev.SkipPayloadCrypto}
				}
			}
			dev.SkipPayloadCrypto = false
			return nil
		},
	},
}

// LinkRegistry is a store for application links.
type LinkRegistry interface {
	// Get returns the link by the application identifiers.
	Get(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error)
	// Range ranges the links and calls the callback function, until false is returned.
	Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error
	// Set creates, updates or deletes the link by the application identifiers.
	Set(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error)
}

// ApplicationUplinkRegistry is a store for uplink messages.
type ApplicationUplinkRegistry interface {
	// Range ranges the uplink messagess and calls the callback function, until false is returned.
	Range(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string, f func(context.Context, *ttnpb.ApplicationUplink) bool) error
	// Push pushes the provided uplink message to the storage.
	Push(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, up *ttnpb.ApplicationUplink) error
	// Clear empties the uplink messages storage by the end device identifiers.
	Clear(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) error
	// BatchClear empties the uplink messages storage of multiple end devices.
	BatchClear(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error
}
