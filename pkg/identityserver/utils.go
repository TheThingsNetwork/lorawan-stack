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

package identityserver

import (
	"context"
	"strconv"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func cleanContactInfo(info []*ttnpb.ContactInfo) {
	for _, info := range info {
		info.ValidatedAt = nil
	}
}

func validateContactInfo(info []*ttnpb.ContactInfo) error {
	for _, info := range info {
		if info.ContactMethod != ttnpb.CONTACT_METHOD_EMAIL {
			continue
		}
		if err := validate.Email(info.Value); err != nil {
			return err
		}
	}
	return nil
}

func setTotalHeader(ctx context.Context, total uint64) {
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(total, 10)))
}
