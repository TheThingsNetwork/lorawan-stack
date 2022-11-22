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

// Package rights implements rights fetching and checking.
package rights

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Map stores rights for a given ID.
type Map struct {
	syncMap sync.Map
}

// NewMap returns a pointer to a new Map.
func NewMap(rights map[string]*ttnpb.Rights) *Map {
	m := &Map{}
	for k, v := range rights {
		m.SetRights(k, v)
	}
	return m
}

// SetRights sets the rights for the given UID.
func (m *Map) SetRights(uid string, rights *ttnpb.Rights) {
	m.syncMap.Store(uid, rights)
}

// GetRights returns the rights stored in the map for a given UID,
// or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) GetRights(uid string) (*ttnpb.Rights, bool) {
	if r, ok := m.syncMap.Load(uid); ok {
		return r.(*ttnpb.Rights), ok
	}
	return nil, false
}

// MissingRights returns the rights that are missing for the given RightsMap.
func (m *Map) MissingRights(uid string, rights ...ttnpb.Right) []ttnpb.Right {
	if r, ok := m.GetRights(uid); ok {
		return ttnpb.RightsFrom(rights...).Sub(r).GetRights()
	}
	return rights
}

// Rights for the request.
type Rights struct {
	ApplicationRights  Map
	ClientRights       Map
	GatewayRights      Map
	OrganizationRights Map
	UserRights         Map
}

// SetApplicationRights sets the rights for the given application.
func (r *Rights) setApplicationRights(appUID string, rights *ttnpb.Rights) {
	r.ApplicationRights.SetRights(appUID, rights)
}

// SetClientRights sets the rights for the given client.
func (r *Rights) setClientRights(cliUID string, rights *ttnpb.Rights) {
	r.ClientRights.SetRights(cliUID, rights)
}

// SetGatewayRights sets the rights for the given gateway.
func (r *Rights) setGatewayRights(gtwUID string, rights *ttnpb.Rights) {
	r.GatewayRights.SetRights(gtwUID, rights)
}

// SetOrganizationRights sets the rights for the given organization.
func (r *Rights) setOrganizationRights(orgUID string, rights *ttnpb.Rights) {
	r.OrganizationRights.SetRights(orgUID, rights)
}

// SetUserRights sets the rights for the given user.
func (r *Rights) setUserRights(usrUID string, rights *ttnpb.Rights) {
	r.UserRights.SetRights(usrUID, rights)
}

// MissingApplicationRights returns the rights that are missing for the given application.
func (r *Rights) MissingApplicationRights(appUID string, rights ...ttnpb.Right) []ttnpb.Right {
	return r.ApplicationRights.MissingRights(appUID, rights...)
}

// MissingClientRights returns the rights that are missing for the given client.
func (r *Rights) MissingClientRights(cliUID string, rights ...ttnpb.Right) []ttnpb.Right {
	return r.ClientRights.MissingRights(cliUID, rights...)
}

// MissingGatewayRights returns the rights that are missing for the given gateway.
func (r *Rights) MissingGatewayRights(gtwUID string, rights ...ttnpb.Right) []ttnpb.Right {
	return r.GatewayRights.MissingRights(gtwUID, rights...)
}

// MissingOrganizationRights returns the rights that are missing for the given organization.
func (r *Rights) MissingOrganizationRights(orgUID string, rights ...ttnpb.Right) []ttnpb.Right {
	return r.OrganizationRights.MissingRights(orgUID, rights...)
}

// MissingUserRights returns the rights that are missing for the given user.
func (r *Rights) MissingUserRights(usrUID string, rights ...ttnpb.Right) []ttnpb.Right {
	return r.UserRights.MissingRights(usrUID, rights...)
}

// IncludesApplicationRights returns whether the given rights are included for the given application.
func (r *Rights) IncludesApplicationRights(appUID string, rights ...ttnpb.Right) bool {
	return len(r.MissingApplicationRights(appUID, rights...)) == 0
}

// IncludesClientRights returns whether the given rights are included for the given client.
func (r *Rights) IncludesClientRights(cliUID string, rights ...ttnpb.Right) bool {
	return len(r.MissingClientRights(cliUID, rights...)) == 0
}

// IncludesGatewayRights returns whether the given rights are included for the given gateway.
func (r *Rights) IncludesGatewayRights(gtwUID string, rights ...ttnpb.Right) bool {
	return len(r.MissingGatewayRights(gtwUID, rights...)) == 0
}

// IncludesOrganizationRights returns whether the given rights are included for the given organization.
func (r *Rights) IncludesOrganizationRights(orgUID string, rights ...ttnpb.Right) bool {
	return len(r.MissingOrganizationRights(orgUID, rights...)) == 0
}

// IncludesUserRights returns whether the given rights are included for the given user.
func (r *Rights) IncludesUserRights(usrUID string, rights ...ttnpb.Right) bool {
	return len(r.MissingUserRights(usrUID, rights...)) == 0
}

type rightsKeyType struct{}

var rightsKey rightsKeyType

func fromContext(ctx context.Context) (*Rights, bool) {
	if rights, ok := ctx.Value(rightsKey).(*Rights); ok {
		return rights, true
	}
	return &Rights{}, false
}

// NewContext returns a derived context with the given rights.
func NewContext(ctx context.Context, rights *Rights) context.Context {
	return context.WithValue(ctx, rightsKey, rights)
}

type rightsCacheKeyType struct{}

var rightsCacheKey rightsCacheKeyType

// NewContextWithCache returns a derived context with a rights cache.
// This should only be used for request contexts.
func NewContextWithCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, rightsCacheKey, &Rights{})
}

func cacheInContext(ctx context.Context, f func(*Rights)) {
	if rights, ok := ctx.Value(rightsCacheKey).(*Rights); ok {
		f(rights)
	}
}

func cacheFromContext(ctx context.Context) (*Rights, bool) {
	if rights, ok := ctx.Value(rightsCacheKey).(*Rights); ok {
		return rights, true
	}
	return &Rights{}, false
}

type authInfoKeyType struct{}

var authInfoKey authInfoKeyType

func authInfoFromContext(ctx context.Context) (*ttnpb.AuthInfoResponse, bool) {
	if authInfo, ok := ctx.Value(authInfoKey).(*ttnpb.AuthInfoResponse); ok {
		return authInfo, true
	}
	return nil, false
}

// NewContextWithAuthInfo returns a derived context with the authInfo.
func NewContextWithAuthInfo(ctx context.Context, authInfo *ttnpb.AuthInfoResponse) context.Context {
	return context.WithValue(ctx, authInfoKey, authInfo)
}

type authInfoCacheKeyType struct{}

var authInfoCacheKey authInfoCacheKeyType

// NewContextWithAuthInfoCache returns a derived context with an authentication info cache.
// This should only be used for request contexts.
func NewContextWithAuthInfoCache(ctx context.Context) context.Context {
	r := &ttnpb.AuthInfoResponse{}
	return context.WithValue(ctx, authInfoCacheKey, &r)
}

func cacheAuthInfoInContext(ctx context.Context, res *ttnpb.AuthInfoResponse) {
	if authInfo, ok := ctx.Value(authInfoCacheKey).(**ttnpb.AuthInfoResponse); ok {
		*authInfo = ttnpb.Clone(res)
	}
}

func cacheAuthInfoFromContext(ctx context.Context) (*ttnpb.AuthInfoResponse, bool) {
	if authInfo, ok := ctx.Value(authInfoCacheKey).(**ttnpb.AuthInfoResponse); ok {
		return *authInfo, true
	}
	return nil, false
}
