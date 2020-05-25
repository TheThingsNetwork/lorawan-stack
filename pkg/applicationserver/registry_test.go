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
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func DeleteDevice(ctx context.Context, r DeviceRegistry, ids ttnpb.EndDeviceIdentifiers) error {
	_, err := r.Set(ctx, ids, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) { return nil, nil, nil })
	return err
}

func handleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	a := assertions.New(t)

	ctx := test.Context()

	pb := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
			DeviceID:               "test-dev",
		},
		Session: &ttnpb.Session{
			DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
		},
		SkipPayloadCryptoOverride: &pbtypes.BoolValue{Value: true},
	}

	ret, err := reg.Get(ctx, pb.EndDeviceIdentifiers, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	start := time.Now()

	ret, err = reg.Set(ctx, pb.EndDeviceIdentifiers,
		[]string{
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
			"session",
			"skip_payload_crypto_override",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return pb, []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
				"pending_session",
				"session",
				"skip_payload_crypto_override",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to create device: %s", err)
	}
	a.So(ret.CreatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pb.CreatedAt = ret.CreatedAt
	pb.UpdatedAt = ret.UpdatedAt
	pb.SkipPayloadCrypto = true // Set because SkipPayloadCryptoOverride.GetValue() == true
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.Set(ctx, pb.EndDeviceIdentifiers,
		[]string{
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
			"session",
			"skip_payload_crypto",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			pb.SkipPayloadCrypto = false
			return pb, []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
				"pending_session",
				"session",
				"skip_payload_crypto",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to update device: %s", err)
	}
	a.So(ret.UpdatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.HappenAfter, ret.CreatedAt)
	if !a.So(ret.SkipPayloadCryptoOverride, should.NotBeNil) || !a.So(ret.SkipPayloadCryptoOverride.Value, should.BeFalse) {
		t.Fatalf("Setting deprecated field failed to update new field")
	}
	pb.UpdatedAt = ret.UpdatedAt
	pb.SkipPayloadCryptoOverride = ret.SkipPayloadCryptoOverride
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.Set(ctx, pb.EndDeviceIdentifiers,
		[]string{
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
			"pending_session",
			"session",
			"skip_payload_crypto_override",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			a.So(stored, should.HaveEmptyDiff, pb)
			return &ttnpb.EndDevice{}, nil, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to get device via Set: %s", err)
	}
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.Get(ctx, pb.EndDeviceIdentifiers, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	err = DeleteDevice(ctx, reg, pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.Get(ctx, pb.EndDeviceIdentifiers, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)
}

func TestDeviceRegistry(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"applicationserver_test",
		"devices",
	}
	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (reg DeviceRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(t testing.TB) (DeviceRegistry, func() error) {
				cl, flush := test.NewRedis(t, namespace[:]...)
				reg := &redis.DeviceRegistry{Redis: cl}
				return reg, func() error {
					flush()
					return cl.Close()
				}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				reg, closeFn := tc.New(t)
				reg = wrapEndDeviceRegistryWithReplacedFields(reg, replacedEndDeviceFields...)
				if closeFn != nil {
					defer func() {
						if err := closeFn(); err != nil {
							t.Errorf("Failed to close registry: %v", err)
						}
					}()
				}
				t.Run("1st run", func(t *testing.T) { handleDeviceRegistryTest(t, reg) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleDeviceRegistryTest(t, reg) })
			})
		}
	}
}

func handleLinkRegistryTest(t *testing.T, reg LinkRegistry) {
	a := assertions.New(t)
	ctx := test.Context()
	app1IDs := ttnpb.ApplicationIdentifiers{
		ApplicationID: "app-1",
	}
	app1 := &ttnpb.ApplicationLink{
		APIKey:               "secret1",
		NetworkServerAddress: "host1",
	}
	app2IDs := ttnpb.ApplicationIdentifiers{
		ApplicationID: "app-2",
	}
	app2 := &ttnpb.ApplicationLink{
		APIKey:               "secret2",
		NetworkServerAddress: "host2",
	}

	for ids, link := range map[ttnpb.ApplicationIdentifiers]*ttnpb.ApplicationLink{
		app1IDs: app1,
		app2IDs: app2,
	} {
		_, err := reg.Get(ctx, ids, ttnpb.ApplicationLinkFieldPathsTopLevel)
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.FailNow()
		}

		_, err = reg.Set(ctx, ids, nil, func(pb *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
			if pb != nil {
				t.Fatal("Link already exists")
			}
			return link, ttnpb.ApplicationLinkFieldPathsTopLevel, nil
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		pb, err := reg.Get(ctx, ids, ttnpb.ApplicationLinkFieldPathsTopLevel)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(pb, should.HaveEmptyDiff, link)
	}

	seen := make(map[string]*ttnpb.ApplicationLink)
	reg.Range(ctx, ttnpb.ApplicationLinkFieldPathsTopLevel, func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, pb *ttnpb.ApplicationLink) bool {
		uid := unique.ID(ctx, ids)
		seen[uid] = pb
		return true
	})
	ok := a.So(seen, should.HaveEmptyDiff, map[string]*ttnpb.ApplicationLink{
		unique.ID(ctx, app1IDs): app1,
		unique.ID(ctx, app2IDs): app2,
	})
	if !ok {
		t.FailNow()
	}

	for _, ids := range []ttnpb.ApplicationIdentifiers{app1IDs, app2IDs} {
		_, err := reg.Set(ctx, ids, nil, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
			return nil, nil, nil
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		_, err = reg.Get(ctx, ids, nil)
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.FailNow()
		}
	}
}

func TestLinkRegistry(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"applicationserver_test",
		"links",
	}
	for _, tc := range []struct {
		Name string
		New  func(t testing.TB) (reg LinkRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(t testing.TB) (LinkRegistry, func() error) {
				cl, flush := test.NewRedis(t, namespace[:]...)
				reg := &redis.LinkRegistry{Redis: cl}
				return reg, func() error {
					flush()
					return cl.Close()
				}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				reg, closeFn := tc.New(t)
				if closeFn != nil {
					defer func() {
						if err := closeFn(); err != nil {
							t.Errorf("Failed to close registry: %v", err)
						}
					}()
				}
				t.Run("1st run", func(t *testing.T) { handleLinkRegistryTest(t, reg) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleLinkRegistryTest(t, reg) })
			})
		}
	}
}
