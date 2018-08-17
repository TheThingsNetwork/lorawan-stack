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

	"github.com/prometheus/client_golang/prometheus"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const subsystem = "rights_hook"

var rightsRequests = metrics.NewContextualCounterVec(
	prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "Rights Requests",
	},
	[]string{"type", "result"},
)

var rightsFetches = metrics.NewContextualCounterVec(
	prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "fetches_total",
		Help:      "Rights Fetches from Identity Server",
	},
	[]string{"type", "result"},
)

func init() {
	metrics.MustRegister(rightsRequests, rightsFetches)
}

func register(ctx context.Context, c *metrics.ContextualCounterVec, entity string, rights []ttnpb.Right, err error) {
	switch {
	case errors.IsUnauthenticated(err):
		c.WithLabelValues(ctx, entity, "unauthenticated").Inc()
	case errors.IsPermissionDenied(err):
		c.WithLabelValues(ctx, entity, "permission_denied").Inc()
	case err != nil:
		c.WithLabelValues(ctx, entity, "error").Inc()
	case len(rights) == 0:
		c.WithLabelValues(ctx, entity, "zero").Inc()
	default:
		c.WithLabelValues(ctx, entity, "ok").Inc()
	}
}

func registerRightsRequest(ctx context.Context, entity string, rights []ttnpb.Right, err error) {
	register(ctx, rightsRequests, entity, rights, err)
}

func registerRightsFetch(ctx context.Context, entity string, rights []ttnpb.Right, err error) {
	register(ctx, rightsFetches, entity, rights, err)
}
