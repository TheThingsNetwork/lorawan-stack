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

package test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

const (
	frequencyPlansDescription = `- id: EU_863_870
  description: Europe 868MHz
  base_freq: 868
  file: EU_863_870.yml
- id: KR_920_923
  description: Korea 920-923MHz
  base_freq: 915
  file: KR_920_923.yml`

	// EUFrequencyPlanID available in the store.
	EUFrequencyPlanID = "EU_863_870"
	euFrequencyPlan   = `band-id: EU_863_870
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
    index: 7`

	// KRFrequencyPlanID available in the store.
	KRFrequencyPlanID = "KR_920_923"
	krFrequencyPlan   = `band-id: KR_920_923
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
  scan-time: 128`
)

// FrequencyPlansStore containing several frequency plans in order to use
// local frequency plans rather than the production ones stored on GitHub.
type FrequencyPlansStore string

// NewFrequencyPlansStore returns a new frequency plans store.
func NewFrequencyPlansStore() (FrequencyPlansStore, error) {
	var store FrequencyPlansStore

	dir, err := ioutil.TempDir("", "frequencyplans")
	if err != nil {
		return store, errors.NewWithCause(err, "Failed to create a new temporary directory for frequency plans")
	}
	store = FrequencyPlansStore(dir)

	defer func() {
		if err != nil {
			os.RemoveAll(dir)
		}
	}()

	for _, document := range []struct {
		filename, content string
	}{
		{"frequency-plans.yml", frequencyPlansDescription},
		{"EU_863_870.yml", euFrequencyPlan},
		{"KR_920_923.yml", krFrequencyPlan},
	} {
		f, fileErr := os.Create(filepath.Join(dir, document.filename))
		if fileErr != nil {
			err = fileErr
			break
		}

		_, err = f.Write([]byte(document.content))
		if err != nil {
			break
		}

		err = f.Close()
		if err != nil {
			break
		}
	}

	return store, err
}

// Directory where the frequency plans are stored.
func (f FrequencyPlansStore) Directory() string {
	return string(f)
}

// Destroy the store. This should be called at the end of every lifecycle.
//
// If the fetcher is used after Destroy() is called, the fetcher will always
// return an error.
func (f FrequencyPlansStore) Destroy() error {
	return os.RemoveAll(string(f))
}
