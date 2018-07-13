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
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/gogoproto"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// RegistryRPC implements the application registry gRPC service.
type RegistryRPC struct {
	Interface
	*component.Component

	setApplicationProcessor func(ctx context.Context, create bool, app *ttnpb.Application, fields ...string) (*ttnpb.Application, []string, error)
}

// RPCOption represents RegistryRPC option
type RPCOption func(*RegistryRPC)

// WithSetApplicationProcessor sets a function, which checks and processes the application and fields,
// which are about to be passed to SetApplication method of RegistryRPC instance.
// After a successful search, SetApplication passes request context, bool, indicating whether the request will trigger a 'Create' or an 'Update',
// application, which is about to be passed to the underlying registry and converted field paths(if such are specified in the request).
// If nil error is returned by fn, SetApplication passes the application and fields returned to the underlying registry,
// otherwise SetApplication returns the error without modifying the registry.
func WithSetApplicationProcessor(fn func(ctx context.Context, create bool, app *ttnpb.Application, fields ...string) (*ttnpb.Application, []string, error)) RPCOption {
	return func(r *RegistryRPC) { r.setApplicationProcessor = fn }
}

// NewRPC returns a new instance of RegistryRPC
func NewRPC(c *component.Component, r Interface, opts ...RPCOption) (*RegistryRPC, error) {
	rpc := &RegistryRPC{
		Component: c,
		Interface: r,
	}

	for _, opt := range opts {
		opt(rpc)
	}

	hook, err := c.RightsHook()
	if err != nil {
		return nil, err
	}
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsApplicationRegistry", rights.HookName, hook.UnaryHook())

	return rpc, nil
}

// GetApplication returns the application associated with id in underlying registry, if found.
func (r *RegistryRPC) GetApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) (*ttnpb.Application, error) {
	if err := rights.RequireApplication(ctx, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	app, err := FindByIdentifiers(r.Interface, id)
	if err != nil {
		return nil, err
	}
	return app.Application, nil
}

// SetApplication sets the application fields to match those of app in underlying registry.
func (r *RegistryRPC) SetApplication(ctx context.Context, req *ttnpb.SetApplicationRequest) (*ttnpb.Application, error) {
	if err := rights.RequireApplication(ctx, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	var fields []string
	if req.FieldMask != nil {
		fields = gogoproto.GoFieldsPaths(req.FieldMask, req.Application)
	}

	app, err := FindByIdentifiers(r.Interface, &req.Application.ApplicationIdentifiers)
	notFound := errors.IsNotFound(err)
	if err != nil && !notFound {
		return nil, err
	}

	setApp := &req.Application
	if r.setApplicationProcessor != nil {
		setApp, fields, err = r.setApplicationProcessor(ctx, notFound, setApp, fields...)
		if err != nil && !errors.IsUnknown(err) {
			return nil, err
		} else if err != nil {
			return nil, errProcessorFailed.WithCause(err)
		}
	}

	if notFound {
		app, err := r.Interface.Create(setApp, fields...)
		if err != nil {
			return nil, err
		}
		events.Publish(evtCreateApplication(ctx, setApp.ApplicationIdentifiers, nil))
		return app.Application, nil
	}

	app.Application = setApp
	if err = app.Store(fields...); err != nil {
		return nil, err
	}
	events.Publish(evtUpdateApplication(ctx, setApp.ApplicationIdentifiers, fields))
	return app.Application, nil
}

// DeleteApplication deletes the application associated with id from underlying registry.
func (r *RegistryRPC) DeleteApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	app, err := FindByIdentifiers(r.Interface, id)
	if err != nil {
		return nil, err
	}

	if err = app.Delete(); err != nil {
		return nil, err
	}
	events.Publish(evtDeleteApplication(ctx, id, nil))
	return ttnpb.Empty, nil
}
