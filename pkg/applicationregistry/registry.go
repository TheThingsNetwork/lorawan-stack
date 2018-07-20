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

// Package applicationregistry contains the implementation of an application registry service.
package applicationregistry

import (
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
//
// Create stores application data and returns a new *Application.
// It may modify CreatedAt and UpdatedAt fields of a and may return error if either of them is set to non-zero value on a.
//
// Range calls f sequentially for each application stored, matching specified application fields.
// If f returns false, Range stops the iteration.
// If orderBy is set to non-empty string, it represents the fieldpath of the field, which the applications, that Range will iterate over will be sorted by.
// If count > 0, then Range will do it's best effort to iterate over at most count applications.
// If count == 0, then Range will iterate over all matching applications.
// Note, that Range provides no guarantees on the count of applications iterated over if count > 0 and
// it's caller's responsibility to handle cases where such are required.
// Range starts iteration at the index specified by the offset. Offset it 0-indexed.
// If len(fields) == 0, then Range uses all fields in a to match applications.
type Interface interface {
	Create(a *ttnpb.Application, fields ...string) (*Application, error)
	Range(a *ttnpb.Application, orderBy string, count, offset uint64, f func(*Application) bool, fields ...string) (uint64, error)
}

var _ Interface = &Registry{}

// Registry is responsible for mapping applications to their identities.
type Registry struct {
	store store.Client
}

// New returns a new Registry with s as an internal Store.
func New(s store.Client) *Registry {
	return &Registry{
		store: s,
	}
}

// Create implements Interface.
// Create modifies CreatedAt and UpdatedAt fields of ed and returns error if either of them is set to non-zero value on ed.
func (r *Registry) Create(a *ttnpb.Application, fields ...string) (app *Application, err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("create").Observe(time.Since(start).Seconds())
	}(time.Now())

	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now
	if len(fields) != 0 {
		fields = append(fields, "CreatedAt", "UpdatedAt")
	}

	id, err := r.store.Create(a, fields...)
	if err != nil {
		return nil, err
	}
	return newApplication(a, r.store, id), nil
}

// Range implements Interface.
func (r *Registry) Range(a *ttnpb.Application, orderBy string, count, offset uint64, f func(*Application) bool, fields ...string) (total uint64, err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		duration := time.Since(start).Seconds()
		latency.WithLabelValues("range").Observe(duration)
		rangeLatency.WithLabelValues(strings.Join(fields, ",")).Observe(duration)
	}(time.Now())

	if a == nil {
		return 0, errNilApplication
	}
	return r.store.Range(
		a,
		func() interface{} { return &ttnpb.Application{} },
		orderBy, count, offset,
		func(k store.PrimaryKey, v interface{}) bool {
			return f(newApplication(v.(*ttnpb.Application), r.store, k))
		},
		fields...,
	)
}

// Identifiers supported in RangeByIdentifiers.
var Identifiers = []string{
	"ApplicationIdentifiers.ApplicationID",
}

// RangeByIdentifiers is a helper function, which allows ranging over r by matching identifiers instead of *ttnpb.Application.
func RangeByIdentifiers(r Interface, id *ttnpb.ApplicationIdentifiers, orderBy string, count, offset uint64, f func(*Application) bool) (uint64, error) {
	if id == nil {
		return 0, errNilIdentifiers
	}

	fields := make([]string, 0, 1)
	if id.ApplicationID != "" {
		fields = append(fields, "ApplicationIdentifiers.ApplicationID")
	}
	return r.Range(&ttnpb.Application{ApplicationIdentifiers: *id}, orderBy, count, offset, f, fields...)
}

// FindByIdentifiers searches for exactly one application matching specified application identifiers in r.
func FindByIdentifiers(r Interface, id *ttnpb.ApplicationIdentifiers) (*Application, error) {
	var app *Application
	total, err := RangeByIdentifiers(r, id, "", 1, 0, func(a *Application) bool {
		app = a
		return false
	})
	if err != nil {
		return nil, err
	}

	switch {
	case total == 0:
		return nil, errApplicationNotFound
	case total > 1:
		return nil, errTooManyApplications
	}
	return app, nil
}
