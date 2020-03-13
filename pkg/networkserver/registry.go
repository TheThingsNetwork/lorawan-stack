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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// DeviceRegistry is a registry, containing devices.
type DeviceRegistry interface {
	GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error)
	GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error)
	RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error
	SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
}

func logRegistryRPCError(ctx context.Context, err error, msg string) {
	logger := log.FromContext(ctx).WithError(err)
	var printLog func(string)
	if errors.IsNotFound(err) || errors.IsInvalidArgument(err) {
		printLog = logger.Debug
	} else {
		printLog = logger.Error
	}
	printLog(msg)
}

type deprecatedDeviceField struct {
	Old          string
	New          string
	GetTransform func(dev *ttnpb.EndDevice)
	SetTransform func(dev *ttnpb.EndDevice, useOld, useNew bool) error
}

type deprecatedDeviceFieldMatch struct {
	deprecatedDeviceField
	MatchedOld bool
	MatchedNew bool
}

type deprecatedDeviceFieldRegistryWrapper struct {
	fields   []deprecatedDeviceField
	registry DeviceRegistry
}

func matchDeprecatedDeviceFields(paths []string, deprecated []deprecatedDeviceField) ([]string, []deprecatedDeviceFieldMatch) {
	var matched []deprecatedDeviceFieldMatch
	for _, f := range deprecated {
		hasOld, hasNew := ttnpb.HasAnyField(paths, f.Old), ttnpb.HasAnyField(paths, f.New)
		switch {
		case !hasOld && !hasNew:
			continue
		case hasOld && hasNew:
		case hasOld:
			paths = ttnpb.AddFields(paths, f.New)
		case hasNew:
			paths = ttnpb.AddFields(paths, f.Old)
		}
		matched = append(matched, deprecatedDeviceFieldMatch{
			deprecatedDeviceField: f,
			MatchedOld:            hasOld,
			MatchedNew:            hasNew,
		})
	}
	return paths, matched
}

func (w deprecatedDeviceFieldRegistryWrapper) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	paths, deprecated := matchDeprecatedDeviceFields(paths, w.fields)
	dev, ctx, err := w.registry.GetByEUI(ctx, joinEUI, devEUI, paths)
	if err != nil || dev == nil {
		return dev, ctx, err
	}
	for _, d := range deprecated {
		d.GetTransform(dev)
	}
	return dev, ctx, nil
}

func (w deprecatedDeviceFieldRegistryWrapper) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	paths, deprecated := matchDeprecatedDeviceFields(paths, w.fields)
	dev, ctx, err := w.registry.GetByID(ctx, appID, devID, paths)
	if err != nil || dev == nil {
		return dev, ctx, err
	}
	for _, d := range deprecated {
		d.GetTransform(dev)
	}
	return dev, ctx, nil
}

func (w deprecatedDeviceFieldRegistryWrapper) RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error {
	paths, deprecated := matchDeprecatedDeviceFields(paths, w.fields)
	return w.registry.RangeByAddr(ctx, devAddr, paths, func(ctx context.Context, dev *ttnpb.EndDevice) bool {
		if dev != nil {
			for _, d := range deprecated {
				d.GetTransform(dev)
			}
		}
		return f(ctx, dev)
	})
}

func (w deprecatedDeviceFieldRegistryWrapper) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	paths, deprecated := matchDeprecatedDeviceFields(paths, w.fields)
	dev, ctx, err := w.registry.SetByID(ctx, appID, devID, paths, func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			for _, d := range deprecated {
				d.GetTransform(dev)
			}
		}
		dev, paths, err := f(ctx, dev)
		if err != nil || dev == nil {
			return dev, paths, err
		}
		for _, d := range deprecated {
			d.SetTransform(dev, d.MatchedOld, d.MatchedNew)
		}
		return dev, paths, nil
	})
	if err != nil || dev == nil {
		return dev, ctx, err
	}
	for _, d := range deprecated {
		d.GetTransform(dev)
	}
	return dev, ctx, nil
}

