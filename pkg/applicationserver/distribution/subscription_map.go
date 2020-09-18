// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package distribution

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// newSubscriptionMap creates a mapping between application identifiers and subscription sets.
// The timeout represents the period after which a set will shut down if empty. If the timeout
// is zero, the sets never timeout.
func newSubscriptionMap(ctx context.Context, timeout time.Duration, setup func(*subscriptionSet, ttnpb.ApplicationIdentifiers) error) *subscriptionMap {
	return &subscriptionMap{
		ctx:     ctx,
		timeout: timeout,
		setup:   setup,
	}
}

type subscriptionMap struct {
	ctx     context.Context
	timeout time.Duration
	setup   func(*subscriptionSet, ttnpb.ApplicationIdentifiers) error
	sets    sync.Map
}

type subscriptionMapSet struct {
	set *subscriptionSet

	init    chan struct{}
	initErr error
}

var errSetNotFound = errors.DefineNotFound("set_not_found", "set not found")

// Load loads the subscription set associated with the application identifiers.
func (m *subscriptionMap) Load(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*subscriptionSet, error) {
	uid := unique.ID(ctx, ids)
	existing, ok := m.sets.Load(uid)
	if !ok {
		return nil, errSetNotFound.New()
	}
	exists := existing.(*subscriptionMapSet)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-exists.init:
	}
	if exists.initErr != nil {
		return nil, exists.initErr
	}
	return exists.set, nil
}

// LoadOrCreate loads the subscription set associated with the application identifiers.
// If the subscription set does not exist, it is created.
func (m *subscriptionMap) LoadOrCreate(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*subscriptionSet, error) {
	uid := unique.ID(ctx, ids)
	s := &subscriptionMapSet{
		init: make(chan struct{}),
	}
	if existing, loaded := m.sets.LoadOrStore(uid, s); loaded {
		exists := existing.(*subscriptionMapSet)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-exists.init:
		}
		if exists.initErr != nil {
			return nil, exists.initErr
		}
		return exists.set, nil
	}

	var err error
	defer func() {
		if err != nil {
			s.initErr = err
			m.sets.Delete(uid)
		}
		close(s.init)
	}()

	ctx = log.NewContextWithField(m.ctx, "application_uid", uid)
	ctx, err = unique.WithContext(ctx, uid)
	if err != nil {
		return nil, err
	}

	set := newSubscriptionSet(ctx, m.timeout)
	if err = m.setup(set, ids); err != nil {
		set.Cancel(err)
		return nil, err
	}
	go func() {
		<-set.Context().Done()
		m.sets.Delete(uid)
	}()
	s.set = set

	return set, nil
}

func noSetup(*subscriptionSet, ttnpb.ApplicationIdentifiers) error {
	return nil
}
