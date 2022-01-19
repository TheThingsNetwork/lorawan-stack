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
		RootKeyId: DefaultRootKeyID,
	}

	baseSessionKeys = ttnpb.SessionKeys{}

	baseSession = ttnpb.Session{
		DevAddr: DefaultDevAddr,
		Keys:    &baseSessionKeys,
	}

	baseEndDeviceIdentifiers = ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &DefaultApplicationIdentifiers,
		DeviceId:       DefaultDeviceID,
	}

	baseMACState = ttnpb.MACState{
		LorawanVersion: DefaultMACVersion,
	}

	baseEndDevice = ttnpb.EndDevice{
		Ids: &baseEndDeviceIdentifiers,
	}
)

func (o SessionKeysOptionNamespace) WithNwkKeys(fNwkSIntKey, nwkSEncKey, sNwkSIntKey *ttnpb.KeyEnvelope) SessionKeysOption {
	return o.Compose(
		o.WithFNwkSIntKey(fNwkSIntKey),
		o.WithNwkSEncKey(nwkSEncKey),
		o.WithSNwkSIntKey(sNwkSIntKey),
	)
}

func (o SessionKeysOptionNamespace) WithDefaultNwkKeys(macVersion ttnpb.MACVersion) SessionKeysOption {
	nwkSEncKey := DefaultNwkSEncKeyEnvelope
	sNwkSIntKey := DefaultSNwkSIntKeyEnvelope
	if macVersion.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
		nwkSEncKey = DefaultFNwkSIntKeyEnvelope
		sNwkSIntKey = DefaultFNwkSIntKeyEnvelope
	}
	return o.WithNwkKeys(DefaultFNwkSIntKeyEnvelope, nwkSEncKey, sNwkSIntKey)
}

func (o SessionKeysOptionNamespace) WithDefaultNwkKeysWrapped(macVersion ttnpb.MACVersion) SessionKeysOption {
	nwkSEncKey := DefaultNwkSEncKeyEnvelopeWrapped
	sNwkSIntKey := DefaultSNwkSIntKeyEnvelopeWrapped
	if macVersion.Compare(ttnpb.MACVersion_MAC_V1_1) < 0 {
		nwkSEncKey = DefaultFNwkSIntKeyEnvelopeWrapped
		sNwkSIntKey = DefaultFNwkSIntKeyEnvelopeWrapped
	}
	return o.WithNwkKeys(DefaultFNwkSIntKeyEnvelopeWrapped, nwkSEncKey, sNwkSIntKey)
}

func (o SessionKeysOptionNamespace) WithDefaultAppSKey() SessionKeysOption {
	return o.WithAppSKey(&ttnpb.KeyEnvelope{
		Key: &DefaultAppSKey,
	})
}

func (o SessionKeysOptionNamespace) WithDefaultSessionKeyID() SessionKeysOption {
	return o.WithSessionKeyId(DefaultSessionKeyID)
}

func (o SessionOptionNamespace) WithSessionKeysOptions(opts ...SessionKeysOption) SessionOption {
	return func(x ttnpb.Session) ttnpb.Session {
		keys := SessionKeysOptions.Compose(opts...)(*x.Keys)
		x.Keys = &keys
		return x
	}
}

func (o SessionOptionNamespace) WithDefaultNwkKeys(macVersion ttnpb.MACVersion) SessionOption {
	return o.WithSessionKeysOptions(SessionKeysOptions.WithDefaultNwkKeys(macVersion))
}

func (o SessionOptionNamespace) WithDefaultAppSKey() SessionOption {
	return o.WithSessionKeysOptions(SessionKeysOptions.WithDefaultAppSKey())
}

func (o MACStateOptionNamespace) AppendRecentUplinks(ups ...*ttnpb.UplinkMessage) MACStateOption {
	return func(x ttnpb.MACState) ttnpb.MACState {
		x.RecentUplinks = append(x.RecentUplinks, ups...)
		return x
	}
}

func (o MACStateOptionNamespace) AppendRecentDownlinks(downs ...*ttnpb.DownlinkMessage) MACStateOption {
	return func(x ttnpb.MACState) ttnpb.MACState {
		x.RecentDownlinks = append(x.RecentDownlinks, downs...)
		return x
	}
}

func (o EndDeviceIdentifiersOptionNamespace) WithDefaultJoinEUI() EndDeviceIdentifiersOption {
	return o.WithJoinEui(&DefaultJoinEUI)
}

func (o EndDeviceIdentifiersOptionNamespace) WithDefaultDevEUI() EndDeviceIdentifiersOption {
	return o.WithDevEui(&DefaultDevEUI)
}

func (o EndDeviceOptionNamespace) WithEndDeviceIdentifiersOptions(opts ...EndDeviceIdentifiersOption) EndDeviceOption {
	return func(x ttnpb.EndDevice) ttnpb.EndDevice {
		ids := EndDeviceIdentifiersOptions.Compose(opts...)(*x.Ids)
		x.Ids = &ids
		return x
	}
}

func (o EndDeviceOptionNamespace) WithJoinEUI(v *types.EUI64) EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithJoinEui(v))
}

func (o EndDeviceOptionNamespace) WithDefaultJoinEUI() EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithDefaultJoinEUI())
}

func (o EndDeviceOptionNamespace) WithDevEUI(v *types.EUI64) EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithDevEui(v))
}

func (o EndDeviceOptionNamespace) WithDefaultDevEUI() EndDeviceOption {
	return o.WithEndDeviceIdentifiersOptions(EndDeviceIdentifiersOptions.WithDefaultDevEUI())
}

func (o EndDeviceOptionNamespace) WithDefaultFrequencyPlanID() EndDeviceOption {
	return o.WithFrequencyPlanId(DefaultFrequencyPlanID)
}

func (o EndDeviceOptionNamespace) WithDefaultLoRaWANVersion() EndDeviceOption {
	return o.WithLorawanVersion(DefaultMACVersion)
}

func (o EndDeviceOptionNamespace) WithDefaultLoRaWANPHYVersion() EndDeviceOption {
	return o.WithLorawanPhyVersion(DefaultPHYVersion)
}

func (o EndDeviceOptionNamespace) WithMACStateOptions(opts ...MACStateOption) EndDeviceOption {
	return func(x ttnpb.EndDevice) ttnpb.EndDevice {
		if x.MacState == nil {
			panic("MACState is nil")
		}
		v := MACStateOptions.Compose(opts...)(*x.MacState)
		x.MacState = &v
		return x
	}
}

func (o EndDeviceOptionNamespace) WithPendingMACStateOptions(opts ...MACStateOption) EndDeviceOption {
	return func(x ttnpb.EndDevice) ttnpb.EndDevice {
		if x.PendingMacState == nil {
			panic("PendingMACState is nil")
		}
		v := MACStateOptions.Compose(opts...)(*x.PendingMacState)
		x.PendingMacState = &v
		return x
	}
}
