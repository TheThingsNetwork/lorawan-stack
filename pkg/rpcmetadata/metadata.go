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

// Package rpcmetadata contains utilities for transporting common request metadata over gRPC.
package rpcmetadata

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

// MD contains the TTN metadata fields
type MD struct {
	ID             string
	AuthType       string
	AuthValue      string
	ServiceType    string
	ServiceVersion string
	NetAddress     string
	Limit          uint64
	Offset         uint64
}

// GetRequestMetadata returns the request metadata with per-rpc credentials
func (m MD) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if m.AuthType == "" || m.AuthValue == "" {
		return nil, nil
	}
	return map[string]string{
		"authorization": m.AuthType + " " + m.AuthValue,
	}, nil
}

// RequireTransportSecurity returns true if authentication is configured
func (m MD) RequireTransportSecurity() bool {
	if m.AuthType == "" || m.AuthValue == "" {
		return false
	}
	return true
}

// ToMetadata puts the TTN metadata fields in a metadata.MD
func (m MD) ToMetadata() metadata.MD {
	var pairs []string
	if m.ID != "" {
		pairs = append(pairs, "id", m.ID)
	}
	if m.ServiceType != "" {
		pairs = append(pairs, "service-type", m.ServiceType)
	}
	if m.ServiceVersion != "" {
		pairs = append(pairs, "service-version", m.ServiceVersion)
	}
	if m.NetAddress != "" {
		pairs = append(pairs, "net-address", m.NetAddress)
	}
	if m.Limit != 0 {
		pairs = append(pairs, "limit", strconv.FormatUint(m.Limit, 10))
	}
	if m.Offset != 0 {
		pairs = append(pairs, "offset", strconv.FormatUint(m.Offset, 10))
	}
	return metadata.Pairs(pairs...)
}

// ToOutgoingContext puts the TTN metadata fields in an outgoing context.Context
func (m MD) ToOutgoingContext(ctx context.Context) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = metadata.Join(m.ToMetadata(), md)
	return metadata.NewOutgoingContext(ctx, md)
}

// ToIncomingContext puts the TTN metadata fields in an incoming context.Context
func (m MD) ToIncomingContext(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	md = metadata.Join(m.ToMetadata(), md)
	return metadata.NewIncomingContext(ctx, md)
}

// FromMetadata returns the TTN metadata from metadata.MD
func FromMetadata(md metadata.MD) (m MD) {
	if id, ok := md["id"]; ok && len(id) > 0 {
		m.ID = id[0]
	}
	if authorization, ok := md["authorization"]; ok && len(authorization) > 0 {
		if parts := strings.SplitN(authorization[0], " ", 2); len(parts) == 2 {
			m.AuthType, m.AuthValue = parts[0], parts[1]
		}
	}
	if serviceType, ok := md["service-type"]; ok && len(serviceType) > 0 {
		m.ServiceType = serviceType[0]
	}
	if serviceVersion, ok := md["service-version"]; ok && len(serviceVersion) > 0 {
		m.ServiceVersion = serviceVersion[0]
	}
	if netAddress, ok := md["net-address"]; ok && len(netAddress) > 0 {
		m.NetAddress = netAddress[0]
	}
	if limit, ok := md["limit"]; ok && len(limit) > 0 {
		m.Limit, _ = strconv.ParseUint(limit[0], 10, 64)
	}
	if offset, ok := md["offset"]; ok && len(offset) > 0 {
		m.Offset, _ = strconv.ParseUint(offset[0], 10, 64)
	}
	return
}

// FromOutgoingContext returns the TTN metadata from the outgoing context.Context
func FromOutgoingContext(ctx context.Context) (m MD) {
	md, _ := metadata.FromOutgoingContext(ctx)
	return FromMetadata(md)
}

// FromIncomingContext returns the TTN metadata from the incoming context.Context
func FromIncomingContext(ctx context.Context) (m MD) {
	md, _ := metadata.FromIncomingContext(ctx)
	return FromMetadata(md)
}
