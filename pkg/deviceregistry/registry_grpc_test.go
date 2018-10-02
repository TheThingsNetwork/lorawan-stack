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

package deviceregistry_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	uri              = "foo/bar"
	host             = "test"
	defaultListCount = 2
)

func init() {
	SetDefaultListCount(defaultListCount)
}

func newContext(md *rpcmetadata.MD, s grpc.ServerTransportStream, rs ...ttnpb.Right) context.Context {
	ctx := rights.NewContextWithFetcher(
		test.Context(),
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
			return &ttnpb.Rights{Rights: rs}, nil
		}),
	)
	if s != nil {
		ctx = grpc.NewContextWithServerTransportStream(ctx, s)
	}
	if md != nil {
		ctx = md.ToIncomingContext(ctx)
	}
	return ctx
}

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	ctx := newContext(
		nil,
		&test.MockServerTransportStream{
			MockServerStream: &test.MockServerStream{
				SetHeaderFunc: func(md metadata.MD) error {
					return nil
				},
				SendHeaderFunc: func(md metadata.MD) error {
					t.Fatal("SendHeader must not be called")
					return errors.New("SendHeader must not be called")
				},
				SetTrailerFunc: func(md metadata.MD) {
					t.Fatal("SetTrailer must not be called")
				},
			},
		},
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
		ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	)

	dev, err := dr.Set(newContext(nil, nil), &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(dev, should.BeNil)

	dev, err = dr.Set(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(dev, should.Resemble, pb) {
		pretty.Ldiff(t, dev, pb)
	}

	devs, err := dr.List(newContext(nil, nil), &ttnpb.ListEndDevicesRequest{ApplicationIdentifiers: pb.EndDeviceIdentifiers.ApplicationIdentifiers})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(devs, should.BeNil)

	devs, err = dr.List(ctx, &ttnpb.ListEndDevicesRequest{ApplicationIdentifiers: pb.EndDeviceIdentifiers.ApplicationIdentifiers})
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = pb.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], pb), should.BeEmpty)
	}

	_, err = dr.Delete(newContext(nil, nil), &pb.EndDeviceIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	v, err := dr.Delete(ctx, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	_, err = dr.List(newContext(nil, nil), &ttnpb.ListEndDevicesRequest{ApplicationIdentifiers: pb.EndDeviceIdentifiers.ApplicationIdentifiers})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	devs, err = dr.List(ctx, &ttnpb.ListEndDevicesRequest{ApplicationIdentifiers: pb.EndDeviceIdentifiers.ApplicationIdentifiers})
	if a.So(err, should.BeNil) && a.So(devs, should.NotBeNil) {
		a.So(devs.EndDevices, should.BeEmpty)
	}
}

func TestSetDeviceNoProcessor(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := newContext(nil, nil, ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	_, err := dr.Set(newContext(nil, nil), &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	dev, err := dr.Set(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.BeNil)
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	if !a.So(dev, should.Resemble, pb) {
		pretty.Ldiff(t, dev, pb)
	}

	old := deepcopy.Copy(pb).(*ttnpb.EndDevice)

	for pb.GetFormatters().GetUpFormatter() == old.GetFormatters().GetUpFormatter() ||
		pb.GetFormatters().GetUpFormatterParameter() == old.GetFormatters().GetUpFormatterParameter() {
		pb.Formatters = ttnpb.NewPopulatedMessagePayloadFormatters(test.Randy, false)
	}

	dev, err = dr.Set(ctx, &ttnpb.SetDeviceRequest{
		Device: *pb,
		FieldMask: pbtypes.FieldMask{
			Paths: []string{"formatters.up_formatter", "formatters.up_formatter_parameter"},
		},
	})
	a.So(err, should.BeNil)
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	if !a.So(dev, should.Resemble, pb) {
		pretty.Ldiff(t, dev, pb)
	}

	if old.Formatters == nil {
		old.Formatters = &ttnpb.MessagePayloadFormatters{}
	}
	old.Formatters.UpFormatter = pb.GetFormatters().GetUpFormatter()
	old.Formatters.UpFormatterParameter = pb.GetFormatters().GetUpFormatterParameter()
	pb = old

	got, err := FindByIdentifiers(dr.Interface, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	pb.CreatedAt = got.GetCreatedAt()
	pb.UpdatedAt = got.GetUpdatedAt()
	a.So(pretty.Diff(got.EndDevice, pb), should.BeEmpty)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	dev, err = dr.Set(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyDevices)
	a.So(dev, should.BeNil)
}

func TestListDevices(t *testing.T) {
	for _, tc := range []struct {
		Host, URI   string
		Limit, Page uint64
		ShouldList  bool
		Headers     map[string][]string
	}{
		{"", "", 0, 0, true, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 1, 0, true, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 0, 1, true, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 1, 1, true, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 2, 0, true, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 2, 1, true, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 0, 2, false, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 1, 2, false, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"", "", 2, 2, false, map[string][]string{
			"x-total-count": {"1"},
		}},
		{"foohost", "", 0, 0, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, defaultListCount),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, defaultListCount),
			},
		}},
		{"foohost", "", 1, 0, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, 1),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, 1),
			},
		}},
		{"foohost", "", 0, 1, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, defaultListCount),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, defaultListCount),
			},
		}},
		{"foohost", "", 1, 1, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, 1),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, 1),
			},
		}},
		{"foohost", "", 2, 0, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, 2),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, 2),
			},
		}},
		{"foohost", "", 2, 1, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, 2),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, 2),
			},
		}},
		{"foohost", "", 0, 2, false, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, defaultListCount),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, defaultListCount),
			},
		}},
		{"foohost", "", 1, 2, false, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, 1),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, 1),
			},
		}},
		{"foohost", "", 2, 2, false, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="first"`, 2),
				fmt.Sprintf(`<foohost?page=1&per_page=%d>; rel="last"`, 2),
			},
		}},
		{"foohost", "bar/42", 0, 0, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, defaultListCount),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, defaultListCount),
			},
		}},
		{"foohost", "bar/42", 1, 0, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, 1),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, 1),
			},
		}},
		{"foohost", "bar/42", 0, 1, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, defaultListCount),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, defaultListCount),
			},
		}},
		{"foohost", "bar/42", 1, 1, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, 1),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, 1),
			},
		}},
		{"foohost", "bar/42", 2, 0, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, 2),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, 2),
			},
		}},
		{"foohost", "bar/42", 2, 1, true, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, 2),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, 2),
			},
		}},
		{"foohost", "bar/42", 0, 2, false, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, defaultListCount),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, defaultListCount),
			},
		}},
		{"foohost", "bar/42", 1, 2, false, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, 1),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, 1),
			},
		}},
		{"foohost", "bar/42", 2, 2, false, map[string][]string{
			"x-total-count": {"1"},
			"link": {
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="first"`, 2),
				fmt.Sprintf(`<foohost/bar/42?page=1&per_page=%d>; rel="last"`, 2),
			},
		}},
	} {
		t.Run(fmt.Sprintf("host:'%s'/uri:'%s'/limit:%d/page:%d", tc.Host, tc.URI, tc.Limit, tc.Page), func(t *testing.T) {
			a := assertions.New(t)
			dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

			setHeaderMD := make(chan metadata.MD, 1)
			setHeaderErr := make(chan error, 1)

			ctx := newContext(
				&rpcmetadata.MD{
					Limit: tc.Limit,
					Page:  tc.Page,
					URI:   tc.URI,
					Host:  tc.Host,
				},
				&test.MockServerTransportStream{
					MockServerStream: &test.MockServerStream{
						SetHeaderFunc: func(md metadata.MD) error {
							setHeaderMD <- md
							return <-setHeaderErr
						},
						SendHeaderFunc: func(md metadata.MD) error {
							t.Fatal("SendHeader must not be called")
							return errors.New("SendHeader must not be called")
						},
						SetTrailerFunc: func(md metadata.MD) {
							t.Fatal("SetTrailer must not be called")
						},
					},
				},
				ttnpb.RIGHT_APPLICATION_DEVICES_READ,
			)

			dev, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			select {
			case setHeaderErr <- nil:
			case <-time.After(test.Delay):
				t.Fatal("Timed out waiting for error to be consumed by SetHeader")
			}

			devs, err := dr.List(ctx, &ttnpb.ListEndDevicesRequest{ApplicationIdentifiers: dev.EndDeviceIdentifiers.ApplicationIdentifiers})
			a.So(err, should.BeNil)
			if tc.ShouldList && a.So(devs.EndDevices, should.HaveLength, 1) {
				a.So(pretty.Diff(devs.EndDevices[0], dev.EndDevice), should.BeEmpty)
			}

			select {
			case md := <-setHeaderMD:
				a.So(md, should.HaveSameElementsDeep, tc.Headers)

			case <-time.After(test.Delay):
				t.Fatal("Timed out waiting for SetHeader to be called")
			}
		})
	}
}

