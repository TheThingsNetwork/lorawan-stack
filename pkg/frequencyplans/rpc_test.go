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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestRPCServer(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	store := frequencyplans.NewStore(fetch.NewMemFetcher(map[string][]byte{
		"frequency-plans.yml": []byte(`- id: A
  band-id: EU_863_870
  description: Frequency Plan A
  base-frequency: 868
  file: A.yml
- id: B
  band-id: AS_923
  base-id: A
  description: Frequency Plan B
  file: B.yml
- id: C
  band-id: US_902_928
  description: Frequency Plan C
  base-frequency: 915
  file: C.yml`),
	}))

	server := frequencyplans.NewRPCServer(store)

	expectedAll := []*ttnpb.FrequencyPlanDescription{
		{
			Id:            "A",
			BandId:        "EU_863_870",
			BaseFrequency: 868,
		},
		{
			Id:            "B",
			BaseId:        "A",
			BaseFrequency: 868,
			BandId:        "AS_923",
		},
		{
			Id:            "C",
			BaseFrequency: 915,
			BandId:        "US_902_928",
		},
	}

	actualAll, err := server.ListFrequencyPlans(context.Background(), &ttnpb.ListFrequencyPlansRequest{})
	a.So(err, should.BeNil)
	a.So(actualAll.FrequencyPlans, should.HaveLength, 3)
	a.So(actualAll.FrequencyPlans[0], should.Resemble, expectedAll[0])

	base915, err := server.ListFrequencyPlans(context.Background(), &ttnpb.ListFrequencyPlansRequest{
		BaseFrequency: 868,
	})
	a.So(err, should.BeNil)
	a.So(base915.FrequencyPlans, should.HaveLength, 2)
	a.So(base915.FrequencyPlans, should.Resemble, expectedAll[:2])

	bandAS, err := server.ListFrequencyPlans(context.Background(), &ttnpb.ListFrequencyPlansRequest{
		BandId: "AS_923",
	})
	a.So(err, should.BeNil)
	a.So(bandAS.FrequencyPlans, should.HaveLength, 1)
	a.So(bandAS.FrequencyPlans[0], should.Resemble, expectedAll[1])
}
