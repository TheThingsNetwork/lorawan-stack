// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package redis

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type mockAssociationRegistry struct {
	ClearAssociationsFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) error
	GetAssociationFunc    func(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageAssociation, error)                                                                                                      // nolint: lll
	ListAssociationsFunc  func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageAssociation, error)                                                                                                                        // nolint: lll
	SetAssociationFunc    func(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error)) (*ttnpb.ApplicationPackageAssociation, error) // nolint: lll
	WithPaginationFunc    func(ctx context.Context, limit, page uint32, total *int64) context.Context
}

// ClearAssociations implements packages.AssociationRegistry.
func (reg *mockAssociationRegistry) ClearAssociations(
	ctx context.Context,
	ids *ttnpb.EndDeviceIdentifiers,
) error {
	if reg.ClearAssociationsFunc == nil {
		panic("ClearAssociations called, but not set")
	}

	return reg.ClearAssociationsFunc(ctx, ids)
}

// GetAssociation implements packages.AssociationRegistry.
func (reg *mockAssociationRegistry) GetAssociation(
	ctx context.Context,
	ids *ttnpb.ApplicationPackageAssociationIdentifiers,
	paths []string,
) (*ttnpb.ApplicationPackageAssociation, error) {
	if reg.GetAssociationFunc == nil {
		panic("GetAssociation called, but not set")
	}

	return reg.GetAssociationFunc(ctx, ids, paths)
}

// ListAssociations implements packages.AssociationRegistry.
func (reg *mockAssociationRegistry) ListAssociations(
	ctx context.Context,
	ids *ttnpb.EndDeviceIdentifiers,
	paths []string,
) ([]*ttnpb.ApplicationPackageAssociation, error) {
	if reg.ListAssociationsFunc == nil {
		panic("ListAssociations called, but not set")
	}

	return reg.ListAssociationsFunc(ctx, ids, paths)
}

// SetAssociation implements packages.AssociationRegistry.
func (reg *mockAssociationRegistry) SetAssociation(
	ctx context.Context,
	ids *ttnpb.ApplicationPackageAssociationIdentifiers,
	gets []string,
	f func(*ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error),
) (*ttnpb.ApplicationPackageAssociation, error) {
	if reg.SetAssociationFunc == nil {
		panic("SetAssociation called, but not set")
	}

	return reg.SetAssociationFunc(ctx, ids, gets, f)
}

// WithPagination implements packages.AssociationRegistry.
func (reg *mockAssociationRegistry) WithPagination(
	ctx context.Context,
	limit uint32,
	page uint32,
	total *int64,
) context.Context {
	if reg.WithPaginationFunc == nil {
		panic("WithPagination called, but not set")
	}

	return reg.WithPaginationFunc(ctx, limit, page, total)
}

// NewAssociationRegistryMock returns a new mock AssociationRegistry.
func NewAssociationRegistryMock(
	clearAssociationsFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) error,
	getAssociationFunc func(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageAssociation, error), // nolint: lll
	listAssociationsFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageAssociation, error), // nolint: lll
	setAssociationFunc func(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error)) (*ttnpb.ApplicationPackageAssociation, error), // nolint: lll
	withPaginationFunc func(ctx context.Context, limit, page uint32, total *int64) context.Context,
) packages.AssociationRegistry {
	return &mockAssociationRegistry{
		ClearAssociationsFunc: clearAssociationsFunc,
		GetAssociationFunc:    getAssociationFunc,
		ListAssociationsFunc:  listAssociationsFunc,
		SetAssociationFunc:    setAssociationFunc,
		WithPaginationFunc:    withPaginationFunc,
	}
}

type mockDefaultAssociationRegistry struct {
	ClearDefaultAssociationsFunc func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) error
	GetDefaultAssociationFunc    func(ctx context.Context, ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageDefaultAssociation, error)                                                                                                                    // nolint: lll
	ListDefaultAssociationsFunc  func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageDefaultAssociation, error)                                                                                                                                           // nolint: lll
	SetDefaultAssociationFunc    func(ctx context.Context, ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageDefaultAssociation) (*ttnpb.ApplicationPackageDefaultAssociation, []string, error)) (*ttnpb.ApplicationPackageDefaultAssociation, error) // nolint: lll
	WithPaginationFunc           func(ctx context.Context, limit, page uint32, total *int64) context.Context                                                                                                                                                                                                   // nolint: lll
}

// ClearDefaultAssociations implements packages.DefaultAssociationRegistry.
func (reg *mockDefaultAssociationRegistry) ClearDefaultAssociations(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
) error {
	if reg.ClearDefaultAssociationsFunc == nil {
		panic("ClearDefaultAssociations called, but not set")
	}

	return reg.ClearDefaultAssociationsFunc(ctx, ids)
}

// GetDefaultAssociation implements packages.DefaultAssociationRegistry.
func (reg *mockDefaultAssociationRegistry) GetDefaultAssociation(
	ctx context.Context,
	ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers,
	paths []string,
) (*ttnpb.ApplicationPackageDefaultAssociation, error) {
	if reg.GetDefaultAssociationFunc == nil {
		panic("GetDefaultAssociation called, but not set")
	}

	return reg.GetDefaultAssociationFunc(ctx, ids, paths)
}

// ListDefaultAssociations implements packages.DefaultAssociationRegistry.
func (reg *mockDefaultAssociationRegistry) ListDefaultAssociations(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
	paths []string,
) ([]*ttnpb.ApplicationPackageDefaultAssociation, error) {
	if reg.ListDefaultAssociationsFunc == nil {
		panic("ListDefaultAssociations called, but not set")
	}

	return reg.ListDefaultAssociationsFunc(ctx, ids, paths)
}

