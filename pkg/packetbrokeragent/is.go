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

package packetbrokeragent

import (
	"context"
	"errors"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
)

var (
	cacheSize     = 2048
	cacheTTL      = 5 * time.Minute
	cacheErrorTTL = time.Minute
)

// Cluster is an interface which selects for methods from the cluster component will be used in the pba.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
}

// cacheItem stores the payload processors as well as the error response.
type cacheItem[Entity ttnpb.IDStringer] struct {
	entity Entity
	err    error
}

// IS defines wrappers that perform grpc calls to the Identity Server.
type IS struct {
	Cluster
	cache gcache.Cache
}

// newIS returns a new IS, which is a set of wrappers that perform calls to the Identity Server.
func newIS(c Cluster) *IS {
	return &IS{
		Cluster: c,
		cache:   gcache.New(cacheSize).ARC().Expiration(cacheTTL).Build(),
	}
}

// GetUser calls the Entity Registry with cluster auth and returns an user.
func (is IS) GetUser(ctx context.Context, req *ttnpb.GetUserRequest) (*ttnpb.User, error) {
	v, err := is.cache.Get(unique.ID(ctx, req))
	if err != nil && !errors.Is(err, gcache.KeyNotFoundError) {
		return nil, err
	}
	if v != nil {
		cacheItem := v.(*cacheItem[*ttnpb.User])
		return cacheItem.entity, cacheItem.err
	}
	registry, err := is.newUserRegistryClient(ctx)
	if err != nil {
		return nil, err
	}

	expire := cacheTTL
	usr, err := registry.Get(ctx, req, is.WithClusterAuth())
	if err != nil {
		expire = cacheErrorTTL
	}

	// Caches error in order to avoid calls in succession, time to live is significantly lower when an error occurs.
	item := &cacheItem[*ttnpb.User]{entity: usr, err: err}
	if err := is.cache.SetWithExpire(unique.ID(ctx, item.entity), item, expire); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to cache user")
	}
	return item.entity, item.err
}

// GetOrganization calls the Entity Registry with cluster auth and returns an organization.
func (is IS) GetOrganization(ctx context.Context, req *ttnpb.GetOrganizationRequest) (*ttnpb.Organization, error) {
	v, err := is.cache.Get(unique.ID(ctx, req))
	if err != nil && !errors.Is(err, gcache.KeyNotFoundError) {
		return nil, err
	}
	if v != nil {
		cacheItem := v.(*cacheItem[*ttnpb.Organization])
		return cacheItem.entity, cacheItem.err
	}
	registry, err := is.newOrganizationRegistryClient(ctx)
	if err != nil {
		return nil, err
	}

	expire := cacheTTL
	org, err := registry.Get(ctx, req, is.WithClusterAuth())
	if err != nil {
		expire = cacheErrorTTL
	}

	// Caches error in order to avoid calls in succession, time to live is significantly lower when an error occurs.
	item := &cacheItem[*ttnpb.Organization]{entity: org, err: err}
	if err := is.cache.SetWithExpire(unique.ID(ctx, item.entity), item, expire); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to cache organization")
	}
	return item.entity, item.err
}

// newRegistryClient returns a new user registry client.
func (is IS) newUserRegistryClient(ctx context.Context) (ttnpb.UserRegistryClient, error) {
	cc, err := is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewUserRegistryClient(cc), nil
}

// newOrganizationRegistryClient returns a new organization registry client.
func (is IS) newOrganizationRegistryClient(ctx context.Context) (ttnpb.OrganizationRegistryClient, error) {
	cc, err := is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewOrganizationRegistryClient(cc), nil
}
