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

package joinserver_test

import (
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var registeredDevice = &ttnpb.EndDevice{
	EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-application",
		},
		DeviceID: "foo-device",
		JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		DevEUI:   eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
	},
	RootKeys: &ttnpb.RootKeys{
		RootKeyID: "test",
		NwkKey: &ttnpb.KeyEnvelope{
			KEKLabel: "test",
			Key:      []byte{0x1, 0x2},
		},
		AppKey: &ttnpb.KeyEnvelope{
			KEKLabel: "test",
			Key:      []byte{0x3, 0x4},
		},
	},
}

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestDeviceRegistryGet(t *testing.T) {
	a := assertions.New(t)
	ctx := test.ContextWithT(test.Context(), t)
	reg := &MockDeviceRegistry{
		GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
			if joinEUI == *registeredDevice.JoinEUI && devEUI == *registeredDevice.DevEUI {
				var res ttnpb.EndDevice
				if err := res.SetFields(registeredDevice, paths...); err != nil {
					return nil, err
				}
				return &res, nil
			}
			return nil, errors.DefineNotFound("not_found", "not found")
		},
	}
	srv := &JsDeviceServer{
		Registry: reg,
	}

	// Permission denied.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): nil,
			},
		})
		_, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	}

	// No EUIs.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
				),
			},
		})
		_, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
				DeviceID:               registeredDevice.DeviceID,
				JoinEUI:                registeredDevice.JoinEUI,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
		_, err = srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
				DeviceID:               registeredDevice.DeviceID,
				DevEUI:                 registeredDevice.DevEUI,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	}

	// Not found.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
				),
			},
		})
		_, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
				DeviceID:               "not-found",
				JoinEUI:                eui64Ptr(types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}),
				DevEUI:                 eui64Ptr(types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}),
			},
		})
		a.So(errors.IsNotFound(err), should.BeTrue)
	}

	// Get without keys.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
				),
			},
		})
		dev, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(err, should.BeNil)
		a.So(dev.EndDeviceIdentifiers, should.Resemble, registeredDevice.EndDeviceIdentifiers)
	}

	// Get keys; permission denied.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
				),
			},
		})
		_, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"root_keys"},
			},
		})
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	}

	// Get keys.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
				),
			},
		})
		dev, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"root_keys"},
			},
		})
		a.So(err, should.BeNil)
		a.So(dev.RootKeys, should.Resemble, registeredDevice.RootKeys)
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	a := assertions.New(t)
	ctx := test.ContextWithT(test.Context(), t)
	reg := &MockDeviceRegistry{
		SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
			var dev *ttnpb.EndDevice
			if joinEUI == *registeredDevice.JoinEUI && devEUI == *registeredDevice.DevEUI {
				dev = registeredDevice
			}
			var err error
			dev, _, err = cb(dev)
			a := assertions.New(test.MustTFromContext(ctx))
			a.So(err, should.BeNil)
			return dev, nil
		},
	}
	srv := &JsDeviceServer{
		Registry: reg,
	}

	// Permission denied.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): nil,
			},
		})
		_, err := srv.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: *registeredDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	}

	// No EUIs.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
				),
			},
		})
		_, err := srv.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
					DeviceID:               registeredDevice.DeviceID,
					JoinEUI:                registeredDevice.JoinEUI,
				},
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
		_, err = srv.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
					DeviceID:               registeredDevice.DeviceID,
					DevEUI:                 registeredDevice.DevEUI,
				},
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	}

	// Set without keys.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
				),
			},
		})
		dev, err := srv.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: *registeredDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"ids"},
			},
		})
		a.So(err, should.BeNil)
		a.So(dev.EndDeviceIdentifiers, should.Resemble, registeredDevice.EndDeviceIdentifiers)
	}

	// Set keys; permission denied.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
				),
			},
		})
		_, err := srv.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: *registeredDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"root_keys"},
			},
		})
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	}

	// Set keys.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				),
			},
		})
		dev, err := srv.Set(ctx, &ttnpb.SetEndDeviceRequest{
			Device: *registeredDevice,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"root_keys"},
			},
		})
		a.So(err, should.BeNil)
		a.So(dev.RootKeys, should.Resemble, registeredDevice.RootKeys)
	}
}

func TestDeviceRegistryDelete(t *testing.T) {
	a := assertions.New(t)
	ctx := test.ContextWithT(test.Context(), t)
	reg := &MockDeviceRegistry{
		SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
			var dev *ttnpb.EndDevice
			if joinEUI == *registeredDevice.JoinEUI && devEUI == *registeredDevice.DevEUI {
				dev = registeredDevice
			}
			var err error
			dev, _, err = cb(dev)
			a := assertions.New(test.MustTFromContext(ctx))
			a.So(err, should.BeNil)
			a.So(dev, should.BeNil)
			return nil, nil
		},
	}
	srv := &JsDeviceServer{
		Registry: reg,
	}

	// Permission denied.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): nil,
			},
		})
		_, err := srv.Delete(ctx, &registeredDevice.EndDeviceIdentifiers)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	}

	// No EUIs.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
				),
			},
		})
		_, err := srv.Delete(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
			DeviceID:               registeredDevice.DeviceID,
			JoinEUI:                registeredDevice.JoinEUI,
		})
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
		_, err = srv.Delete(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: registeredDevice.ApplicationIdentifiers,
			DeviceID:               registeredDevice.DeviceID,
			DevEUI:                 registeredDevice.DevEUI,
		})
		a.So(errors.IsInvalidArgument(err), should.BeTrue)
	}

	// Delete.
	{
		ctx := rights.NewContext(ctx, rights.Rights{
			ApplicationRights: map[string]*ttnpb.Rights{
				unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
					ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
				),
			},
		})
		_, err := srv.Delete(ctx, &registeredDevice.EndDeviceIdentifiers)
		a.So(err, should.BeNil)
	}
}

// Test the attack described on https://github.com/TheThingsIndustries/lorawan-stack/issues/1469.
func TestDeviceRegistryGetUnrightfulAccess(t *testing.T) {
	a := assertions.New(t)
	ctx := test.ContextWithT(test.Context(), t)
	reg := &MockDeviceRegistry{
		GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
			if joinEUI == *registeredDevice.JoinEUI && devEUI == *registeredDevice.DevEUI {
				var res ttnpb.EndDevice
				if err := res.SetFields(registeredDevice, paths...); err != nil {
					return nil, err
				}
				return &res, nil
			}
			return nil, errors.DefineNotFound("not_found", "not found")
		},
	}
	srv := &JsDeviceServer{
		Registry: reg,
	}

	other := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "foo-application-other",
			},
			DeviceID: "foo-device-other",
			JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			DevEUI:   eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		},
	}

	ctx = rights.NewContext(ctx, rights.Rights{
		ApplicationRights: map[string]*ttnpb.Rights{
			unique.ID(ctx, other.ApplicationIdentifiers): ttnpb.RightsFrom(
				ttnpb.RIGHT_APPLICATION_DEVICES_READ,
				ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
			),
		},
	})

	dev, err := srv.Get(ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIdentifiers: other.EndDeviceIdentifiers,
		FieldMask: pbtypes.FieldMask{
			Paths: []string{"root_keys"},
		},
	})
	a.So(err, should.NotBeNil)
	a.So(dev.RootKeys, should.BeNil)
}
