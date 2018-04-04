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

package gatewayserver_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func removeFPStore(a *assertions.Assertion, dir string) {
	err := os.RemoveAll(dir)
	a.So(err, should.BeNil)
}

func createFPStore(a *assertions.Assertion) string {
	dir, err := ioutil.TempDir("", "gs-frequency-plans-store")
	a.So(err, should.BeNil)

	f, err := os.Create(filepath.Join(dir, "frequency-plans.yml"))
	a.So(err, should.BeNil)
	_, err = f.Write([]byte(`- id: EU_863_870
  description: Europe 868MHz
  base_freq: 868
  file: EU_863_870.yml
- id: KR_920_923
  description: Korea 920-923MHz
  base_freq: 915
  file: KR_920_923.yml`))
	a.So(err, should.BeNil)
	err = f.Close()
	a.So(err, should.BeNil)

	f, err = os.Create(filepath.Join(dir, "EU_863_870.yml"))
	a.So(err, should.BeNil)
	_, err = f.Write([]byte(`band-id: EU_863_870
channels:
  - frequency: 867100000
  - frequency: 867300000
  - frequency: 867500000
  - frequency: 867700000
  - frequency: 867900000
  - frequency: 868100000
  - frequency: 868300000
  - frequency: 868500000
lora-std-channel:
  frequency: 863000000
  data-rate:
    index: 6
fsk-channel:
  frequency: 868800000
  data-rate:
    index: 7`))
	a.So(err, should.BeNil)
	err = f.Close()
	a.So(err, should.BeNil)

	f, err = os.Create(filepath.Join(dir, "KR_920_923.yml"))
	a.So(err, should.BeNil)
	_, err = f.Write([]byte(`band-id: KR_920_923
channels:
- frequency: 922100000
- frequency: 922300000
- frequency: 922500000
- frequency: 922700000
- frequency: 922900000
- frequency: 923100000
- frequency: 923300000
lbt:
  rssi-target: -80
  scan-time: 128`))
	a.So(err, should.BeNil)
	err = f.Close()
	a.So(err, should.BeNil)

	return dir
}
