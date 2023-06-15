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

package redis

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

var appIDs = &ttnpb.ApplicationIdentifiers{
	ApplicationId: "test-app-id",
}

var devIDs = &ttnpb.EndDeviceIdentifiers{
	ApplicationIds: appIDs,
	DeviceId:       "test-dev-id",
}

func TestPkgRegistryClearDefaultAssociations(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)
	redisCl, cleanup := test.NewRedis(ctx, "assoc_test")
	t.Cleanup(func() {
		cleanup()
		if err := redisCl.Close(); err != nil {
			t.FailNow()
		}
	})

	registry, err := NewApplicationPackagesRegistry(ctx, redisCl, 10*time.Second)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	appPkgIds := &ttnpb.ApplicationPackageDefaultAssociationIdentifiers{
		ApplicationIds: appIDs,
		FPort:          201,
	}
	_, err = registry.SetDefaultAssociation(
		ctx, appPkgIds, nil, func(apa *ttnpb.ApplicationPackageDefaultAssociation) (
			*ttnpb.ApplicationPackageDefaultAssociation, []string, error,
		) {
			return &ttnpb.ApplicationPackageDefaultAssociation{
					Ids:         appPkgIds,
					PackageName: "lora-cloud-geolocation-v3",
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"key": {
								Kind: &structpb.Value_StringValue{
									StringValue: "value",
								},
							},
						},
					},
				}, []string{
					"ids",
					"package_name",
					"data",
				}, nil
		},
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	actual, err := registry.ListDefaultAssociations(ctx, appIDs, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(len(actual), should.Equal, 1)

	err = registry.ClearDefaultAssociations(ctx, appIDs)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	actual, err = registry.ListDefaultAssociations(ctx, appIDs, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(len(actual), should.Equal, 0)
}

func TestPackageClearAssociations(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)
	redisCl, cleanup := test.NewRedis(ctx, "assoc_test")
	t.Cleanup(func() {
		cleanup()
		if err := redisCl.Close(); err != nil {
			t.FailNow()
		}
	})

	registry, err := NewApplicationPackagesRegistry(ctx, redisCl, 10*time.Second)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	appPkgIds := &ttnpb.ApplicationPackageAssociationIdentifiers{
		EndDeviceIds: devIDs,
		FPort:        201,
	}
	_, err = registry.SetAssociation(
		ctx, appPkgIds, nil, func(apa *ttnpb.ApplicationPackageAssociation) (
			*ttnpb.ApplicationPackageAssociation, []string, error,
		) {
			return &ttnpb.ApplicationPackageAssociation{
					Ids:         appPkgIds,
					PackageName: "lora-cloud-geolocation-v3",
					Data: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"key": {
								Kind: &structpb.Value_StringValue{
									StringValue: "value",
								},
							},
						},
					},
				}, []string{
					"ids",
					"package_name",
					"data",
				}, nil
		},
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	actual, err := registry.ListAssociations(ctx, devIDs, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(len(actual), should.Equal, 1)

	err = registry.ClearAssociations(ctx, devIDs)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	actual, err = registry.ListAssociations(ctx, devIDs, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(len(actual), should.Equal, 0)
}
