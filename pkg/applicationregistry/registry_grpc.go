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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// RegistryRPC implements the application registry gRPC service.
type RegistryRPC struct {
	Interface
	*component.Component

	checks struct {
		GetApplication    func(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error
		SetApplication    func(ctx context.Context, app *ttnpb.Application, fields ...string) error
		DeleteApplication func(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error
	}
}

// RPCOption represents RegistryRPC option
type RPCOption func(*RegistryRPC)

// WithGetApplicationCheck sets a check to GetApplication method of RegistryRPC instance.
// GetApplication first executes fn and if error is returned by it,
// returns error, otherwise execution advances as usual.
func WithGetApplicationCheck(fn func(context.Context, *ttnpb.ApplicationIdentifiers) error) RPCOption {
	return func(r *RegistryRPC) { r.checks.GetApplication = fn }
}

// WithSetApplicationCheck sets a check to SetApplication method of RegistryRPC instance.
// SetApplication first executes fn and if error is returned by it,
// returns error, otherwise execution advances as usual.
func WithSetApplicationCheck(fn func(context.Context, *ttnpb.Application, ...string) error) RPCOption {
	return func(r *RegistryRPC) { r.checks.SetApplication = fn }
}

// WithDeleteApplicationCheck sets a check to DeleteApplication method of RegistryRPC instance.
// DeleteApplication first executes fn and if error is returned by it,
// returns error, otherwise execution advances as usual.
func WithDeleteApplicationCheck(fn func(context.Context, *ttnpb.ApplicationIdentifiers) error) RPCOption {
	return func(r *RegistryRPC) { r.checks.DeleteApplication = fn }
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
	if err := rights.RequireApplication(ctx, id, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	if r.checks.GetApplication != nil {
		if err := r.checks.GetApplication(ctx, id); err != nil {
			if errors.GetType(err) != errors.Unknown {
				return nil, err
			}
			return nil, common.ErrCheckFailed.NewWithCause(nil, err)
		}
	}

	app, err := FindByIdentifiers(r.Interface, id)
	if err != nil {
		return nil, err
	}
	return app.Application, nil
}

// SetApplication sets the application fields to match those of app in underlying registry.
func (r *RegistryRPC) SetApplication(ctx context.Context, req *ttnpb.SetApplicationRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	var fields []string
	if req.FieldMask != nil {
		fields = req.FieldMask.Paths
	}
	if r.checks.SetApplication != nil {
		if err := r.checks.SetApplication(ctx, &req.Application, fields...); err != nil {
			if errors.GetType(err) != errors.Unknown {
				return nil, err
			}
			return nil, common.ErrCheckFailed.NewWithCause(nil, err)
		}
	}

	app, err := FindByIdentifiers(r.Interface, &req.Application.ApplicationIdentifiers)
	notFound := errors.Descriptor(err) == ErrApplicationNotFound
	if err != nil && !notFound {
		return nil, err
	}

	if notFound {
		_, err := r.Interface.Create(&req.Application, fields...)
		return ttnpb.Empty, err
	}
	app.Application = &req.Application
	return ttnpb.Empty, app.Store(fields...)
}

// DeleteApplication deletes the application associated with id from underlying registry.
func (r *RegistryRPC) DeleteApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, id, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	if r.checks.DeleteApplication != nil {
		if err := r.checks.DeleteApplication(ctx, id); err != nil {
			if errors.GetType(err) != errors.Unknown {
				return nil, err
			}
			return nil, common.ErrCheckFailed.NewWithCause(nil, err)
		}
	}

	app, err := FindByIdentifiers(r.Interface, id)
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, app.Delete()
}
