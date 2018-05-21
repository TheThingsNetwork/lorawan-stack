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
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
type Interface interface {
	Create(a *ttnpb.Application, fields ...string) (*Application, error)
	Range(a *ttnpb.Application, batchSize uint64, f func(*Application) bool, fields ...string) error
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

// Create stores applications data in underlying store.Interface and returns a new *Application.
// It modifies CreatedAt and UpdatedAt fields of a and returns error if either of them is non-zero on a.
func (r *Registry) Create(a *ttnpb.Application, fields ...string) (*Application, error) {
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

// Range calls f sequentially for each application stored, matching specified application fields.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Registry's
// contents: no application will be visited more than once, but if the device is
// created or deleted concurrently, Range may or may not call f on that device.
//
// If batchSize argument is non-zero, Range will retrieve applications
// from the underlying store in chunks of (approximately) batchSize applications.
//
// If len(fields) == 0, then Range uses all fields in a to match applications.
func (r *Registry) Range(a *ttnpb.Application, batchSize uint64, f func(*Application) bool, fields ...string) error {
	if a == nil {
		return errors.New("Application specified is nil")
	}
	return r.store.Range(
		a,
		func() interface{} { return &ttnpb.Application{} },
		batchSize,
		func(k store.PrimaryKey, v interface{}) bool {
			return f(newApplication(v.(*ttnpb.Application), r.store, k))
		},
		fields...,
	)
}

// Range calls f sequentially for each application stored in r, matching specified application identifiers.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Intefaces's
// contents: no application will be visited more than once, but if the device is
// created or deleted concurrently, Range may or may not call f on that device.
//
// If batchSize argument is non-zero, Range will retrieve applications
// from the underlying store in chunks of (approximately) batchSize applications.
func RangeByIdentifiers(r Interface, id *ttnpb.ApplicationIdentifiers, batchSize uint64, f func(*Application) bool) error {
	if id == nil {
		return errors.New("Identifiers specified are nil")
	}

	fields := []string{}
	switch {
	case id.ApplicationID != "":
		fields = append(fields, "ApplicationIdentifiers.ApplicationID")
	}
	return r.Range(&ttnpb.Application{ApplicationIdentifiers: *id}, batchSize, f, fields...)
}

// FindByIdentifiers searches for exactly one application matching specified application identifiers in r.
func FindByIdentifiers(r Interface, id *ttnpb.ApplicationIdentifiers) (*Application, error) {
	var app *Application
	var i uint64
	err := RangeByIdentifiers(r, id, 1, func(a *Application) bool {
		i++
		if i > 1 {
			return false
		}
		app = a
		return true
	})
	if err != nil {
		return nil, err
	}
	switch i {
	case 0:
		return nil, ErrApplicationNotFound.New(nil)
	case 1:
		return app, nil
	default:
		return nil, ErrTooManyApplications.New(nil)
	}
}
