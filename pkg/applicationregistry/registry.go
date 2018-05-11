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

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Interface represents the interface exposed by the *Registry.
type Interface interface {
	Create(a *ttnpb.Application, fields ...string) (*Application, error)
	FindBy(a *ttnpb.Application, fields ...string) ([]*Application, error)
}

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

// FindBy searches for applications matching specified application fields in underlying store.Interface.
// The returned slice contains unique applications, matching at least one of values in a.
func (r *Registry) FindBy(a *ttnpb.Application, fields ...string) ([]*Application, error) {
	if a == nil {
		return nil, errors.New("Application specified is nil")
	}

	found, err := r.store.FindBy(a, func() interface{} { return &ttnpb.Application{} }, fields...)
	if err != nil {
		return nil, err
	}

	applications := make([]*Application, 0, len(found))
	for id, a := range found {
		applications = append(applications, newApplication(a.(*ttnpb.Application), r.store, id))
	}
	return applications, nil
}

// FindApplicationByIdentifiers searches for applications matching specified application identifiers in r.
func FindApplicationByIdentifiers(r Interface, id *ttnpb.ApplicationIdentifiers) ([]*Application, error) {
	if id == nil {
		return nil, errors.New("Identifiers specified are nil")
	}

	fields := []string{}
	switch {
	case id.ApplicationID != "":
		fields = append(fields, "ApplicationIdentifiers.ApplicationID")
	}
	return r.FindBy(&ttnpb.Application{ApplicationIdentifiers: *id}, fields...)
}

// FindOneApplicationByIdentifiers searches for exactly one application matching specified application identifiers in r.
func FindOneApplicationByIdentifiers(r Interface, id *ttnpb.ApplicationIdentifiers) (*Application, error) {
	apps, err := FindApplicationByIdentifiers(r, id)
	if err != nil {
		return nil, err
	}
	switch len(apps) {
	case 0:
		return nil, ErrApplicationNotFound.New(nil)
	case 1:
		return apps[0], nil
	default:
		return nil, ErrTooManyApplications.New(nil)
	}
}