// SetDefaultAssociation implements packages.DefaultAssociationRegistry.
func (reg *mockDefaultAssociationRegistry) SetDefaultAssociation(
	ctx context.Context,
	ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers,
	gets []string,
	f func(*ttnpb.ApplicationPackageDefaultAssociation) (*ttnpb.ApplicationPackageDefaultAssociation, []string, error), // nolint: lll
) (*ttnpb.ApplicationPackageDefaultAssociation, error) {
	if reg.SetDefaultAssociationFunc == nil {
		panic("SetDefaultAssociation called, but not set")
	}

	return reg.SetDefaultAssociationFunc(ctx, ids, gets, f)
}

// WithPagination implements packages.DefaultAssociationRegistry.
func (reg *mockDefaultAssociationRegistry) WithPagination(
	ctx context.Context,
	limit uint32,
	page uint32,
	total *int64,
) context.Context {
	if reg.WithPaginationFunc == nil {
		panic("WithPagination called, but not set")
	}

	return reg.WithPaginationFunc(ctx, limit, page, total)
}

// NewDefaultAssociationRegistryMock returns a new mock for the default association registry.
func NewDefaultAssociationRegistryMock(
	clearDefaultAssociationsFunc func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) error,
	getDefaultAssociationFunc func(ctx context.Context, ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageDefaultAssociation, error), // nolint: lll
	listDefaultAssociationsFunc func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageDefaultAssociation, error), // nolint: lll
	setDefaultAssociationFunc func(ctx context.Context, ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageDefaultAssociation) (*ttnpb.ApplicationPackageDefaultAssociation, []string, error)) (*ttnpb.ApplicationPackageDefaultAssociation, error), // nolint: lll
	withPaginationFunc func(ctx context.Context, limit, page uint32, total *int64) context.Context,
) packages.DefaultAssociationRegistry {
	return &mockDefaultAssociationRegistry{
		ClearDefaultAssociationsFunc: clearDefaultAssociationsFunc,
		GetDefaultAssociationFunc:    getDefaultAssociationFunc,
		ListDefaultAssociationsFunc:  listDefaultAssociationsFunc,
		SetDefaultAssociationFunc:    setDefaultAssociationFunc,
		WithPaginationFunc:           withPaginationFunc,
	}
}

type mockTransactionRegistry struct {
	EndDeviceTransactionFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, fPort uint32, packageName string, fn func(ctx context.Context) error) error // nolint: lll
}

// EndDeviceTransaction implements packages.TransactionRegistry.
func (reg *mockTransactionRegistry) EndDeviceTransaction(
	ctx context.Context,
	ids *ttnpb.EndDeviceIdentifiers,
	fPort uint32,
	packageName string,
	fn func(ctx context.Context) error,
) error {
	if reg.EndDeviceTransactionFunc == nil {
		panic("EndDeviceTransaction called, but not set")
	}

	return reg.EndDeviceTransactionFunc(ctx, ids, fPort, packageName, fn)
}

// NewTransactionRegistryMock returns a new mock TransactionRegistry.
func NewTransactionRegistryMock(
	endDeviceTransactionFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, fPort uint32, packageName string, fn func(ctx context.Context) error) error, // nolint: lll
) packages.TransactionRegistry {
	return &mockTransactionRegistry{
		EndDeviceTransactionFunc: endDeviceTransactionFunc,
	}
}

type mockApplicationPackagesRegistry struct {
	packages.AssociationRegistry
	packages.DefaultAssociationRegistry
	packages.TransactionRegistry

	RangeFunc          func(ctx context.Context, paths []string, devFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.ApplicationPackageAssociation) bool, appFunc func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationPackageDefaultAssociation) bool) error // nolint: lll
	WithPaginationFunc func(ctx context.Context, limit uint32, page uint32, total *int64) context.Context
}

// Range implements packages.Registry.
func (reg *mockApplicationPackagesRegistry) Range(
	ctx context.Context,
	paths []string,
	devFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.ApplicationPackageAssociation) bool,
	appFunc func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationPackageDefaultAssociation) bool,
) error {
	if reg.RangeFunc == nil {
		panic("Range is called, but not set")
	}

	return reg.RangeFunc(ctx, paths, devFunc, appFunc)
}

// WithPagination implements packages.Registry.
func (reg *mockApplicationPackagesRegistry) WithPagination(
	ctx context.Context,
	limit uint32,
	page uint32,
	total *int64,
) context.Context {
	if reg.WithPaginationFunc == nil {
		panic("WithPagination is called, but not set")
	}

	return reg.WithPaginationFunc(ctx, limit, page, total)
}

// NewAppPkgsRegistryWithMockedHandlers creates a new application packages registry with mocked handlers.
func NewAppPkgsRegistryWithMockedHandlers(
	associationRegistry packages.AssociationRegistry,
	defaultAssociationRegistry packages.DefaultAssociationRegistry,
	transactionRegistry packages.TransactionRegistry,

	rangeFunc func(ctx context.Context, paths []string, devFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.ApplicationPackageAssociation) bool, appFunc func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationPackageDefaultAssociation) bool) error, // nolint: lll
	withPaginationFunc func(ctx context.Context, limit uint32, page uint32, total *int64) context.Context,
) packages.Registry {
	return &mockApplicationPackagesRegistry{
		AssociationRegistry:        associationRegistry,
		DefaultAssociationRegistry: defaultAssociationRegistry,
		TransactionRegistry:        transactionRegistry,
		RangeFunc:                  rangeFunc,
		WithPaginationFunc:         withPaginationFunc,
	}
}
