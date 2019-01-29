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
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
	ttnblob "go.thethings.network/lorawan-stack/pkg/blob"
	"go.thethings.network/lorawan-stack/pkg/identityserver/picture"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/randutil"
)

const maxProfilePictureStoredDimensions = 1024

var profilePictureDimensions = []int{64, 128, 256, 512}

var profilePictureRand = rand.New(randutil.NewLockedSource(rand.NewSource(time.Now().UnixNano())))

func (is *IdentityServer) processUserProfilePicture(ctx context.Context, usr *ttnpb.User) (err error) {
	// External pictures, consider only largest.
	if usr.ProfilePicture.Sizes != nil {
		original := usr.ProfilePicture.Sizes[0]
		if original == "" {
			var max uint32
			for size, url := range usr.ProfilePicture.Sizes {
				if size > max {
					max = size
					original = url
				}
			}
		}
		if original != "" {
			usr.ProfilePicture.Sizes = map[uint32]string{0: original}
		} else {
			usr.ProfilePicture.Sizes = nil
		}
	}

	// Embedded (uploaded) picture. Make square.
	if usr.ProfilePicture.Embedded != nil && len(usr.ProfilePicture.Embedded.Data) > 0 {
		usr.ProfilePicture, err = picture.MakeSquare(bytes.NewBuffer(usr.ProfilePicture.Embedded.Data), maxProfilePictureStoredDimensions)
		if err != nil {
			return err
		}
	}

	// Store picture to bucket.
	bucket, err := ttnblob.Config(is.Component.GetBaseConfig(ctx).Blob).GetBucket(ctx, is.configFromContext(ctx).ProfilePicture.Bucket)
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s.%s", unique.ID(ctx, usr.UserIdentifiers), ulid.MustNew(ulid.Now(), profilePictureRand).String())
	usr.ProfilePicture, err = picture.Store(ctx, bucket, id, usr.ProfilePicture, profilePictureDimensions...)
	if err != nil {
		return err
	}

	return
}
