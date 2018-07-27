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

// Package cluster contains cluster authentication-related utilities.
package cluster

import (
	"context"
	"crypto/subtle"
	"encoding/hex"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
)

type (
	clusterAuthKeyType        struct{}
	clusterAuthKeyFailureType struct{}
)

var (
	// AuthType used to identify components.
	AuthType = "ClusterKey"

	clusterAuthKey        = clusterAuthKeyType{}
	clusterAuthFailureKey = clusterAuthKeyFailureType{}

	errMissingClusterKey   = errors.DefineUnauthenticated("missing_cluster_key", "missing cluster key auth")
	errUnsupportedAuthType = errors.DefineInvalidArgument("auth_type", "cluster auth type `{auth_type}` is not supported")
	errInvalidClusterKey   = errors.DefinePermissionDenied("cluster_key", "invalid cluster key")
)

// NewContext returns a context containing the cluster authentication result.
func NewContext(ctx context.Context, err error) context.Context {
	ctx = context.WithValue(ctx, clusterAuthKey, err == nil)
	ctx = context.WithValue(ctx, clusterAuthFailureKey, err)
	return ctx
}

// VerifySource inspects whether the context contains one of the passed cluster keys,
// and returns a context containing the result.
func VerifySource(ctx context.Context, validKeys [][]byte) context.Context {
	err := verifySource(ctx, validKeys)
	return NewContext(ctx, err)
}

func verifySource(ctx context.Context, validKeys [][]byte) error {
	md := rpcmetadata.FromIncomingContext(ctx)
	switch md.AuthType {
	case AuthType:
	case "":
		return errMissingClusterKey
	default:
		return errUnsupportedAuthType.WithAttributes("auth_type", md.AuthType)
	}
	key, err := hex.DecodeString(md.AuthValue)
	if err != nil {
		return errInvalidClusterKey.WithCause(err)
	}
	for _, acceptedKey := range validKeys {
		if subtle.ConstantTimeCompare(acceptedKey, key) == 1 {
			return nil
		}
	}
	return errInvalidClusterKey
}

// Authorized returns whether the context has been identified as a cluster call.
// It panics if it does not inherit from `NewContext`.
func Authorized(ctx context.Context) error {
	ok, isStored := ctx.Value(clusterAuthKey).(bool)
	if !isStored {
		panic("call source not verified; did you register the cluster authentication hook for this call?")
	}
	if ok {
		return nil
	}
	return ctx.Value(clusterAuthFailureKey).(error)
}
