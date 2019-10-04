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
	"strconv"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// appendImplicitAssociationsGetPaths appends implicit ttnpb.ApplicationPackageAssociation get paths to paths.
func appendImplicitAssociationsGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 1+len(paths)),
		"package_name",
	), paths...)
}

// List implements ttnpb.ApplicationPackageRegistryServer.
func (s *server) List(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationPackages, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_PACKAGES); err != nil {
		return nil, err
	}
	var packages ttnpb.ApplicationPackages
	for _, p := range registeredPackages {
		packages.Packages = append(packages.Packages, &p.ApplicationPackage)
	}
	return &packages, nil
}

// GetAssociation implements ttnpb.ApplicationPackageRegistryServer.
func (s *server) GetAssociation(ctx context.Context, req *ttnpb.GetApplicationPackageAssociationRequest) (*ttnpb.ApplicationPackageAssociation, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_PACKAGES); err != nil {
		return nil, err
	}
	return s.registry.Get(ctx, req.ApplicationPackageAssociationIdentifiers, appendImplicitAssociationsGetPaths(req.FieldMask.Paths...))
}

// ListAssociations implements tnpb.ApplicationPackageRegistryServer.
func (s *server) ListAssociations(ctx context.Context, req *ttnpb.ListApplicationPackageAssociationRequest) (assoc *ttnpb.ApplicationPackageAssociations, err error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_PACKAGES); err != nil {
		return nil, err
	}
	var total int64
	ctx = s.registry.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	associations, err := s.registry.List(ctx, req.EndDeviceIdentifiers, appendImplicitAssociationsGetPaths(req.FieldMask.Paths...))
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationPackageAssociations{
		Associations: associations,
	}, nil
}

// SetAssociation implements ttnpb.ApplicationPackageRegistryServer.
func (s *server) SetAssociation(ctx context.Context, req *ttnpb.SetApplicationPackageAssociationRequest) (*ttnpb.ApplicationPackageAssociation, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_PACKAGES); err != nil {
		return nil, err
	}
	return s.registry.Set(ctx, req.ApplicationPackageAssociationIdentifiers, appendImplicitAssociationsGetPaths(req.FieldMask.Paths...),
		func(assoc *ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error) {
			if assoc != nil {
				return &req.ApplicationPackageAssociation, req.FieldMask.Paths, nil
			}
			return &req.ApplicationPackageAssociation, append(req.FieldMask.Paths,
				"ids.end_device_ids",
				"ids.f_port",
			), nil
		},
	)
}

// DeleteAssociation implements ttnpb.ApplicationPackageRegistryServer.
func (s *server) DeleteAssociation(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_PACKAGES); err != nil {
		return nil, err
	}
	_, err := s.registry.Set(ctx, *ids, nil,
		func(assoc *ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error) {
			return nil, nil, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func setTotalHeader(ctx context.Context, total int64) {
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(total, 10)))
}
