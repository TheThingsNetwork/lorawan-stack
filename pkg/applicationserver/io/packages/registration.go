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

package packages

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// ApplicationPackageHandler handles upstream traffic from the Application Server.
type ApplicationPackageHandler interface {
	RegisterServices(s *grpc.Server)
	RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn)
	HandleUp(context.Context, *ttnpb.ApplicationPackageAssociation, *ttnpb.ApplicationUp) error
}

// CreateApplicationPackage is a function that creates a traffic handler for a given package.
type CreateApplicationPackage func(io.Server, Registry) ApplicationPackageHandler

type registeredPackage struct {
	ttnpb.ApplicationPackage
	new CreateApplicationPackage
}

var (
	errNotImplemented    = errors.DefineUnimplemented("package_not_implemented", "package `{name}` is not implemented")
	errAlreadyRegistered = errors.DefineAlreadyExists("package_already_registered", "package `{name}` already registered")

	registeredPackages = map[string]*registeredPackage{}
)

// RegisterPackage registers the given package on the application packages frontend.
func RegisterPackage(p ttnpb.ApplicationPackage, new CreateApplicationPackage) error {
	if _, ok := registeredPackages[p.Name]; ok {
		return errAlreadyRegistered.WithAttributes("name", p.Name)
	}
	registeredPackages[p.Name] = &registeredPackage{
		ApplicationPackage: p,
		new:                new,
	}
	return nil
}