func wrapDeviceRegistryWithDeprecatedFields(r DeviceRegistry, fields ...deprecatedDeviceField) DeviceRegistry {
	return deprecatedDeviceFieldRegistryWrapper{
		fields:   fields,
		registry: r,
	}
}

var deprecatedDeviceFields = []deprecatedDeviceField{
	{
		Old: "mac_state.current_parameters.adr_ack_delay",
		New: "mac_state.current_parameters.adr_ack_delay_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.MACState == nil {
				return
			}
			dev.MACState.CurrentParameters.ADRAckDelay = uint32(dev.MACState.CurrentParameters.ADRAckDelayExponent.GetValue())
		},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MACState == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MACState.CurrentParameters.ADRAckDelay = 0
			return nil
		},
	},
	{
		Old: "mac_state.current_parameters.adr_ack_limit",
		New: "mac_state.current_parameters.adr_ack_limit_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.MACState == nil {
				return
			}
			dev.MACState.CurrentParameters.ADRAckLimit = uint32(dev.MACState.CurrentParameters.ADRAckLimitExponent.GetValue())
		},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MACState == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MACState.CurrentParameters.ADRAckLimit = 0
			return nil
		},
	},
	{
		Old: "mac_state.current_parameters.ping_slot_data_rate_index",
		New: "mac_state.current_parameters.ping_slot_data_rate_index_value",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.MACState == nil {
				return
			}
			dev.MACState.CurrentParameters.PingSlotDataRateIndex = dev.MACState.CurrentParameters.PingSlotDataRateIndexValue.GetValue()
		},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MACState == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MACState.CurrentParameters.PingSlotDataRateIndex = 0
			return nil
		},
	},
	{
		Old: "mac_state.desired_parameters.adr_ack_delay",
		New: "mac_state.desired_parameters.adr_ack_delay_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.MACState == nil {
				return
			}
			dev.MACState.DesiredParameters.ADRAckDelay = uint32(dev.MACState.DesiredParameters.ADRAckDelayExponent.GetValue())
		},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MACState == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MACState.DesiredParameters.ADRAckDelay = 0
			return nil
		},
	},
	{
		Old: "mac_state.desired_parameters.adr_ack_limit",
		New: "mac_state.desired_parameters.adr_ack_limit_exponent",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.MACState == nil {
				return
			}
			dev.MACState.DesiredParameters.ADRAckLimit = uint32(dev.MACState.DesiredParameters.ADRAckLimitExponent.GetValue())
		},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MACState == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MACState.DesiredParameters.ADRAckLimit = 0
			return nil
		},
	},
	{
		Old: "mac_state.desired_parameters.ping_slot_data_rate_index",
		New: "mac_state.desired_parameters.ping_slot_data_rate_index_value",
		GetTransform: func(dev *ttnpb.EndDevice) {
			if dev.MACState == nil {
				return
			}
			dev.MACState.DesiredParameters.PingSlotDataRateIndex = dev.MACState.DesiredParameters.PingSlotDataRateIndexValue.GetValue()
		},
		SetTransform: func(dev *ttnpb.EndDevice, _, _ bool) error {
			if dev.MACState == nil {
				return nil
			}
			// Replicate old behavior for backwards-compatibility.
			dev.MACState.DesiredParameters.PingSlotDataRateIndex = 0
			return nil
		},
	},
	{
		Old: "queued_application_downlinks",
		New: "session.queued_application_downlinks",
		GetTransform: func(dev *ttnpb.EndDevice) {
			switch {
			case dev.QueuedApplicationDownlinks == nil && dev.GetSession().GetQueuedApplicationDownlinks() == nil:
				return

			case dev.QueuedApplicationDownlinks != nil:
				if dev.Session == nil {
					dev.Session = &ttnpb.Session{}
				}
				dev.Session.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks

			default:
				dev.QueuedApplicationDownlinks = dev.Session.QueuedApplicationDownlinks
			}
		},
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
					if !oldValue[i].Equal(newValue[i]) {
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
