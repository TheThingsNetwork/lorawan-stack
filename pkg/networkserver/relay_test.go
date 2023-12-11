// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package networkserver_test

import (
	"context"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestBatchDeriveRootWorSKey(t *testing.T) {
	t.Parallel()
	keyVault := cryptoutil.NewMemKeyVault(nil)
	keyService := crypto.NewKeyService(keyVault)
	filterFields := func(
		t *testing.T, a *assertions.Assertion, devices []*ttnpb.EndDevice, paths []string,
	) []*ttnpb.EndDevice {
		t.Helper()
		devices = append(devices[:0:0], devices...)
		for i, dev := range devices {
			var err error
			devices[i], err = ttnpb.FilterGetEndDevice(dev, paths...)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
		}
		return devices
	}
	for _, tc := range []struct {
		Name string

		ApplicationIDs *ttnpb.ApplicationIdentifiers
		DeviceIDs      []string
		SessionKeyIDs  [][]byte

		DevAddrs []*types.DevAddr
		Keys     []*types.AES128Key

		BatchGetByIDFunc func(
			ctx context.Context,
			t *testing.T,
			a *assertions.Assertion,
			appIDs *ttnpb.ApplicationIdentifiers,
			devIDs []string,
			paths []string,
		) ([]*ttnpb.EndDevice, error)
	}{
		{
			Name: "no devices",

			ApplicationIDs: test.DefaultApplicationIdentifiers,

			BatchGetByIDFunc: func(
				ctx context.Context,
				t *testing.T,
				a *assertions.Assertion,
				appIDs *ttnpb.ApplicationIdentifiers,
				devIDs []string,
				paths []string,
			) ([]*ttnpb.EndDevice, error) {
				t.Helper()
				t.Fatal("BatchGetByID must not be called")
				return nil, nil
			},
		},
		{
			Name: "no sessions",

			ApplicationIDs: test.DefaultApplicationIdentifiers,
			DeviceIDs:      []string{"device1", "device2"},
			SessionKeyIDs:  [][]byte{{0x01}, {0x02}},

			DevAddrs: []*types.DevAddr{nil, nil},
			Keys:     []*types.AES128Key{nil, nil},

			BatchGetByIDFunc: func(
				ctx context.Context,
				t *testing.T,
				a *assertions.Assertion,
				appIDs *ttnpb.ApplicationIdentifiers,
				devIDs []string,
				paths []string,
			) ([]*ttnpb.EndDevice, error) {
				t.Helper()
				a.So(appIDs, should.Resemble, test.DefaultApplicationIdentifiers)
				a.So(devIDs, should.Resemble, []string{"device1", "device2"})
				return filterFields(t, a, []*ttnpb.EndDevice{
					{
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: test.DefaultApplicationIdentifiers,
							DeviceId:       "device1",
						},
					},
					{
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: test.DefaultApplicationIdentifiers,
							DeviceId:       "device2",
						},
					},
				}, paths), nil
			},
		},
		{
			Name: "only sessions",

			ApplicationIDs: test.DefaultApplicationIdentifiers,
			DeviceIDs:      []string{"device1", "device2"},
			SessionKeyIDs:  [][]byte{{0x01}, {0x02}},

			DevAddrs: []*types.DevAddr{{0x42, 0x42, 0x42, 0x42}, {0x43, 0x43, 0x43, 0x43}},
			Keys: []*types.AES128Key{
				{0xEE, 0x91, 0xDC, 0x1A, 0x66, 0x66, 0xC0, 0x6E, 0x82, 0x77, 0xDE, 0x6D, 0xB4, 0xDB, 0x94, 0x5F},
				{0xC2, 0x7E, 0x77, 0x4E, 0x20, 0x73, 0x18, 0x96, 0xFE, 0x20, 0x5D, 0x77, 0x1D, 0x7B, 0xC1, 0xF1},
			},

			BatchGetByIDFunc: func(
				ctx context.Context,
				t *testing.T,
				a *assertions.Assertion,
				appIDs *ttnpb.ApplicationIdentifiers,
				devIDs []string,
				paths []string,
			) ([]*ttnpb.EndDevice, error) {
				t.Helper()
				a.So(appIDs, should.Resemble, test.DefaultApplicationIdentifiers)
				a.So(devIDs, should.Resemble, []string{"device1", "device2"})
				return filterFields(t, a, []*ttnpb.EndDevice{
					{
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: test.DefaultApplicationIdentifiers,
							DeviceId:       "device1",
						},
						Session: &ttnpb.Session{
							DevAddr: types.DevAddr{0x42, 0x42, 0x42, 0x42}.Bytes(),
							Keys: &ttnpb.SessionKeys{
								SessionKeyId: []byte{0x01},
								NwkSEncKey: &ttnpb.KeyEnvelope{
									Key: []byte{0xCE, 0x07, 0xA0, 0x09, 0xA3, 0x97, 0x0A, 0xC0, 0x51, 0x9A, 0x09, 0x9E, 0xD5, 0x3E, 0x55, 0x0B}, // nolint:lll
								},
							},
						},
					},
					{
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: test.DefaultApplicationIdentifiers,
							DeviceId:       "device2",
						},
						Session: &ttnpb.Session{
							DevAddr: types.DevAddr{0x43, 0x43, 0x43, 0x43}.Bytes(),
							Keys: &ttnpb.SessionKeys{
								SessionKeyId: []byte{0x02},
								NwkSEncKey: &ttnpb.KeyEnvelope{
									Key: []byte{0xCE, 0x07, 0xA0, 0x09, 0xA3, 0x97, 0x0A, 0xC0, 0x51, 0x9A, 0x09, 0x9E, 0xD5, 0x3E, 0x55, 0x0C}, // nolint:lll
								},
							},
						},
					},
				}, paths), nil
			},
		},
		{
			Name: "with pending sessions",

			ApplicationIDs: test.DefaultApplicationIdentifiers,
			DeviceIDs:      []string{"device1", "device2"},
			SessionKeyIDs:  [][]byte{{0x03}, {0x02}},

			DevAddrs: []*types.DevAddr{{0x44, 0x44, 0x44, 0x44}, {0x43, 0x43, 0x43, 0x43}},
			Keys: []*types.AES128Key{
				{0x67, 0x11, 0x92, 0xD5, 0x9C, 0x0D, 0x35, 0x7D, 0xEF, 0xE2, 0xA9, 0x45, 0x21, 0xC4, 0x22, 0x7C},
				{0xC2, 0x7E, 0x77, 0x4E, 0x20, 0x73, 0x18, 0x96, 0xFE, 0x20, 0x5D, 0x77, 0x1D, 0x7B, 0xC1, 0xF1},
			},

			BatchGetByIDFunc: func(
				ctx context.Context,
				t *testing.T,
				a *assertions.Assertion,
				appIDs *ttnpb.ApplicationIdentifiers,
				devIDs []string,
				paths []string,
			) ([]*ttnpb.EndDevice, error) {
				t.Helper()
				a.So(appIDs, should.Resemble, test.DefaultApplicationIdentifiers)
				a.So(devIDs, should.Resemble, []string{"device1", "device2"})
				return filterFields(t, a, []*ttnpb.EndDevice{
					{
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: test.DefaultApplicationIdentifiers,
							DeviceId:       "device1",
						},
						Session: &ttnpb.Session{
							DevAddr: types.DevAddr{0x42, 0x42, 0x42, 0x42}.Bytes(),
							Keys: &ttnpb.SessionKeys{
								SessionKeyId: []byte{0x01},
								NwkSEncKey: &ttnpb.KeyEnvelope{
									Key: []byte{0xCE, 0x07, 0xA0, 0x09, 0xA3, 0x97, 0x0A, 0xC0, 0x51, 0x9A, 0x09, 0x9E, 0xD5, 0x3E, 0x55, 0x0B}, // nolint:lll
								},
							},
						},
						PendingSession: &ttnpb.Session{
							DevAddr: types.DevAddr{0x44, 0x44, 0x44, 0x44}.Bytes(),
							Keys: &ttnpb.SessionKeys{
								SessionKeyId: []byte{0x03},
								NwkSEncKey: &ttnpb.KeyEnvelope{
									Key: []byte{0xCE, 0x07, 0xA0, 0x09, 0xA3, 0x97, 0x0A, 0xC0, 0x51, 0x9A, 0x09, 0x9E, 0xD5, 0x3E, 0x55, 0x0D}, // nolint:lll
								},
							},
						},
					},
					{
						Ids: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: test.DefaultApplicationIdentifiers,
							DeviceId:       "device2",
						},
						Session: &ttnpb.Session{
							DevAddr: types.DevAddr{0x43, 0x43, 0x43, 0x43}.Bytes(),
							Keys: &ttnpb.SessionKeys{
								SessionKeyId: []byte{0x02},
								NwkSEncKey: &ttnpb.KeyEnvelope{
									Key: []byte{0xCE, 0x07, 0xA0, 0x09, 0xA3, 0x97, 0x0A, 0xC0, 0x51, 0x9A, 0x09, 0x9E, 0xD5, 0x3E, 0x55, 0x0C}, // nolint:lll
								},
							},
						},
					},
				}, paths), nil
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				relayKeyService := networkserver.NewRelayKeyService(
					&networkserver.MockDeviceRegistry{
						BatchGetByIDFunc: func(
							ctx context.Context,
							appIDs *ttnpb.ApplicationIdentifiers,
							devIDs []string,
							paths []string,
						) ([]*ttnpb.EndDevice, error) {
							t.Helper()
							return tc.BatchGetByIDFunc(ctx, t, a, appIDs, devIDs, paths)
						},
					},
					keyService,
				)
				devAddrs, keys, err := relayKeyService.BatchDeriveRootWorSKey(
					ctx, tc.ApplicationIDs, tc.DeviceIDs, tc.SessionKeyIDs,
				)
				if a.So(err, should.BeNil) {
					a.So(devAddrs, should.Resemble, tc.DevAddrs)
					a.So(keys, should.Resemble, tc.Keys)
				}
			},
		})
	}
}
