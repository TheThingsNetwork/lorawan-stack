// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package gatewaytokens provides functions to work with GatewayTokens.
package gatewaytokens

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	// AuthType is the type of authentication.
	AuthType = "GatewayToken"

	errChecksumMismatch = errors.DefineAborted("checksum_mismatch", "checksum mismatch")
	errTokenExpired     = errors.DefineAborted("token_expired", "token expired")
)

type gatewayTokenContextKeyType struct{}

var gatewayTokenContextKey gatewayTokenContextKeyType

// FromContext returns the gatewaytokens.Token from the context.
func FromContext(ctx context.Context) (*Token, bool) {
	if t, ok := ctx.Value(gatewayTokenContextKey).(*Token); ok {
		return t, true
	}
	return nil, false
}

// NewContext returns a new context with the given gatewaytokens.Token.
func NewContext(ctx context.Context, token *Token) context.Context {
	return context.WithValue(ctx, gatewayTokenContextKey, token)
}

// KeyService provides HMAC hashing.
type KeyService interface {
	HMACHash(ctx context.Context, payload []byte, id string) ([]byte, error)
}

// Token wraps ttnpb.GatewayToken with additional semantics.
type Token struct {
	token *ttnpb.GatewayToken
	ks    KeyService
}

// New generates a new Token with the given information.
func New(
	keyID string,
	ids *ttnpb.GatewayIdentifiers,
	rights *ttnpb.Rights,
	ks KeyService,
) *Token {
	return &Token{
		token: &ttnpb.GatewayToken{
			KeyId: keyID,
			Payload: &ttnpb.GatewayToken_Payload{
				GatewayIds: ids,
				Rights:     rights,
			},
		},
		ks: ks,
	}
}

// Generate generates a new ttnpb.GatewayToken with the checksum calculated and the timestamp set.
// This functions allows generating multiple tokens with the same information but different timestamps.
func (t *Token) Generate(ctx context.Context) (ret *ttnpb.GatewayToken, err error) {
	ret = ttnpb.Clone(t.token)
	ret.Payload.CreatedAt = timestamppb.Now()
	enc, err := proto.Marshal(ret.Payload)
	if err != nil {
		return nil, err
	}
	ret.Checksum, err = t.ks.HMACHash(ctx, enc, t.token.KeyId)
	if err != nil {
		return nil, err
	}
	return
}

// Verify verifies the hash of the payload of a GatewayToken.
// If verified, the rights embedded in the token are retrieved.
// The function also checks if the token is still valid.
func Verify(
	ctx context.Context, token *ttnpb.GatewayToken, validity time.Duration, ks KeyService,
) (*ttnpb.Rights, error) {
	enc, err := proto.Marshal(token.Payload)
	if err != nil {
		return nil, err
	}
	createdAt := time.Unix(token.GetPayload().CreatedAt.Seconds, 0)
	if time.Now().After(createdAt.Add(validity)) {
		return nil, errTokenExpired.New()
	}
	checksum, err := ks.HMACHash(ctx, enc, token.KeyId)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare(checksum, token.Checksum) == 0 {
		return nil, errChecksumMismatch.New()
	}
	return token.Payload.Rights, nil
}

// DecodeFromString decodes the GatewayToken from a hex encoded string.
func DecodeFromString(s string) (*ttnpb.GatewayToken, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	token := ttnpb.GatewayToken{}
	if err := proto.Unmarshal(b, &token); err != nil {
		return nil, err
	}
	return &token, err
}

// AuthenticatedContext checks the context for a gatewaytokens.Token.
// If it exists, it generates a new GatewayToken with timestamp set and the hash calculated.
// The function returns a context with the GatewayToken as a metadata item.
// If there is no gatewaytokens.Token in the context, the function returns the original context.
func AuthenticatedContext(ctx context.Context) (context.Context, error) {
	t, ok := FromContext(ctx)
	if !ok {
		return ctx, nil
	}
	token, err := t.Generate(ctx)
	if err != nil {
		return nil, err
	}
	msg, err := proto.Marshal(token)
	if err != nil {
		return nil, err
	}

	md := metadata.New(map[string]string{
		"authorization": fmt.Sprintf("%s %s", AuthType, hex.EncodeToString(msg)),
	})
	return metadata.NewIncomingContext(ctx, md), nil
}
