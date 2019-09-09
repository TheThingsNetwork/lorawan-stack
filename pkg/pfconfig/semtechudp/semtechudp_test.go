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

package semtechudp

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func TestBuild(t *testing.T) {
	var fetcher fetch.Interface
	localPath := os.ExpandEnv("$GOPATH/src/github.com/TheThingsNetwork/lorawan-frequency-plans")
	if _, err := os.Stat(localPath); err == nil {
		fetcher = fetch.FromFilesystem(localPath)
	} else {
		var err error
		fetcher, err = fetch.FromHTTP("https://raw.githubusercontent.com/TheThingsNetwork/lorawan-frequency-plans/master", true)
		if err != nil {
			t.Fatalf("Failed to construct HTTP fetcher: %v", err)
		}
	}
	store := frequencyplans.NewStore(fetcher)

	var referenceFetcher fetch.Interface
	referenceLocalPath := os.ExpandEnv("$GOPATH/src/github.com/TheThingsNetwork/gateway-conf")
	if _, err := os.Stat(referenceLocalPath); err == nil {
		referenceFetcher = fetch.FromFilesystem(referenceLocalPath)
	} else {
		var err error
		referenceFetcher, err = fetch.FromHTTP("https://raw.githubusercontent.com/TheThingsNetwork/gateway-conf/master", true)
		if err != nil {
			t.Fatalf("Failed to construct HTTP fetcher: %v", err)
		}
	}

	var shouldResembleReference = func(actual interface{}, expected ...interface{}) string {
		referenceBytes, err := referenceFetcher.File(expected[0].(string))
		if err != nil {
			panic(err)
		}
		var expectedMap map[string]interface{}
		err = json.Unmarshal(referenceBytes, &expectedMap)
		if err != nil {
			panic(err)
		}
		removeDescs(expectedMap)
		configJSON, _ := json.Marshal(actual.(*Config))
		var actualMap map[string]interface{}
		json.Unmarshal(configJSON, &actualMap)
		return should.BeEmpty(pretty.Diff(actualMap, expectedMap))
	}

	for _, tt := range []struct {
		Name    string
		Gateway *ttnpb.Gateway
		Assert  func(a *assertions.Assertion, config *Config)
	}{
		{
			Name: "Reference: EU global_conf",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID:      "EU_863_870_TTN",
				GatewayServerAddress: "router.eu.thethings.network",
			},
			Assert: func(a *assertions.Assertion, config *Config) {
				a.So(config, shouldResembleReference, "EU-global_conf.json")
			},
		},
		{
			Name: "Reference: US global_conf",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID:      "US_902_928_FSB_2",
				GatewayServerAddress: "router.us.thethings.network",
			},
			Assert: func(a *assertions.Assertion, config *Config) {
				a.So(config, shouldResembleReference, "US-global_conf.json")
			},
		},
		{
			Name: "Reference: AU global_conf",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID:      "AU_915_928_FSB_2",
				GatewayServerAddress: "router.au.thethings.network",
			},
			Assert: func(a *assertions.Assertion, config *Config) {
				a.So(config, shouldResembleReference, "AU-global_conf.json")
			},
		},
		{
			Name: "Reference: AS1 global_conf",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID:      "AS_920_923_LBT",
				GatewayServerAddress: "router.as1.thethings.network",
			},
			Assert: func(a *assertions.Assertion, config *Config) {
				a.So(config, shouldResembleReference, "AS1-global_conf.json")
			},
		},
		{
			Name: "Reference: KR global_conf",
			Gateway: &ttnpb.Gateway{
				FrequencyPlanID:      "KR_920_923_TTN",
				GatewayServerAddress: "router.kr.thethings.network",
			},
			Assert: func(a *assertions.Assertion, config *Config) {
				a.So(config, shouldResembleReference, "KR-global_conf.json")
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			config, err := Build(tt.Gateway, store)
			if a.So(err, should.BeNil) {
				tt.Assert(a, config)
			}
		})
	}

}

func removeDescs(m map[string]interface{}) {
	for k, v := range m {
		if strings.HasSuffix(k, "desc") {
			delete(m, k)
			continue
		}
		if subMap, ok := v.(map[string]interface{}); ok {
			removeDescs(subMap)
		}
	}
}
