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
	"time"

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

	start := time.Now()

	pb := ttnpb.NewPopulatedApplication(test.Randy, false)

	found, err := FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	app, err := r.Create(deepcopy.Copy(pb).(*ttnpb.Application))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So([]time.Time{start, app.Application.GetCreatedAt(), time.Now()}, should.BeChronological)
	a.So(app.Application.GetCreatedAt(), should.Equal, app.Application.GetUpdatedAt())
	if a.So(app, should.NotBeNil) {
		pb.CreatedAt = app.Application.GetCreatedAt()
		pb.UpdatedAt = app.Application.GetUpdatedAt()
		a.So(app.Application, should.Resemble, pb)
	}

	i := 0
	total, err := r.Range(pb, "", 1, 0, func(found *Application) bool {
		i++
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return false
	}, "ApplicationIdentifiers")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(total, should.Equal, 1)
	a.So(i, should.Equal, 1)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 1, 0, func(found *Application) bool {
		i++
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return false
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(total, should.Equal, 1)
	a.So(i, should.Equal, 1)

	found, err = FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(pretty.Diff(found.Application, pb), should.BeEmpty)

	dev2, err := r.Create(deepcopy.Copy(pb).(*ttnpb.Application))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So([]time.Time{start, dev2.Application.GetCreatedAt(), time.Now()}, should.BeChronological)
	a.So(dev2.Application.GetCreatedAt(), should.Equal, dev2.Application.GetUpdatedAt())
	if a.So(dev2, should.NotBeNil) {
		pb2 := deepcopy.Copy(pb).(*ttnpb.Application)
		pb2.CreatedAt = dev2.Application.GetCreatedAt()
		pb2.UpdatedAt = dev2.Application.GetUpdatedAt()
		a.So(pretty.Diff(dev2.Application, pb2), should.BeEmpty)
	}

	i = 0
	total, err = r.Range(pb, "", 1, 0, func(found *Application) bool { i++; return false }, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 2)
	a.So(i, should.Equal, 1)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 0, 0, func(found *Application) bool { i++; return true })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 2)
	a.So(i, should.Equal, 2)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 0, 0, func(found *Application) bool { i++; return false })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 2)
	a.So(i, should.Equal, 1)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 1, 0, func(found *Application) bool { i++; return false })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 2)
	a.So(i, should.Equal, 1)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 0, 1, func(found *Application) bool { i++; return false })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 2)
	a.So(i, should.Equal, 1)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 1, 1, func(found *Application) bool { i++; return false })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 2)
	a.So(i, should.Equal, 1)

	found, err = FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	err = dev2.Delete()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	updated := ttnpb.NewPopulatedApplication(test.Randy, false)
	for app.Application.ApplicationIdentifiers.Equal(updated.ApplicationIdentifiers) {
		updated = ttnpb.NewPopulatedApplication(test.Randy, false)
	}
	app.Application = updated

	if !a.So(app.Store(), should.BeNil) {
		t.FailNow()
	}

	i = 0
	total, err = r.Range(pb, "", 1, 0, func(*Application) bool { i++; return false }, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 0)
	a.So(i, should.Equal, 0)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 1, 0, func(*Application) bool { i++; return false })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 0)
	a.So(i, should.Equal, 0)

	found, err = FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	pb = updated

	i = 0
	total, err = r.Range(pb, "", 1, 0, func(found *Application) bool {
		i++
		pb.UpdatedAt = found.Application.GetUpdatedAt()
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return false
	}, "ApplicationIdentifiers")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(total, should.Equal, 1)
	a.So(i, should.Equal, 1)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 1, 0, func(found *Application) bool {
		i++
		pb.UpdatedAt = found.Application.GetUpdatedAt()
		a.So(pretty.Diff(found.Application, pb), should.BeEmpty)
		return false
	})
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 1)
	a.So(i, should.Equal, 1)

	found, err = FindByIdentifiers(r, &pb.ApplicationIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(pretty.Diff(found.Application, pb), should.BeEmpty)

	if !a.So(app.Delete(), should.BeNil) {
		t.FailNow()
	}

	i = 0
	total, err = r.Range(pb, "", 1, 0, func(*Application) bool { i++; return false }, "ApplicationIdentifiers")
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 0)
	a.So(i, should.Equal, 0)

	i = 0
	total, err = RangeByIdentifiers(r, &pb.ApplicationIdentifiers, "", 1, 0, func(*Application) bool { i++; return false })
	a.So(err, should.BeNil)
	a.So(total, should.Equal, 0)
	a.So(i, should.Equal, 0)

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
		UpFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
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
