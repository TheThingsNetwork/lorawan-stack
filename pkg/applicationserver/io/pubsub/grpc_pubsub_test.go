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

package pubsub_test

import (
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestPubSubRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	ctx := newContextWithRightsFetcher(test.Context())

	redisClient, flush := test.NewRedis(t, "applicationserver_test")
	defer flush()
	defer redisClient.Close()
	pubsubReg := &redis.PubSubRegistry{Redis: redisClient}
	srv := pubsub.NewPubSubRegistryRPC(pubsubReg)
	authorizedCtx := contextWithKey(ctx, registeredApplicationKey)

	// Formats.
	{
		res, err := srv.GetFormats(authorizedCtx, ttnpb.Empty)
		a.So(err, should.BeNil)
		a.So(res.Formats, should.HaveSameElementsDeep, map[string]string{
			"json": "JSON",
		})
	}

	// Check empty.
	{
		res, err := srv.List(authorizedCtx, &ttnpb.ListApplicationPubSubsRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"provider"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Pubsubs, should.BeEmpty)
	}

	// Add.
	{
		_, err := srv.Set(authorizedCtx, &ttnpb.SetApplicationPubSubRequest{
			ApplicationPubSub: ttnpb.ApplicationPubSub{
				ApplicationPubSubIdentifiers: ttnpb.ApplicationPubSubIdentifiers{
					ApplicationIdentifiers: registeredApplicationID,
					PubSubID:               registeredPubSubID,
				},
				Provider: ttnpb.ApplicationPubSub_AWSSNSSQS,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"provider"},
			},
		})
		a.So(err, should.BeNil)
	}

	// List; assert one.
	{
		res, err := srv.List(authorizedCtx, &ttnpb.ListApplicationPubSubsRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"provider"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Pubsubs, should.HaveLength, 1)
		a.So(res.Pubsubs[0].Provider, should.Equal, ttnpb.ApplicationPubSub_AWSSNSSQS)
	}

	// Get.
	{
		res, err := srv.Get(authorizedCtx, &ttnpb.GetApplicationPubSubRequest{
			ApplicationPubSubIdentifiers: ttnpb.ApplicationPubSubIdentifiers{
				ApplicationIdentifiers: registeredApplicationID,
				PubSubID:               registeredPubSubID,
			},
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"provider"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Provider, should.Equal, ttnpb.ApplicationPubSub_AWSSNSSQS)
	}

	// Delete.
	{
		_, err := srv.Delete(authorizedCtx, &ttnpb.ApplicationPubSubIdentifiers{
			ApplicationIdentifiers: registeredApplicationID,
			PubSubID:               registeredPubSubID,
		})
		a.So(err, should.BeNil)
	}

	// Check empty.
	{
		res, err := srv.List(authorizedCtx, &ttnpb.ListApplicationPubSubsRequest{
			ApplicationIdentifiers: registeredApplicationID,
			FieldMask: pbtypes.FieldMask{
				Paths: []string{"provider"},
			},
		})
		a.So(err, should.BeNil)
		a.So(res.Pubsubs, should.BeEmpty)
	}
}
