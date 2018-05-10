// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package applicationregistry_test

import (
	"context"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/pkg/applicationregistry"
	"github.com/TheThingsNetwork/ttn/pkg/auth/rights"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	errshould "github.com/TheThingsNetwork/ttn/pkg/errors/should"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.AsApplicationRegistryServer = &RegistryRPC{}

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	_, err := dr.SetApplication(context.Background(), &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, errshould.Describe, ErrPermissionDenied)

	v, err := dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	app, err := dr.GetApplication(context.Background(), &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrPermissionDenied)

	app, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	if a.So(err, should.BeNil) {
		app.CreatedAt = pb.GetCreatedAt()
		app.UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(app, pb), should.BeEmpty)
	}

	_, err = dr.DeleteApplication(context.Background(), &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrPermissionDenied)

	v, err = dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	_, err = dr.GetApplication(context.Background(), &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrPermissionDenied)

	_, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrApplicationNotFound)
}

func TestSetApplicationNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	_, err := dr.SetApplication(context.Background(), &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, errshould.Describe, ErrPermissionDenied)

	v, err := dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, errshould.Describe, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestGetApplicationNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	v, err := dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrApplicationNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestDeleteApplicationNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	v, err := dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrApplicationNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, errshould.Describe, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestCheck(t *testing.T) {
	errTest := &errors.ErrDescriptor{
		MessageFormat: "Test",
		Type:          errors.Internal,
		Code:          1,
	}
	errTest.Register()

	var checkErr error

	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())),
		WithGetApplicationCheck(func(context.Context, *ttnpb.ApplicationIdentifiers) error { return checkErr }),
		WithSetApplicationCheck(func(context.Context, *ttnpb.Application, ...string) error { return checkErr }),
		WithDeleteApplicationCheck(func(context.Context, *ttnpb.ApplicationIdentifiers) error { return checkErr }),
	)).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	t.Run("SetApplication", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		v, err := dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
		a.So(err, errshould.Describe, ErrCheckFailed)
		a.So(v, should.BeNil)

		checkErr = errTest.New(nil)
		v, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
		a.So(err, should.Equal, checkErr)
		a.So(v, should.BeNil)

		checkErr = nil
		v, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
		a.So(err, should.BeNil)
		a.So(v, should.Equal, ttnpb.Empty)
	})

	if !t.Run("GetApplication", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		ret, err := dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
		a.So(err, errshould.Describe, ErrCheckFailed)
		a.So(ret, should.BeNil)

		checkErr = errTest.New(nil)
		ret, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
		a.So(err, should.Equal, checkErr)
		a.So(ret, should.BeNil)

		checkErr = nil
		ret, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		ret.CreatedAt = pb.GetCreatedAt()
		ret.UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(ret, pb), should.BeEmpty)
	}) {
		t.FailNow()
	}

	t.Run("DeleteApplication", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		_, err := dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
		a.So(err, errshould.Describe, ErrCheckFailed)

		checkErr = errTest.New(nil)
		_, err = dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
		a.So(err, should.Equal, checkErr)

		checkErr = nil
		_, err = dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
		a.So(err, should.BeNil)

		_, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
		a.So(err, errshould.Describe, ErrApplicationNotFound)
	})
}
