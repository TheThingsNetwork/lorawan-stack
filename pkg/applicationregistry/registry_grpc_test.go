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
	removetheseerrors "go.thethings.network/lorawan-stack/pkg/errors"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

var _ ttnpb.AsApplicationRegistryServer = &RegistryRPC{}

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	app, err := dr.SetApplication(context.Background(), &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	pb.CreatedAt = app.GetCreatedAt()
	pb.UpdatedAt = app.GetUpdatedAt()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(app, should.Resemble, pb) {
		pretty.Ldiff(t, app, pb)
	}

	app, err = dr.GetApplication(context.Background(), &pb.ApplicationIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if !a.So(app, should.Resemble, pb) {
		pretty.Ldiff(t, app, pb)
	}

	v, err := dr.DeleteApplication(context.Background(), &pb.ApplicationIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(v, should.BeNil)

	v, err = dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	app, err = dr.GetApplication(context.Background(), &pb.ApplicationIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, should.DescribeError, ErrApplicationNotFound)
	a.So(app, should.BeNil)
}

func TestSetApplicationNoProcessor(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)
	a.So(validate.ID(pb.GetApplicationID()), should.BeNil)

	_, err := dr.SetApplication(context.Background(), &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	v, err := dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.DescribeError, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestGetApplication(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	v, err := dr.GetApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, should.DescribeError, ErrApplicationNotFound)
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
	a.So(err, should.DescribeError, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestDeleteApplication(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	v, err := dr.DeleteApplication(ctx, &pb.ApplicationIdentifiers)
	a.So(err, should.DescribeError, ErrApplicationNotFound)
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
	a.So(err, should.DescribeError, ErrTooManyApplications)
	a.So(v, should.BeNil)
}

func TestSetApplicationProcessor(t *testing.T) {
	errTest := &removetheseerrors.ErrDescriptor{
		MessageFormat: "Test",
		Type:          removetheseerrors.Internal,
		Code:          1,
	}
	errTest.Register()

	var procErr error

	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())),
		WithSetApplicationProcessor(func(_ context.Context, _ bool, dev *ttnpb.Application, fields ...string) (*ttnpb.Application, []string, error) {
			if procErr != nil {
				return nil, nil, procErr
			}
			return dev, fields, nil
		}),
	)).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
	})

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	a := assertions.New(t)

	procErr = removetheseerrors.New("err")
	app, err := dr.SetApplication(context.Background(), &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(app, should.BeNil)

	app, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: ttnpb.Application{}})
	a.So(errors.IsInvalidArgument(err), should.BeTrue)
	a.So(app, should.BeNil)

	procErr = errTest.New(nil)
	app, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.Equal, procErr)
	a.So(app, should.BeNil)

	procErr = nil
	app, err = dr.SetApplication(ctx, &ttnpb.SetApplicationRequest{Application: *pb})
	a.So(err, should.BeNil)
	pb.CreatedAt = app.GetCreatedAt()
	pb.UpdatedAt = app.GetUpdatedAt()
	a.So(app, should.Resemble, pb)
}
