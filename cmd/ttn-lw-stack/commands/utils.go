// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package commands

import (
	"context"
	"crypto/tls"
	"strconv"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

const defaultPaginationLimit = 1000

// NewClusterComponentConnection connects returns a new cluster instance and a connection to a specified peer.
// The connection to a cluster peer is retried specified number of times before returning an error in case
// of connection not being ready.
func NewClusterComponentConnection(ctx context.Context,
	config *Config,
	delay time.Duration,
	maxRetries int,
	role ttnpb.ClusterRole,
) (*grpc.ClientConn, cluster.Cluster, error) {
	clusterOpts := []cluster.Option{}
	if config.Cluster.TLS {
		tlsConf := config.TLS
		tls := &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: tlsConf.InsecureSkipVerify,
		}
		if err := tlsConf.Client.ApplyTo(tls); err != nil {
			return nil, nil, err
		}
		clusterOpts = append(clusterOpts, cluster.WithTLSConfig(tls))
	}
	c, err := cluster.New(ctx, &config.Cluster, clusterOpts...)
	if err != nil {
		return nil, nil, err
	}
	if err := c.Join(); err != nil {
		return nil, nil, err
	}
	var cc *grpc.ClientConn
	for i := 0; i < maxRetries; i++ {
		time.Sleep(delay)
		cc, err = c.GetPeerConn(ctx, role, nil)
		if err == nil {
			return cc, c, nil
		}
	}
	return nil, nil, err
}

// FetchIdentityServerApplications returns the list of all non-expired applications in the Identity Server.
func FetchIdentityServerApplications(ctx context.Context, client ttnpb.ApplicationRegistryClient, clusterAuth grpc.CallOption, paginationDelay time.Duration) ([]*ttnpb.Application, error) {
	pageCounter := uint32(1)
	applicationList := make([]*ttnpb.Application, 0)
	for {
		res, err := client.List(ctx, &ttnpb.ListApplicationsRequest{
			Collaborator: nil,
			FieldMask:    ttnpb.FieldMask("ids"),
			Limit:        defaultPaginationLimit,
			Page:         pageCounter,
			Deleted:      true,
		}, clusterAuth)
		if err != nil {
			return nil, err
		}
		applicationList = append(applicationList, res.Applications...)
		if len(res.Applications) == 0 {
			break
		}
		pageCounter++
		time.Sleep(paginationDelay)
	}
	return applicationList, nil
}

// FetchIdentityServerEndDevices returns the list of all devices in the Identity Server.
func FetchIdentityServerEndDevices(ctx context.Context, client ttnpb.EndDeviceRegistryClient, clusterAuth grpc.CallOption, paginationDelay time.Duration) ([]*ttnpb.EndDevice, error) {
	pageCounter := uint32(1)
	deviceList := make([]*ttnpb.EndDevice, 0)
	for {
		res, err := client.List(ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: nil,
			FieldMask:      ttnpb.FieldMask("ids"),
			Limit:          defaultPaginationLimit,
			Page:           pageCounter,
		}, clusterAuth)
		if err != nil {
			return nil, err
		}
		deviceList = append(deviceList, res.EndDevices...)
		if len(res.EndDevices) == 0 {
			break
		}
		pageCounter++
		time.Sleep(paginationDelay)
	}
	return deviceList, nil
}

func setToArray(set map[string]struct{}) []string {
	keys := make([]string, len(set))
	i := 0
	for k := range set {
		keys[i] = k
		i++
	}
	return keys
}

func schemaVersionKey(cl *ttnredis.Client) string {
	return cl.Key("schema-version")
}

func recordSchemaVersion(cl *ttnredis.Client, version int) error {
	logger.WithField("version", version).Info("Setting schema version")
	return cl.Set(ctx, schemaVersionKey(cl), version, 0).Err()
}

func getSchemaVersion(cl *ttnredis.Client) (int, error) {
	schemaVersionString, err := cl.Get(ctx, schemaVersionKey(cl)).Result()
	if err != nil {
		if errors.IsNotFound(ttnredis.ConvertError(err)) {
			return 0, nil
		}
		return 0, err
	}
	schemaVersion, err := strconv.ParseInt(schemaVersionString, 10, 32)
	if err != nil {
		return 0, err
	}
	logger.WithField("version", schemaVersion).Info("Existing database schema version")
	return int(schemaVersion), nil
}
