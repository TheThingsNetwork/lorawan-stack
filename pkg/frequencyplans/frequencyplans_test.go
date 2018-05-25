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

package frequencyplans_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func Example() {
	store := frequencyplans.NewStore(fetch.FromHTTP("https://raw.githubusercontent.com/TheThingsNetwork/gateway-conf/yaml-master", true))

	ids, err := store.GetAllIDs()
	if err != nil {
		panic(err)
	}

	fmt.Println("Frequency plans available:")
	for _, id := range ids {
		fmt.Println("-", id)
	}

	euFP, err := store.GetByID("EU_863_870")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content of the EU frequency plan:")
	fmt.Println(euFP.String())
}

func TestFailToLoadDescriptions(t *testing.T) {
	a := assertions.New(t)

	dir, err := ioutil.TempDir("", "")
	if !a.So(err, should.BeNil) {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// Fill frequency plans store
	{
		err = ioutil.WriteFile(
			filepath.Join(dir, "frequency-plans.yml"),
			[]byte(`- id: AS_923
        description: Asian plan
base_freq: 923
  file: AS_923.yml
`),
			0755,
		)
		if !a.So(err, should.BeNil) {
			panic(err)
		}
	}

	s := frequencyplans.NewStore(fetch.FromFilesystem(dir))

	// GetAllIDs
	{
		_, err := s.GetAllIDs()
		a.So(err, should.NotBeNil)
	}
}

func TestStore(t *testing.T) {
	a := assertions.New(t)

	dir, err := ioutil.TempDir("", "")
	if !a.So(err, should.BeNil) {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// Fill frequency plans store
	{
		err = ioutil.WriteFile(
			filepath.Join(dir, "frequency-plans.yml"),
			[]byte(`- id: AS_923
  description: Asian plan
  base_freq: 923
  file: AS_923.yml
- id: JP
  description: Japanese plan
  base_freq: 923
  file: JP.yml
  base: AS_923
- id: EU_863_870
  description: European frequency plan (Not saved on disk)
  file: EU.yml
  base_freq: 863
- id: US_915
  description: US frequency plan (Invalid YAML)
  file: US.yml
  base_freq: 915
- id: SA
  description: South African frequency plan (Inexistant base)
  file: AS_923.yml
  base_freq: 863
  base: AFRICA
- id: CA
  description: Canadian frequency plan (Base has invalid yaml)
  file: EU.yml
  base_freq: 863
  base: US_915
`),
			0755,
		)
		if !a.So(err, should.BeNil) {
			panic(err)
		}

		err = ioutil.WriteFile(
			filepath.Join(dir, "AS_923.yml"),
			[]byte(`band-id: AS_923
channels: [{frequency: 923000000}]`),
			0755,
		)
		if !a.So(err, should.BeNil) {
			panic(err)
		}

		err = ioutil.WriteFile(
			filepath.Join(dir, "US.yml"),
			[]byte(`band-id: US_915
            channels: [{frequency: 915000000}]`),
			0755,
		)
		if !a.So(err, should.BeNil) {
			panic(err)
		}

		err = ioutil.WriteFile(
			filepath.Join(dir, "JP.yml"),
			[]byte(`lbt: {rssi-target: 1.1, rssi-offset: 2.2, scan-time: 80}`),
			0755,
		)
		if !a.So(err, should.BeNil) {
			panic(err)
		}
	}

	s := frequencyplans.NewStore(fetch.FromFilesystem(dir))

	// GetAllIDs
	{
		ids, err := s.GetAllIDs()
		if !a.So(err, should.BeNil) {
			panic(err)
		}

		a.So(ids, should.Contain, "AS_923")
		a.So(ids, should.Contain, "JP")
	}

	assertAS923Content := func(fp ttnpb.FrequencyPlan) {
		a.So(fp.Channels, should.HaveLength, 1)
		a.So(fp.Channels[0].Frequency, should.Equal, 923000000)
		a.So(fp.BandID, should.Equal, "AS_923")
	}

	// AS923 Frequency plan
	{
		fp, err := s.GetByID("AS_923")
		if !a.So(err, should.BeNil) {
			panic(err)
		}

		assertAS923Content(fp)
	}

	// JP Frequency plan
	{
		fp, err := s.GetByID("JP")
		if !a.So(err, should.BeNil) {
			panic(err)
		}

		assertAS923Content(fp)
		a.So(fp.LBT, should.NotBeNil)
		a.So(fp.LBT.RSSIOffset, should.AlmostEqual, 2.2, 0.00001)
		a.So(fp.LBT.ScanTime, should.Equal, 80)
	}

	// Unknown frequency plan
	{
		_, err = s.GetByID("Unknown")
		a.So(err, should.NotBeNil)
	}

	// Frequency plan not saved
	{
		_, err = s.GetByID("EU_863_870")
		a.So(err, should.NotBeNil)
	}

	// Frequency plan with invalid YAML
	{
		_, err = s.GetByID("US_915")
		a.So(err, should.NotBeNil)
	}

	// Frequency plan with inexistant base
	{
		_, err = s.GetByID("SA")
		a.So(err, should.NotBeNil)
	}

	// Frequency plan with base with invalid YAML
	{
		_, err = s.GetByID("CA")
		a.So(err, should.NotBeNil)
	}
}

func TestNotInitializedStore(t *testing.T) {
	a := assertions.New(t)

	dir, err := ioutil.TempDir("", "")
	if !a.So(err, should.BeNil) {
		panic(err)
	}
	defer os.RemoveAll(dir)

	s := frequencyplans.NewStore(fetch.FromFilesystem(dir))

	_, err = s.GetAllIDs()
	a.So(err, should.NotBeNil)

	_, err = s.GetByID("EU")
	a.So(err, should.NotBeNil)
}
