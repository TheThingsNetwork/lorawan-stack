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

package identityserver

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// specializers is the variable where specializations must be registered.
var specializers = make(specializersRegistry)

func init() {
	// These are the default specializers per entity that are used if no
	// specializer ID is specified. These does not make any specialization as
	// returns the same base given ttnpb.XXX types.
	specializers.RegisterUser("", func(base ttnpb.User) store.User { return &base })
	specializers.RegisterApplication("", func(base ttnpb.Application) store.Application { return &base })
	specializers.RegisterGateway("", func(base ttnpb.Gateway) store.Gateway { return &base })
	specializers.RegisterClient("", func(base ttnpb.Client) store.Client { return &base })
	specializers.RegisterOrganization("", func(base ttnpb.Organization) store.Organization { return &base })
}

// The registry where all specializers will be registered on `init` calls using
// the `RegisterXXX` methods and retrieved using `GetXXX` methods. It panics if
// a specializer with duplicated ID is tried to be registered (IDs are namespaced
// by specializer type).
type specializersRegistry map[string]interface{}

func (r specializersRegistry) makeID(entity, id string) string {
	return fmt.Sprintf("%s:%s", entity, id)
}

func (r specializersRegistry) RegisterUser(id string, s store.UserSpecializer) {
	r.register(r.makeID("user", id), s)
}

func (r specializersRegistry) RegisterApplication(id string, s store.ApplicationSpecializer) {
	r.register(r.makeID("application", id), s)
}

func (r specializersRegistry) RegisterGateway(id string, s store.GatewaySpecializer) {
	r.register(r.makeID("gateway", id), s)
}

func (r specializersRegistry) RegisterClient(id string, s store.ClientSpecializer) {
	r.register(r.makeID("client", id), s)
}

func (r specializersRegistry) RegisterOrganization(id string, s store.OrganizationSpecializer) {
	r.register(r.makeID("organization", id), s)
}

func (r specializersRegistry) register(id string, s interface{}) {
	if _, exists := r[id]; exists {
		parts := strings.SplitN(id, ":", 2)
		panic(errors.Errorf("Another (%s) specializer with id `%s` is already registered", parts[0], parts[1]))
	}
	r[id] = s
}

func (r specializersRegistry) GetUser(id string) (store.UserSpecializer, error) {
	s, exists := r[r.makeID("user", id)].(store.UserSpecializer)
	if !exists {
		return nil, errors.Errorf("User specializer with id `%s` is not registered", id)
	}
	return s, nil
}

func (r specializersRegistry) GetApplication(id string) (store.ApplicationSpecializer, error) {
	s, exists := r[r.makeID("application", id)].(store.ApplicationSpecializer)
	if !exists {
		return nil, errors.Errorf("Application specializer with id `%s` is not registered", id)
	}
	return s, nil
}
func (r specializersRegistry) GetGateway(id string) (store.GatewaySpecializer, error) {
	s, exists := r[r.makeID("gateway", id)].(store.GatewaySpecializer)
	if !exists {
		return nil, errors.Errorf("Gateway specializer with id `%s` is not registered", id)
	}
	return s, nil
}
func (r specializersRegistry) GetClient(id string) (store.ClientSpecializer, error) {
	s, exists := r[r.makeID("client", id)].(store.ClientSpecializer)
	if !exists {
		return nil, errors.Errorf("Client specializer with id `%s` is not registered", id)
	}
	return s, nil
}
func (r specializersRegistry) GetOrganization(id string) (store.OrganizationSpecializer, error) {
	s, exists := r[r.makeID("organization", id)].(store.OrganizationSpecializer)
	if !exists {
		return nil, errors.Errorf("Organization specializer with id `%s` is not registered", id)
	}
	return s, nil
}
