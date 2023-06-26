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

package errors_test

import (
	"context"
	stderrors "errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func Example() {
	errApplicationNotFound := errors.DefineNotFound(
		"application_not_found",
		"Application with ID `{id}` not found",
	// Public attribute "id" is parsed from the message format.
	)
	errCouldNotCreateDevice := errors.Define(
		"could_not_create_device",
		"Could not create Device",
		"right_answer", // right_answer could be some extra attribute (that isn't rendered in the message format)
	)

	findApplication := func(id *ttnpb.ApplicationIdentifiers) (*ttnpb.Application, error) {
		// try really hard, but fail
		return nil, errApplicationNotFound.WithAttributes("id", id.GetApplicationId())
	}

	createDevice := func(dev *ttnpb.EndDevice) error {
		app, err := findApplication(dev.Ids.ApplicationIds)
		if err != nil {
			return err // you can just pass errors up
		}
		// create device
		_ = app
		return nil
	}

	if err := createDevice(&ttnpb.EndDevice{Ids: &ttnpb.EndDeviceIdentifiers{}}); err != nil {
		fmt.Println(errCouldNotCreateDevice.WithCause(err).WithAttributes("right_answer", 42))
	}

	// Output:
	// error:pkg/errors_test:could_not_create_device (Could not create Device)
}

func TestFields(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	errBack := stderrors.New("back")
	errIntermediary := errors.Define("intermediary", "intermediary")
	errFront := errors.Define("front", "front")

	err := errFront.WithCause(errIntermediary.WithCause(errBack))
	fields := err.Fields()
	a.So(fields, should.HaveEmptyDiff, map[string]any{
		"error_cause":       "error:pkg/errors_test:intermediary (intermediary)",
		"error_cause_cause": "back",
	})
}

func TestContextCanceled(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	<-ctx.Done()

	err, ok := errors.From(ctx.Err())
	a.So(ok, should.BeTrue)
	a.So(errors.IsCanceled(err), should.BeTrue)
}

func TestContextDeadlineExceeded(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Millisecond))
	defer cancel()

	<-ctx.Done()

	err, ok := errors.From(ctx.Err())
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(errors.IsDeadlineExceeded(err), should.BeTrue)
}

func TestNetErrors(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	errDummy := fmt.Errorf("dummy")

	for _, tc := range []struct {
		Name     string
		Error    error
		Validate func(err error, e *errors.Error, a *assertions.Assertion)
	}{
		{
			Name: "DNSError",
			Error: &net.DNSError{
				Err:         "SERVFAIL",
				Name:        "invalid-name",
				IsNotFound:  true,
				IsTemporary: false,
				IsTimeout:   false,
				Server:      "dns-server",
			},
			Validate: func(err error, e *errors.Error, a *assertions.Assertion) {
				a.So(e.FullName(), should.Equal, "pkg/errors:net_dns")
				a.So(errors.IsUnavailable(e), should.BeTrue)
				a.So(e.PublicAttributes(), should.Resemble, map[string]any{
					"message":   err.Error(),
					"timeout":   false,
					"not_found": true,
				})
			},
		},
		{
			Name:  "UnknownNetworkError",
			Error: net.UnknownNetworkError("1.1.1.1"),
			Validate: func(err error, e *errors.Error, a *assertions.Assertion) {
				a.So(e.FullName(), should.Equal, "pkg/errors:net_unknown_network")
				a.So(errors.IsNotFound(e), should.BeTrue)
				a.So(e.PublicAttributes(), should.Resemble, map[string]any{
					"message": err.Error(),
					"timeout": false,
				})
			},
		},
		{
			Name:  "InvalidAddrError",
			Error: net.InvalidAddrError("1.1.1.1"),
			Validate: func(err error, e *errors.Error, a *assertions.Assertion) {
				a.So(e.FullName(), should.Equal, "pkg/errors:net_invalid_addr")
				a.So(errors.IsInvalidArgument(e), should.BeTrue)
				a.So(e.PublicAttributes(), should.Resemble, map[string]any{
					"message": err.Error(),
					"timeout": false,
				})
			},
		},
		{
			Name: "AddrError",
			Error: &net.AddrError{
				Addr: "1.1.1.1",
				Err:  "no route",
			},
			Validate: func(err error, e *errors.Error, a *assertions.Assertion) {
				a.So(e.FullName(), should.Equal, "pkg/errors:net_addr")
				a.So(errors.IsUnavailable(e), should.BeTrue)
				a.So(e.PublicAttributes(), should.Resemble, map[string]any{
					"message": err.Error(),
					"timeout": false,
				})
			},
		},
		{
			Name: "OpErrorWithNil",
			Error: &net.OpError{
				Op:     "read",
				Addr:   &net.IPAddr{IP: net.IP{1, 1, 1, 1}},
				Source: &net.IPAddr{IP: net.IP{2, 2, 2, 2}},
				Net:    "0.0.0.0",
				Err:    nil,
			},
			Validate: func(_ error, e *errors.Error, a *assertions.Assertion) {
				a.So(e.FullName(), should.Equal, "pkg/errors:net_operation")
				a.So(errors.IsUnavailable(e), should.BeTrue)
				a.So(e.PublicAttributes(), should.Resemble, map[string]any{
					"timeout": false,
					"address": "1.1.1.1",
					"source":  "2.2.2.2",
					"net":     "0.0.0.0",
					"op":      "read",
				})
			},
		},
		{
			Name: "OpErrorWithErr",
			Error: &net.OpError{
				Op:     "read",
				Addr:   &net.IPAddr{IP: net.IP{1, 1, 1, 1}},
				Source: &net.IPAddr{IP: net.IP{2, 2, 2, 2}},
				Net:    "0.0.0.0",
				Err:    errDummy,
			},
			Validate: func(_ error, e *errors.Error, a *assertions.Assertion) {
				a.So(e.FullName(), should.Equal, "pkg/errors:net_operation")
				a.So(errors.IsUnavailable(e), should.BeTrue)
				a.So(e.PublicAttributes(), should.Resemble, map[string]any{
					"timeout": false,
					"address": "1.1.1.1",
					"source":  "2.2.2.2",
					"net":     "0.0.0.0",
					"op":      "read",
				})
				a.So(e.Cause(), should.Resemble, errDummy)
			},
		},
	} {
		tc := tc // shadow range variable.
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err, ok := errors.From(tc.Error)
			a.So(ok, should.BeTrue)
			tc.Validate(tc.Error, err, a)
		})
	}
}
