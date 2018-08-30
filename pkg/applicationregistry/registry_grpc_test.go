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

	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/applicationregistry"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

var _ ttnpb.AsApplicationRegistryServer = &RegistryRPC{}

var (
	ctxWithoutRights = rights.NewContextWithFetcher(
		test.Context(),
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
			return nil, nil
		}),
	)
	ctxWithRights = rights.NewContextWithFetcher(
		test.Context(),
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
			return ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC), nil
		}),
	)
)

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	ar := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	app, err := ar.SetApplication(ctxWithoutRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = ar.SetApplication(ctxWithRights, &ttnpb.SetApplicationRequest{Application: *pb})
	pb.CreatedAt = app.GetCreatedAt()
	pb.UpdatedAt = app.GetUpdatedAt()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(app, should.Resemble, pb) {
		pretty.Ldiff(t, app, pb)
	}

	app, err = ar.GetApplication(ctxWithoutRights, &pb.ApplicationIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = ar.GetApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if !a.So(app, should.Resemble, pb) {
		pretty.Ldiff(t, app, pb)
	}

	v, err := ar.DeleteApplication(ctxWithoutRights, &pb.ApplicationIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(v, should.BeNil)

	v, err = ar.DeleteApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	app, err = ar.GetApplication(ctxWithoutRights, &pb.ApplicationIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = ar.GetApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrApplicationNotFound)
	a.So(app, should.BeNil)
}

func TestSetApplicationNoProcessor(t *testing.T) {
	a := assertions.New(t)
	ar := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)
	a.So(validate.ID(pb.GetApplicationID()), should.BeNil)

	_, err := ar.SetApplication(ctxWithoutRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	v, err := ar.SetApplication(ctxWithRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = ar.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = ar.SetApplication(ctxWithRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestGetApplication(t *testing.T) {
	a := assertions.New(t)
	ar := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	v, err := ar.GetApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrApplicationNotFound)
	a.So(v, should.BeNil)

	_, err = ar.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = ar.GetApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = ar.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = ar.GetApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestDeleteApplication(t *testing.T) {
	a := assertions.New(t)
	ar := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	v, err := ar.DeleteApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrApplicationNotFound)
	a.So(v, should.BeNil)

	_, err = ar.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = ar.DeleteApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = ar.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = ar.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = ar.DeleteApplication(ctxWithRights, &pb.ApplicationIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestSetApplicationProcessor(t *testing.T) {
	errTest := errors.DefineInternal("test", "test")

	var procErr error

	ar := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())),
		WithSetApplicationProcessor(func(_ context.Context, _ bool, app *ttnpb.Application, fields ...string) (*ttnpb.Application, []string, error) {
			if procErr != nil {
				return nil, nil, procErr
			}
			return app, fields, nil
		}),
	)).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	a := assertions.New(t)

	app, err := ar.SetApplication(ctxWithoutRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	procErr = errors.New("err")
	app, err = ar.SetApplication(ctxWithRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.HaveSameErrorDefinitionAs, ErrProcessorFailed)
	a.So(app, should.BeNil)

	procErr = errTest.WithAttributes("foo", "bar")
	app, err = ar.SetApplication(ctxWithRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.Resemble, procErr)
	a.So(app, should.BeNil)

	procErr = nil
	app, err = ar.SetApplication(ctxWithRights, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.BeNil)
	pb.CreatedAt = app.GetCreatedAt()
	pb.UpdatedAt = app.GetUpdatedAt()
	a.So(app, should.Resemble, pb)
}
