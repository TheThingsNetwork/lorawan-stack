// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// Package profilefetcher contains definitions of end-device's profile fetchers based on the VersionIDs or VendorIDs.
package profilefetcher

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// TemplateFetcher implements the steps required to obtain a end-device template from the Device Repository.
type TemplateFetcher interface {
	GetTemplate(ctx context.Context, in *ttnpb.GetTemplateRequest) (*ttnpb.EndDeviceTemplate, error)
}

// Component abstracts the underlying *component.Component.
type Component interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

type templateFetcher struct {
	c Component
}

// NewTemplateFetcher returns an end-device template fetcher with predefined call options.
func NewTemplateFetcher(c Component) TemplateFetcher {
	return &templateFetcher{
		c: c,
	}
}

// GetTemplate makes a request to the Device Repository server with its predefined call options.
func (tf *templateFetcher) GetTemplate(
	ctx context.Context,
	in *ttnpb.GetTemplateRequest,
) (*ttnpb.EndDeviceTemplate, error) {
	conn, err := tf.c.GetPeerConn(ctx, ttnpb.ClusterRole_DEVICE_REPOSITORY, nil)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get Device Repository peer")
		return nil, err
	}

	opt, err := rpcmetadata.WithForwardedAuth(ctx, tf.c.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	return ttnpb.NewDeviceRepositoryClient(conn).GetTemplate(ctx, in, opt)
}

type templateFetcherContextKeyType struct{}

var templateFetcherContextKey templateFetcherContextKeyType

// NewContextWithFetcher returns a context with the TemplateFetcher.
func NewContextWithFetcher(ctx context.Context, tf TemplateFetcher) context.Context {
	return context.WithValue(ctx, templateFetcherContextKey, tf)
}

// fetcherFromContext returns a TemplateFetcher from the context provided. Will return false if not present.
func fetcherFromContext(ctx context.Context) (TemplateFetcher, bool) {
	if fetcher, ok := ctx.Value(templateFetcherContextKey).(TemplateFetcher); ok {
		return fetcher, true
	}
	return nil, false
}
