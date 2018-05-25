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
	"fmt"
	"testing"

	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/applicationregistry"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRegistry(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	app, err := r.Create(deepcopy.Copy(pb).(*ttnpb.Application))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(app, should.NotBeNil) {
		pb.CreatedAt = app.Application.GetCreatedAt()
		pb.UpdatedAt = app.Application.GetUpdatedAt()
		a.So(app.Application, should.Resemble, pb)
	}

	i := 0
	err = r.Range(pb, 1, func(found *Application) bool {
		i++
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return true
	}, "ApplicationIdentifiers")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(i, should.Equal, 1)

	updated := ttnpb.NewPopulatedApplication(test.Randy, false)
	for app.Application.ApplicationIdentifiers.Equal(updated.ApplicationIdentifiers) {
		updated = ttnpb.NewPopulatedApplication(test.Randy, false)
	}
	app.Application = updated

	if !a.So(app.Store(), should.BeNil) {
		return
	}

	i = 0
	err = r.Range(pb, 1, func(*Application) bool { i++; return true }, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)

	pb = updated

	i = 0
	err = r.Range(pb, 1, func(found *Application) bool {
		i++
		pb.UpdatedAt = found.Application.GetUpdatedAt()
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return true
	}, "ApplicationIdentifiers")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(i, should.Equal, 1)

	a.So(app.Delete(), should.BeNil)

	i = 0
	err = r.Range(pb, 1, func(*Application) bool { i++; return true }, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)
}

func TestFindApplicationByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	app, err := r.Create(deepcopy.Copy(pb).(*ttnpb.Application))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(app, should.NotBeNil) {
		pb.CreatedAt = app.Application.GetCreatedAt()
		pb.UpdatedAt = app.Application.GetUpdatedAt()
		a.So(app.Application, should.Resemble, pb)
	}

	i := 0
	err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, 1, func(found *Application) bool {
		i++
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return true
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(i, should.Equal, 1)

	updated := ttnpb.NewPopulatedApplication(test.Randy, false)
	for app.Application.ApplicationIdentifiers.Equal(updated.ApplicationIdentifiers) {
		updated = ttnpb.NewPopulatedApplication(test.Randy, false)
	}
	app.Application = updated

	if !a.So(app.Store(), should.BeNil) {
		return
	}

	i = 0
	err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, 1, func(*Application) bool { i++; return true })
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)

	pb = updated

	i = 0
	err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, 1, func(found *Application) bool {
		i++
		pb.UpdatedAt = found.Application.GetUpdatedAt()
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return true
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(i, should.Equal, 1)

	a.So(app.Delete(), should.BeNil)

	i = 0
	err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, 1, func(*Application) bool { i++; return true })
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)
}

func TestFindOneApplicationByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	found, err := FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	app, err := r.Create(deepcopy.Copy(pb).(*ttnpb.Application))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(app, should.NotBeNil) {
		pb.CreatedAt = app.Application.GetCreatedAt()
		pb.UpdatedAt = app.Application.GetUpdatedAt()
		a.So(app.Application, should.Resemble, pb)
	}

	found, err = FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(pretty.Diff(found.Application, pb), should.BeEmpty)

	app, err = r.Create(deepcopy.Copy(pb).(*ttnpb.Application))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(app, should.NotBeNil) {
		pb.CreatedAt = app.Application.GetCreatedAt()
		pb.UpdatedAt = app.Application.GetUpdatedAt()
		a.So(app.Application, should.Resemble, pb)
	}

	found, err = FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)
}

func ExampleRegistry() {
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := &ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "test",
		},
		Description: "My test application",
		UpFormatter: ttnpb.PayloadFormatter_FORMATTER_DEFAULT,
	}

	app, err := r.Create(pb)
	if err != nil {
		panic(fmt.Errorf("Failed to create application %s", err))
	}

	app.Description = "Another description"
	app.UpFormatter = ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
	err = app.Store("Description", "UpFormatter")
	if err != nil {
		panic(fmt.Errorf("Failed to update application %s", err))
	}

	app, err = FindByIdentifiers(r, &ttnpb.ApplicationIdentifiers{ApplicationID: "test"})
	if err != nil {
		panic(fmt.Errorf("Failed to find application by identifiers %s", err))
	}

	err = app.Delete()
	if err != nil {
		panic(fmt.Errorf("Failed to delete application %s", err))
	}
}
