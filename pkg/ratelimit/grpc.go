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

package ratelimit

import (
	"context"
	"fmt"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
)

func grpcEntityFromRequest(ctx context.Context, req any) string {
	if r, ok := req.(RateLimitKeyer); ok {
		return r.RateLimitKey()
	}
	if r, ok := req.(ttnpb.IDStringer); ok {
		if r.IDString() == "" {
			return fmt.Sprintf("%s:_", r.EntityType())
		}
		return fmt.Sprintf("%s:%s", r.EntityType(), unique.ID(ctx, r))
	}
	return ""
}

func grpcIsClusterAuthCall(ctx context.Context) bool {
	return rpcmetadata.FromIncomingContext(ctx).AuthType == clusterauth.AuthType
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that rate limits incoming gRPC requests.
func UnaryServerInterceptor(limiter Interface) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if grpcIsClusterAuthCall(ctx) {
			return handler(ctx, req)
		}

		resource := grpcMethodResource(ctx, info.FullMethod, req)
		limit, result := limiter.RateLimit(resource)
		if err := grpc.SetHeader(ctx, result.GRPCHeaders()); err != nil && !limit {
			log.FromContext(ctx).WithError(err).Error("Failed to set rate limit headers")
		}
		if limit {
			return nil, errRateLimitExceeded.WithAttributes("key", resource.Key(), "rate", result.Limit)
		}
		return handler(ctx, req)
	}
}

type rateLimitedServerStream struct {
	grpc.ServerStream

	limiter  Interface
	resource Resource
}

func (s *rateLimitedServerStream) RecvMsg(msg any) error {
	if err := Require(s.limiter, s.resource); err != nil {
		return err
	}
	return s.ServerStream.RecvMsg(msg)
}

// StreamServerInterceptor is a grpc.StreamServerInterceptor that rate limits new gRPC requests and messages sent by the client.
// If the X-Real-IP header is not set, it is assumed that the gRPC request originates from the cluster, and no rate limits are enforced.
func StreamServerInterceptor(limiter Interface) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()

		if grpcIsClusterAuthCall(ctx) {
			return handler(ctx, stream)
		}

		acceptResource := grpcStreamAcceptResource(ctx, info.FullMethod)
		limit, result := limiter.RateLimit(acceptResource)
		stream.SetHeader(result.GRPCHeaders())
		if limit {
			return errRateLimitExceeded.WithAttributes("key", acceptResource.Key(), "rate", result.Limit)
		}

		return handler(srv, &rateLimitedServerStream{
			ServerStream: stream,
			limiter:      limiter,
			resource:     grpcStreamUpResource(ctx, info.FullMethod),
		})
	}
}
