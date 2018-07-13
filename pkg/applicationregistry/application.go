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

package applicationregistry

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Application represents the application stored in the registry.
type Application struct {
	*ttnpb.Application
	key   store.PrimaryKey
	store store.Client
}

func newApplication(a *ttnpb.Application, s store.Client, k store.PrimaryKey) *Application {
	return &Application{
		Application: a,
		store:       s,
		key:         k,
	}
}

// Store updates applications data in the underlying store.Interface.
// It modifies the UpdatedAt field of a.Application.
func (a *Application) Store(fields ...string) (err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("update").Observe(time.Since(start).Seconds())
	}(time.Now())

	a.Application.UpdatedAt = time.Now().UTC()
	if len(fields) != 0 {
		fields = append(fields, "UpdatedAt")
	}
	return a.store.Update(a.key, a.Application, fields...)
}

// Load returns a snapshot of current application data in underlying store.Interface.
func (a *Application) Load() (app *Application, err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("load").Observe(time.Since(start).Seconds())
	}(time.Now())

	pb := &ttnpb.Application{}
	if err := a.store.Find(a.key, pb); err != nil {
		return nil, err
	}
	return newApplication(pb, a.store, a.key), nil
}

// Delete removes application from the underlying store.Interface.
func (a *Application) Delete() (err error) {
	defer func(start time.Time) {
		if err != nil {
			return
		}
		latency.WithLabelValues("delete").Observe(time.Since(start).Seconds())
	}(time.Now())

	return a.store.Delete(a.key)
}
