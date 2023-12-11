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

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/internal/registry"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/proto"
)

type UplinkMatch struct {
	ApplicationIdentifiers *ttnpb.ApplicationIdentifiers
	DeviceID               string
	LoRaWANVersion         ttnpb.MACVersion
	FNwkSIntKey            *ttnpb.KeyEnvelope
	LastFCnt               uint32
	ResetsFCnt             *ttnpb.BoolValue
	Supports32BitFCnt      *ttnpb.BoolValue
	IsPending              bool
}

// DeviceRegistry is a registry, containing devices.
type DeviceRegistry interface {
	GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error)
	GetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error)
	RangeByUplinkMatches(ctx context.Context, up *ttnpb.UplinkMessage, f func(context.Context, *UplinkMatch) (bool, error)) error
	SetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
	Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error
	BatchGetByID(
		ctx context.Context, appID *ttnpb.ApplicationIdentifiers, deviceIDs []string, paths []string,
	) ([]*ttnpb.EndDevice, error)
	BatchDelete(
		ctx context.Context,
		appIDs *ttnpb.ApplicationIdentifiers,
		deviceIDs []string,
	) ([]*ttnpb.EndDeviceIdentifiers, error)
}

var errDeviceExists = errors.DefineAlreadyExists("device_exists", "device already exists")

// CreateDevice creates device dev in r.
func CreateDevice(ctx context.Context, r DeviceRegistry, dev *ttnpb.EndDevice, paths ...string) (*ttnpb.EndDevice, context.Context, error) {
	return r.SetByID(ctx, dev.Ids.ApplicationIds, dev.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel, func(_ context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if stored != nil {
			return nil, nil, errDeviceExists.New()
		}
		return dev, paths, nil
	})
}

// DeleteDevice deletes device identified by appID, devID from r.
func DeleteDevice(ctx context.Context, r DeviceRegistry, appID *ttnpb.ApplicationIdentifiers, devID string) error {
	_, _, err := r.SetByID(ctx, appID, devID, nil, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) { return nil, nil, nil })
	return err
}

func logRegistryRPCError(ctx context.Context, err error, msg string) {
	logger := log.FromContext(ctx).WithError(err)
	var printLog func(args ...any)
	switch {
	case errors.IsNotFound(err), errors.IsInvalidArgument(err), errors.IsCanceled(err):
		printLog = logger.Debug
	case errors.IsFailedPrecondition(err), errors.IsResourceExhausted(err):
		printLog = logger.Warn
	default:
		printLog = logger.Error
	}
	printLog(msg)
}

type replacedEndDeviceFieldRegistryWrapper struct {
	DeviceRegistry
	fields []registry.ReplacedEndDeviceField
}

func (w replacedEndDeviceFieldRegistryWrapper) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	dev, ctx, err := w.DeviceRegistry.GetByEUI(ctx, joinEUI, devEUI, paths)
	if err != nil || dev == nil {
		return dev, ctx, err
	}
	for _, d := range replaced {
		d.GetTransform(dev)
	}
	return dev, ctx, nil
}

func (w replacedEndDeviceFieldRegistryWrapper) GetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	dev, ctx, err := w.DeviceRegistry.GetByID(ctx, appID, devID, paths)
	if err != nil || dev == nil {
		return dev, ctx, err
	}
	for _, d := range replaced {
		d.GetTransform(dev)
	}
	return dev, ctx, nil
}

func (w replacedEndDeviceFieldRegistryWrapper) SetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	dev, ctx, err := w.DeviceRegistry.SetByID(ctx, appID, devID, paths, func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			for _, d := range replaced {
				d.GetTransform(dev)
			}
		}
		dev, paths, err := f(ctx, dev)
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
		return dev, ctx, err
	}
	for _, d := range replaced {
		d.GetTransform(dev)
	}
	return dev, ctx, nil
}

func (w replacedEndDeviceFieldRegistryWrapper) BatchDelete(
	ctx context.Context,
	appIDs *ttnpb.ApplicationIdentifiers,
	deviceIDs []string,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	return w.DeviceRegistry.BatchDelete(ctx, appIDs, deviceIDs)
}

