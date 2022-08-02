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

func DeleteDevice(ctx context.Context, r DeviceRegistry, ids *ttnpb.EndDeviceIdentifiers) error {
	_, err := r.Set(ctx, ids, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) { return nil, nil, nil })
	return err
}

func handleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	a := assertions.New(t)

	ctx := test.Context()

	pb := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			JoinEui:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			DevEui:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
			DeviceId:       "test-dev",
		},
		Session: &ttnpb.Session{
			DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff}.Bytes(),
			Keys: &ttnpb.SessionKeys{
				SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
			},
		},
		SkipPayloadCryptoOverride: &pbtypes.BoolValue{Value: true},
	}

	ret, err := reg.Get(ctx, pb.Ids, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	start := time.Now()

	ret, err = reg.Set(ctx, pb.Ids,
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
	a.So(*ttnpb.StdTime(ret.CreatedAt), should.HappenAfter, start)
	a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pb.CreatedAt = ret.CreatedAt
	pb.UpdatedAt = ret.UpdatedAt
	pb.SkipPayloadCrypto = true // Set because SkipPayloadCryptoOverride.GetValue() == true
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.Set(ctx, pb.Ids,
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
	a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, start)
	a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, *ttnpb.StdTime(ret.CreatedAt))
	if !a.So(ret.SkipPayloadCryptoOverride, should.NotBeNil) || !a.So(ret.SkipPayloadCryptoOverride.Value, should.BeFalse) {
		t.Fatalf("Setting deprecated field failed to update new field")
	}
	pb.UpdatedAt = ret.UpdatedAt
	pb.SkipPayloadCryptoOverride = ret.SkipPayloadCryptoOverride
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.Set(ctx, pb.Ids,
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

	ret, err = reg.Get(ctx, pb.Ids, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	err = DeleteDevice(ctx, reg, pb.Ids)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.Get(ctx, pb.Ids, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)
}

func TestDeviceRegistry(t *testing.T) {
	namespace := [...]string{
		"applicationserver_test",
		"devices",
	}
	for _, tc := range []struct {
		Name string
		New  func(ctx context.Context) (reg DeviceRegistry, closeFn func() error, err error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(ctx context.Context) (DeviceRegistry, func() error, error) {
				cl, flush := test.NewRedis(ctx, namespace[:]...)
				registry := &redis.DeviceRegistry{
					Redis:   cl,
					LockTTL: test.Delay << 10,
				}
				if err := registry.Init(ctx); err != nil {
					return nil, nil, err
				}
				return registry, func() error {
					flush()
					return cl.Close()
				}, nil
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%s/%d", tc.Name, i),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					reg, closeFn, err := tc.New(ctx)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
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
				},
			})
		}
	}
}

func handleLinkRegistryTest(t *testing.T, reg LinkRegistry) {
	a := assertions.New(t)
	ctx := test.Context()
	app1IDs := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "app-1",
	}
	app1 := &ttnpb.ApplicationLink{
		SkipPayloadCrypto: &pbtypes.BoolValue{
			Value: true,
		},
	}
	app2IDs := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "app-2",
	}
	app2 := &ttnpb.ApplicationLink{
		SkipPayloadCrypto: &pbtypes.BoolValue{
			Value: false,
		},
	}

	for ids, link := range map[*ttnpb.ApplicationIdentifiers]*ttnpb.ApplicationLink{
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
	err := reg.Range(ctx, ttnpb.ApplicationLinkFieldPathsTopLevel, func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, pb *ttnpb.ApplicationLink) bool {
		uid := unique.ID(ctx, ids)
		seen[uid] = pb
		return true
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(len(seen), should.Equal, 2) ||
		!a.So(seen[unique.ID(ctx, app1IDs)], should.Resemble, app1) ||
		!a.So(seen[unique.ID(ctx, app2IDs)], should.Resemble, app2) {
		t.FailNow()
	}

	for _, ids := range []*ttnpb.ApplicationIdentifiers{app1IDs, app2IDs} {
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
	namespace := [...]string{
		"applicationserver_test",
		"links",
	}
	for _, tc := range []struct {
		Name string
		New  func(ctx context.Context) (reg LinkRegistry, closeFn func() error, err error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(ctx context.Context) (LinkRegistry, func() error, error) {
				cl, flush := test.NewRedis(ctx, namespace[:]...)
				registry := &redis.LinkRegistry{
					Redis:   cl,
					LockTTL: test.Delay << 10,
				}
				if err := registry.Init(ctx); err != nil {
					return nil, nil, err
				}
				return registry, func() error {
					flush()
					return cl.Close()
				}, nil
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%s/%d", tc.Name, i),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					reg, closeFn, err := tc.New(ctx)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
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
				},
			})
		}
	}
}

func TestApplicationUplinksRegistry(t *testing.T) {
	namespace := [...]string{
		"applicationserver_test",
		"application_ups",
	}
	test.RunTest(t, test.TestConfig{
		Func: func(ctx context.Context, a *assertions.Assertion) {
			cl, flush := test.NewRedis(ctx, namespace[:]...)
			defer flush()
			defer cl.Close()
			registry := &redis.ApplicationUplinkRegistry{
				Redis: cl,
				Limit: 16,
			}

			ids := &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
				DeviceId:       "test-dev",
			}

			assertEmpty := func() {
				err := registry.Range(ctx, ids, []string{}, func(ctx context.Context, up *ttnpb.ApplicationUplink) bool {
					t.Fatal("list should be empty")
					return false
				})
				a.So(err, should.BeNil)
			}
			assertEmpty()

			push := func(start uint32, end uint32) {
				for i := start; i < end; i++ {
					err := registry.Push(ctx, ids, &ttnpb.ApplicationUplink{FCnt: i})
					a.So(err, should.BeNil)
				}
			}
			push(0, 16)

			assertList := func(start uint32, end uint32) {
				found := make(map[uint32]struct{})
				err := registry.Range(ctx, ids, []string{
					"f_cnt",
				}, func(ctx context.Context, up *ttnpb.ApplicationUplink) bool {
					a.So(up.FCnt, should.BeBetweenOrEqual, start, end-1)
					if _, f := found[up.FCnt]; f {
						t.Fatalf("Uplink already seen: %d", up.FCnt)
					}
					found[up.FCnt] = struct{}{}
					return true
				})
				a.So(err, should.BeNil)
				a.So(len(found), should.Equal, end-start)
			}
			assertList(0, 16)

			push(32, 64)
			// Since we allow at most 16 uplinks, only the last 16 are stored.
			assertList(48, 64)

			err := registry.Clear(ctx, ids)
			a.So(err, should.BeNil)
			assertEmpty()
		},
	})
}
