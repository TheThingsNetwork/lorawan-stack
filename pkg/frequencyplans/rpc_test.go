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

package frequencyplans_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRPCServer(t *testing.T) {
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{
		"frequency-plans.yml": []byte(`- id: A
  description: Frequency Plan A
  base-frequency: 868
  file: A.yml
- id: B
  base-id: A
  description: Frequency Plan B
  file: B.yml
- id: C
  description: Frequency Plan C
  base-frequency: 915
  file: C.yml`),
	}))

	server := frequencyplans.NewRPCServer(store)

	all, err := server.ListFrequencyPlans(context.Background(), &ttnpb.ListFrequencyPlansRequest{})
	a.So(err, should.BeNil)
	a.So(all.FrequencyPlans, should.HaveLength, 3)

	base915, err := server.ListFrequencyPlans(context.Background(), &ttnpb.ListFrequencyPlansRequest{
		BaseFrequency: 868,
	})
	a.So(err, should.BeNil)
	a.So(base915.FrequencyPlans, should.HaveLength, 2)
}