func (w replacedEndDeviceFieldRegistryWrapper) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error {
	paths, replaced := registry.MatchReplacedEndDeviceFields(paths, w.fields)
	return w.DeviceRegistry.Range(ctx, paths, func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, dev *ttnpb.EndDevice) bool {
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
		DeviceRegistry: r,
		fields:         fields,
	}
}

var replacedEndDeviceFields = []registry.ReplacedEndDeviceField{
	{
		Old:          "mac_state.current_parameters.adr_ack_delay",
		New:          "mac_state.current_parameters.adr_ack_delay_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MacState == nil || dev.MacState.CurrentParameters == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MacState.CurrentParameters.AdrAckDelay = 0
			return nil
		},
	},
	{
		Old:          "mac_state.current_parameters.adr_ack_limit",
		New:          "mac_state.current_parameters.adr_ack_limit_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MacState == nil || dev.MacState.CurrentParameters == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MacState.CurrentParameters.AdrAckLimit = 0
			return nil
		},
	},
	{
		Old:          "mac_state.current_parameters.ping_slot_data_rate_index",
		New:          "mac_state.current_parameters.ping_slot_data_rate_index_value",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MacState == nil || dev.MacState.CurrentParameters == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MacState.CurrentParameters.PingSlotDataRateIndex = 0
			return nil
		},
	},
	{
		Old:          "mac_state.desired_parameters.adr_ack_delay",
		New:          "mac_state.desired_parameters.adr_ack_delay_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MacState == nil || dev.MacState.DesiredParameters == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MacState.DesiredParameters.AdrAckDelay = 0
			return nil
		},
	},
	{
		Old:          "mac_state.desired_parameters.adr_ack_limit",
		New:          "mac_state.desired_parameters.adr_ack_limit_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MacState == nil || dev.MacState.DesiredParameters == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MacState.DesiredParameters.AdrAckLimit = 0
			return nil
		},
	},
	{
		Old:          "mac_state.desired_parameters.ping_slot_data_rate_index",
		New:          "mac_state.desired_parameters.ping_slot_data_rate_index_value",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MacState == nil || dev.MacState.DesiredParameters == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MacState.DesiredParameters.PingSlotDataRateIndex = 0
			return nil
		},
	},
	{
		Old:          "queued_application_downlinks",
		New:          "session.queued_application_downlinks",
		GetTransform: func(dev *ttnpb.EndDevice) {},
		SetTransform: func(dev *ttnpb.EndDevice, useOld, useNew bool) error {
			switch {
			case useOld && useNew:
				oldValue := dev.QueuedApplicationDownlinks
				newValue := dev.GetSession().GetQueuedApplicationDownlinks()
				n := len(oldValue)
				if n != len(newValue) {
					return errInvalidFieldValue.WithAttributes("field", "queued_application_downlinks")
				}
				for i := 0; i < n; i++ {
					if !proto.Equal(oldValue[i], newValue[i]) {
						return errInvalidFieldValue.WithAttributes("field", "queued_application_downlinks")
					}
				}

			case useNew:
				dev.QueuedApplicationDownlinks = nil

			case dev.QueuedApplicationDownlinks == nil:
				if dev.Session != nil {
					dev.Session.QueuedApplicationDownlinks = nil
				}

			default:
				if dev.Session == nil {
					dev.Session = &ttnpb.Session{}
				}
				dev.Session.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks
			}
			dev.QueuedApplicationDownlinks = nil
			return nil
		},
	},
}

// ScheduledDownlinkMatcher matches scheduled downlinks with the TxAcknowledgement received by a gateway.
type ScheduledDownlinkMatcher interface {
	// Add stores metadata for a scheduled downlink message. Implementations may use the downlink
	// message correlation IDs to uniquely identify the scheduled downlink message.
	Add(ctx context.Context, down *ttnpb.DownlinkMessage) error
	// Match matches metadata of a scheduled downlink message from a TxAcknowledgement that was received by a gateway.
	// In case of a successful match, the scheduled downlink message is returned. If no downlink is matched, then an
	// error is returned instead. Implementations are free to return an error even when a match should have been
	// successful, for example if a long time has passed since the downlink was scheduled.
	Match(ctx context.Context, ack *ttnpb.TxAcknowledgment) (*ttnpb.DownlinkMessage, error)
}
