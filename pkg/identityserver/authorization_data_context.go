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

package identityserver

import "context"

type authorizationDataKey struct{}

// newContextWithAuthorizationData returns a new context but with the provided
// authorization data.
func newContextWithAuthorizationData(ctx context.Context, ad *authorizationData) context.Context {
	return context.WithValue(ctx, authorizationDataKey{}, ad)
}

// authorizationDataFromContext returns the authorization data from the context.
// If not found, it returns an empty `authorizationData`.
func authorizationDataFromContext(ctx context.Context) *authorizationData {
	ad, ok := ctx.Value(authorizationDataKey{}).(*authorizationData)
	if !ok {
		return new(authorizationData)
	}
	return ad
}
