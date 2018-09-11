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

package rights

import (
	"context"
	"sync"
	"time"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var now = time.Now // override this in unit tests

func newReq(ctx context.Context, id ttnpb.Identifiers) cachedReq {
	md := rpcmetadata.FromIncomingContext(ctx)
	return cachedReq{UniqueID: unique.ID(ctx, id), AuthType: md.AuthType, AuthValue: md.AuthValue}
}

type cachedReq struct {
	UniqueID  string
	AuthType  string
	AuthValue string
}

func newRes() *cachedRes {
	return &cachedRes{waitChan: make(chan struct{})}
}

type cachedRes struct {
	waitChan chan struct{}
	time     time.Time
	rights   *ttnpb.Rights
	err      error
}

func (c *cachedRes) valid(successTTL, errorTTL time.Duration) bool {
	if c == nil {
		return false
	}
	select {
	case <-c.waitChan:
	default:
		return true // still fetching...
	}
	now := now()
	if c.err != nil {
		return now.Sub(c.time) <= errorTTL
	}
	return now.Sub(c.time) <= successTTL
}

func (c *cachedRes) set(rights *ttnpb.Rights, err error) {
	c.rights, c.err = rights, err
	if err == nil || (!errors.IsCanceled(err) && !errors.IsDeadlineExceeded(err)) {
		c.time = now() // we only have a real result if the request wasn't canceled or timed out.
	}
	close(c.waitChan)
}

func (c *cachedRes) wait() {
	<-c.waitChan
}

// NewInMemoryCache returns a new in-memory cache on top of the given fetcher.
// Successful responses are valid for the duration of successTTL, unsuccessful
// responses are valid for the duration of errorTTL.
func NewInMemoryCache(fetcher Fetcher, successTTL, errorTTL time.Duration) Fetcher {
	return &inMemoryCache{
		Fetcher:            fetcher,
		successTTL:         successTTL,
		errorTTL:           errorTTL,
		lastCleanup:        now(),
		applicationRights:  make(map[cachedReq]*cachedRes),
		clientRights:       make(map[cachedReq]*cachedRes),
		gatewayRights:      make(map[cachedReq]*cachedRes),
		organizationRights: make(map[cachedReq]*cachedRes),
		userRights:         make(map[cachedReq]*cachedRes),
	}
}

type inMemoryCache struct {
	Fetcher
	successTTL         time.Duration
	errorTTL           time.Duration
	mu                 sync.Mutex
	lastCleanup        time.Time
	applicationRights  map[cachedReq]*cachedRes
	clientRights       map[cachedReq]*cachedRes
	gatewayRights      map[cachedReq]*cachedRes
	organizationRights map[cachedReq]*cachedRes
	userRights         map[cachedReq]*cachedRes
}

// maybeCleanup cleans up expired results if necessary.
func (f *inMemoryCache) maybeCleanup() {
	cleanupTTL := f.successTTL
	if f.errorTTL > cleanupTTL {
		cleanupTTL = f.errorTTL
	}
	if now().Sub(f.lastCleanup) <= cleanupTTL*10 {
		return
	}
	for req, res := range f.applicationRights {
		if !res.valid(f.successTTL, f.errorTTL) {
			delete(f.applicationRights, req)
		}
	}
	for req, res := range f.clientRights {
		if !res.valid(f.successTTL, f.errorTTL) {
			delete(f.clientRights, req)
		}
	}
	for req, res := range f.gatewayRights {
		if !res.valid(f.successTTL, f.errorTTL) {
			delete(f.gatewayRights, req)
		}
	}
	for req, res := range f.organizationRights {
		if !res.valid(f.successTTL, f.errorTTL) {
			delete(f.organizationRights, req)
		}
	}
	for req, res := range f.userRights {
		if !res.valid(f.successTTL, f.errorTTL) {
			delete(f.userRights, req)
		}
	}
	f.lastCleanup = now()
}

func (f *inMemoryCache) ApplicationRights(ctx context.Context, appID ttnpb.ApplicationIdentifiers) (rights *ttnpb.Rights, err error) {
	defer func() { registerRightsRequest(ctx, "application", rights, err) }()
	req := newReq(ctx, appID)
	f.mu.Lock()
	res := f.applicationRights[req]
	if !res.valid(f.successTTL, f.errorTTL) {
		res = newRes()
		f.applicationRights[req] = res
		go res.set(f.Fetcher.ApplicationRights(ctx, appID))
	}
	f.maybeCleanup()
	f.mu.Unlock()
	res.wait()
	return res.rights, res.err
}

func (f *inMemoryCache) ClientRights(ctx context.Context, clientID ttnpb.ClientIdentifiers) (rights *ttnpb.Rights, err error) {
	defer func() { registerRightsRequest(ctx, "client", rights, err) }()
	req := newReq(ctx, clientID)
	f.mu.Lock()
	res := f.clientRights[req]
	if !res.valid(f.successTTL, f.errorTTL) {
		res = newRes()
		f.clientRights[req] = res
		go res.set(f.Fetcher.ClientRights(ctx, clientID))
	}
	f.maybeCleanup()
	f.mu.Unlock()
	res.wait()
	return res.rights, res.err
}

func (f *inMemoryCache) GatewayRights(ctx context.Context, gtwID ttnpb.GatewayIdentifiers) (rights *ttnpb.Rights, err error) {
	defer func() { registerRightsRequest(ctx, "gateway", rights, err) }()
	req := newReq(ctx, gtwID)
	f.mu.Lock()
	res := f.gatewayRights[req]
	if !res.valid(f.successTTL, f.errorTTL) {
		res = newRes()
		f.gatewayRights[req] = res
		go res.set(f.Fetcher.GatewayRights(ctx, gtwID))
	}
	f.maybeCleanup()
	f.mu.Unlock()
	res.wait()
	return res.rights, res.err
}

func (f *inMemoryCache) OrganizationRights(ctx context.Context, orgID ttnpb.OrganizationIdentifiers) (rights *ttnpb.Rights, err error) {
	defer func() { registerRightsRequest(ctx, "organization", rights, err) }()
	req := newReq(ctx, orgID)
	f.mu.Lock()
	res := f.organizationRights[req]
	if !res.valid(f.successTTL, f.errorTTL) {
		res = newRes()
		f.organizationRights[req] = res
		go res.set(f.Fetcher.OrganizationRights(ctx, orgID))
	}
	f.maybeCleanup()
	f.mu.Unlock()
	res.wait()
	return res.rights, res.err
}

func (f *inMemoryCache) UserRights(ctx context.Context, userID ttnpb.UserIdentifiers) (rights *ttnpb.Rights, err error) {
	defer func() { registerRightsRequest(ctx, "user", rights, err) }()
	req := newReq(ctx, userID)
	f.mu.Lock()
	res := f.userRights[req]
	if !res.valid(f.successTTL, f.errorTTL) {
		res = newRes()
		f.userRights[req] = res
		go res.set(f.Fetcher.UserRights(ctx, userID))
	}
	f.maybeCleanup()
	f.mu.Unlock()
	res.wait()
	return res.rights, res.err
}
