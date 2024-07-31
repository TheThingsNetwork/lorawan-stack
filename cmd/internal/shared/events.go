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

package shared

import (
	"context"
	"crypto/tls"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/events/cloud"
	"go.thethings.network/lorawan-stack/v3/pkg/events/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/events/redis"
	managedclient "go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver/managed/client"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	_ "gocloud.dev/pubsub/awssnssqs" // AWS backend for PubSub.
	_ "gocloud.dev/pubsub/gcppubsub" // GCP backend for PubSub.
	"google.golang.org/grpc"
)

// Component contains a minimal component.Component definition.
type Component interface {
	workerpool.Component
	GetTLSClientConfig(context.Context, ...tlsconfig.Option) (*tls.Config, error)
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

// InitializeEvents initializes the event system.
func InitializeEvents(ctx context.Context, component Component, conf config.ServiceBase) error {
	var ps events.PubSub
	switch conf.Events.Backend {
	case "internal":
		ps = basic.NewPubSub()
	case "redis":
		ps = redis.NewPubSub(ctx, component, conf.Events.Redis, conf.Events.Batch)
	case "cloud":
		cloudPS, err := cloud.NewPubSub(ctx, component, conf.Events.Cloud.PublishURL, conf.Events.Cloud.SubscribeURL)
		if err != nil {
			return err
		}
		ps = cloudPS
	default:
		return fmt.Errorf("unknown events backend: %s", conf.Events.Backend)
	}

	var extraStreams []mux.Option
	if conf.TTGC.Enabled {
		matcher, err := mux.MatchPatterns(managedclient.EventNamePattern)
		if err != nil {
			return err
		}
		extraStreams = append(extraStreams, mux.WithStream(managedclient.NewEvents(component), matcher))
	}

	if len(extraStreams) > 0 {
		ps = mux.New(component, ps, extraStreams...)
	}
	events.SetDefaultPubSub(ps)
	return nil
}
