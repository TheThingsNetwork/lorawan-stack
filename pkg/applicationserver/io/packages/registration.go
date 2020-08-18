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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// ApplicationPackageHandler handles upstream traffic from the Application Server.
type ApplicationPackageHandler interface {
	Package() *ttnpb.ApplicationPackage
	RegisterServices(s *grpc.Server)
	RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn)
	HandleUp(context.Context, *ttnpb.ApplicationPackageDefaultAssociation, *ttnpb.ApplicationPackageAssociation, *ttnpb.ApplicationUp) error
}

var (
	errNotImplemented = errors.DefineUnimplemented("package_not_implemented", "package `{name}` is not implemented")
)
