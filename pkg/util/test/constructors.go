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

package test

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

//go:generate go run ./generate_constructors.go

var (
	baseRootKeys = ttnpb.RootKeys{
		RootKeyID: DefaultRootKeyID,
	}

	baseSessionKeys = ttnpb.SessionKeys{
		SessionKeyID: DefaultSessionKeyID,
	}

	baseSession = ttnpb.Session{
		DevAddr:     DefaultDevAddr,
		SessionKeys: baseSessionKeys,
	}

	baseEndDeviceIdentifiers = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: DefaultApplicationIdentifiers,
		DeviceID:               DefaultDeviceID,
	}

	baseMACState = ttnpb.MACState{
		LoRaWANVersion: DefaultMACVersion,
	}

	baseEndDevice = ttnpb.EndDevice{
		EndDeviceIdentifiers: baseEndDeviceIdentifiers,
	}
)

func (o sessionKeysOptions) WithDefaultNwkKeys(macVersion ttnpb.MACVersion) SessionKeysOption {
	nwkSEncKey := DefaultNwkSEncKey
	sNwkSIntKey := DefaultSNwkSIntKey
	if macVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		nwkSEncKey = DefaultFNwkSIntKey
		sNwkSIntKey = DefaultFNwkSIntKey
	}
	return o.Compose(
		o.WithFNwkSIntKey(&ttnpb.KeyEnvelope{
			Key: &DefaultFNwkSIntKey,
		}),
		o.WithNwkSEncKey(&ttnpb.KeyEnvelope{
			Key: &nwkSEncKey,
		}),
		o.WithSNwkSIntKey(&ttnpb.KeyEnvelope{
			Key: &sNwkSIntKey,
		}),
	)
}

func (o sessionKeysOptions) WithDefaultAppSKey() SessionKeysOption {
	return o.WithAppSKey(&ttnpb.KeyEnvelope{
		Key: &DefaultAppSKey,
	})
}

func (o sessionOptions) WithSessionKeysOptions(opts ...SessionKeysOption) SessionOption {
	return func(x ttnpb.Session) ttnpb.Session {
		x.SessionKeys = SessionKeysOptions.Compose(opts...)(x.SessionKeys)
		return x
	}
}

func (o endDeviceIdentifiersOptions) WithDefaultJoinEUI() EndDeviceIdentifiersOption {
	return o.WithJoinEUI(&DefaultJoinEUI)
}

func (o endDeviceIdentifiersOptions) WithDefaultDevEUI() EndDeviceIdentifiersOption {
	return o.WithDevEUI(&DefaultDevEUI)
}

func (o endDeviceOptions) WithEndDeviceIdentifiersOptions(opts ...EndDeviceIdentifiersOption) EndDeviceOption {
	return func(x ttnpb.EndDevice) ttnpb.EndDevice {
		x.EndDeviceIdentifiers = EndDeviceIdentifiersOptions.Compose(opts...)(x.EndDeviceIdentifiers)
		return x
	}
}

func (o endDeviceOptions) WithJoinEUI(v *types.EUI64) EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithJoinEUI(v))
}

func (o endDeviceOptions) WithDefaultJoinEUI() EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithDefaultJoinEUI())
}

func (o endDeviceOptions) WithDevEUI(v *types.EUI64) EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithDevEUI(v))
}

func (o endDeviceOptions) WithDefaultDevEUI() EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithDefaultDevEUI())
}
