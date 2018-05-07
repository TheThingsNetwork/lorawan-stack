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

	. "github.com/TheThingsNetwork/ttn/pkg/applicationregistry"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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

	found, err := r.FindBy(pb, "ApplicationIdentifiers")
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		found[0].CreatedAt = app.Application.GetCreatedAt()
		found[0].UpdatedAt = app.Application.GetUpdatedAt()
		a.So(pretty.Diff(found[0].Application, pb), should.BeEmpty)
	}

	updated := ttnpb.NewPopulatedApplication(test.Randy, false)
	for app.Application.ApplicationIdentifiers.Equal(updated.ApplicationIdentifiers) {
		updated = ttnpb.NewPopulatedApplication(test.Randy, false)
	}
	updated.CreatedAt = app.Application.GetCreatedAt()
	updated.UpdatedAt = app.Application.GetUpdatedAt()
	app.Application = updated

	if !a.So(app.Store(), should.BeNil) {
		return
	}

	found, err = r.FindBy(pb, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}

	pb = updated

	found, err = r.FindBy(pb, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		found[0].CreatedAt = pb.GetCreatedAt()
		found[0].UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(found[0].Application, pb), should.BeEmpty)
	}

	a.So(app.Delete(), should.BeNil)

	found, err = r.FindBy(pb, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}
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

	found, err := FindApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		a.So(pretty.Diff(found[0].Application, pb), should.BeEmpty)
	}

	updated := ttnpb.NewPopulatedApplication(test.Randy, false)
	for app.Application.ApplicationIdentifiers.Equal(updated.ApplicationIdentifiers) {
		updated = ttnpb.NewPopulatedApplication(test.Randy, false)
	}
	updated.CreatedAt = pb.GetCreatedAt()
	updated.UpdatedAt = pb.GetUpdatedAt()
	app.Application = updated

	if !a.So(app.Store(), should.BeNil) {
		return
	}

	found, err = FindApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}

	pb = updated

	found, err = FindApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		pb.CreatedAt = found[0].Application.GetCreatedAt()
		pb.UpdatedAt = found[0].Application.GetUpdatedAt()
		a.So(pretty.Diff(found[0].Application, pb), should.BeEmpty)
	}

	a.So(app.Delete(), should.BeNil)

	found, err = FindApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}
}

func TestFindOneApplicationByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	found, err := FindOneApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
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

	found, err = FindOneApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
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

	found, err = FindOneApplicationByIdentifiers(r, &pb.ApplicationIdentifiers)
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

	app, err = FindOneApplicationByIdentifiers(r, &ttnpb.ApplicationIdentifiers{ApplicationID: "test"})
	if err != nil {
		panic(fmt.Errorf("Failed to find application by identifiers %s", err))
	}

	err = app.Delete()
	if err != nil {
		panic(fmt.Errorf("Failed to delete application %s", err))
	}
}
