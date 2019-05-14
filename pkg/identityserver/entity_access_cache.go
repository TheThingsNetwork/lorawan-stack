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

package identityserver

import (
	"context"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

func (is *IdentityServer) cachedMembershipsForAccount(ctx context.Context, ouIDs *ttnpb.OrganizationOrUserIdentifiers) map[ttnpb.Identifiers]*ttnpb.Rights {
	if is.redis == nil || is.config.AuthCache.MembershipTTL == 0 {
		return nil
	}
	hash, err := is.redis.HGetAll(is.redis.Key("memberships", unique.ID(ctx, ouIDs.Identifiers()))).Result()
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		is.Logger().WithError(err).Error("Failed to get cached membership info")
		return nil
	}
	if len(hash) == 0 {
		return nil
	}
	entityRights := make(map[ttnpb.Identifiers]*ttnpb.Rights, len(hash))
	for k, v := range hash {
		var entity ttnpb.EntityIdentifiers
		var rights ttnpb.Rights
		if err = ttnredis.UnmarshalProto(k, &entity); err != nil {
			return nil
		}
		if err = ttnredis.UnmarshalProto(v, &rights); err != nil {
			return nil
		}
		entityRights[entity.Identifiers()] = &rights
	}
	return entityRights
}

func (is *IdentityServer) invalidateCachedMembershipsForAccount(ctx context.Context, ouIDs *ttnpb.OrganizationOrUserIdentifiers) {
	if is.redis == nil || is.config.AuthCache.MembershipTTL == 0 {
		return
	}
	if err := is.redis.Del(is.redis.Key("memberships", unique.ID(ctx, ouIDs.Identifiers()))).Err(); err != nil {
		is.Logger().WithError(err).Error("Failed to delete cached membership info")
		return
	}
}

func (is *IdentityServer) cacheMembershipsForAccount(ctx context.Context, ouIDs *ttnpb.OrganizationOrUserIdentifiers, entityRights map[ttnpb.Identifiers]*ttnpb.Rights) {
	if is.redis == nil || is.config.AuthCache.MembershipTTL == 0 {
		return
	}
	if len(entityRights) == 0 {
		return
	}
	hash := make(map[string]interface{}, len(entityRights))
	for entity, rights := range entityRights {
		k, err := ttnredis.MarshalProto(entity.EntityIdentifiers())
		if err != nil {
			return
		}
		v, err := ttnredis.MarshalProto(rights)
		if err != nil {
			return
		}
		hash[k] = v
	}
	key := is.redis.Key("memberships", unique.ID(ctx, ouIDs.Identifiers()))
	_, err := is.redis.Pipelined(func(pipe redis.Pipeliner) (err error) {
		if err = pipe.Del(key).Err(); err != nil {
			return err
		}
		if err = pipe.HMSet(key, hash).Err(); err != nil {
			return err
		}
		if err = pipe.Expire(key, is.config.AuthCache.MembershipTTL).Err(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		is.Logger().WithError(err).Error("Failed to set cached membership info")
		return
	}
}