func TestGetDevice(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := newContext(nil, nil, ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.Get(ctx, &ttnpb.GetEndDeviceRequest{EndDeviceIdentifiers: pb.EndDeviceIdentifiers})
	a.So(err, should.EqualErrorOrDefinition, ErrDeviceNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.Get(ctx, &ttnpb.GetEndDeviceRequest{EndDeviceIdentifiers: pb.EndDeviceIdentifiers})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.Get(ctx, &ttnpb.GetEndDeviceRequest{EndDeviceIdentifiers: pb.EndDeviceIdentifiers})
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestDeleteDevice(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := newContext(nil, nil, ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.Delete(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrDeviceNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.Delete(ctx, &pb.EndDeviceIdentifiers)
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

	v, err = dr.Delete(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestSetDeviceProcessor(t *testing.T) {
	errTest := errors.DefineInternal("test", "test")

	var procErr error

	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())),
		WithSetDeviceProcessor(func(_ context.Context, _ bool, dev *ttnpb.EndDevice, fields ...string) (*ttnpb.EndDevice, []string, error) {
			if procErr != nil {
				return nil, nil, procErr
			}
			return dev, fields, nil
		}),
	)).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	a := assertions.New(t)

	dev, err := dr.Set(newContext(nil, nil), &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(dev, should.BeNil)

	ctx := newContext(nil, nil, ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE)

	procErr = errors.New("err")
	dev, err = dr.Set(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.HaveSameErrorDefinitionAs, ErrProcessor)
	a.So(dev, should.BeNil)

	procErr = errTest.WithAttributes("foo", "bar")
	dev, err = dr.Set(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.Resemble, procErr)
	a.So(dev, should.BeNil)

	procErr = nil
	dev, err = dr.Set(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.BeNil)
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	a.So(dev, should.Resemble, pb)
}
