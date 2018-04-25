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

// Package frequencyplans contains abstractions to fetch and manipulate frequency plans.
package frequencyplans

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/fetch"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	yaml "gopkg.in/yaml.v2"
)

const yamlFetchErrorCache = 1 * time.Minute

var (
	// ErrFrequencyPlanNotFound is returned when a frequency plan could not be found.
	ErrFrequencyPlanNotFound = &errors.ErrDescriptor{
		MessageFormat:  "Frequency plan `{frequency_plan_id}` not found",
		Code:           1,
		Type:           errors.NotFound,
		SafeAttributes: []string{"frequency_plan_id"},
	}
	// ErrFetchFailed is returned when a frequency plan could not be fetched.
	ErrFetchFailed = &errors.ErrDescriptor{
		MessageFormat: "Fetching the file failed",
		Code:          2,
		Type:          errors.External,
	}
	// ErrIDNotSpecified is returned when the ID of a frequency plan to return was not specified.
	ErrIDNotSpecified = &errors.ErrDescriptor{
		MessageFormat: "ID of the frequency plan to return not specified",
		Code:          2,
		Type:          errors.InvalidArgument,
	}
)

func init() {
	ErrFrequencyPlanNotFound.Register()
	ErrFetchFailed.Register()
	ErrIDNotSpecified.Register()
}

// FrequencyPlanDescription describes a frequency plan in the YAML format.
type FrequencyPlanDescription struct {
	// ID to identify the frequency plan.
	ID string `yaml:"id"`
	// Description of the frequency plan.
	Description string `yaml:"description"`
	// BaseFrequency in Mhz.
	BaseFrequency uint16 `yaml:"base_freq"`
	// Filename of the frequency plan within the repo.
	Filename string `yaml:"file"`

	// BaseID is the ID of the frequency plan that's the basis for this extended frequency plan.
	BaseID string `yaml:"base,omitempty"`
}

func (d FrequencyPlanDescription) content(f fetch.Interface) ([]byte, error) {
	return f.File(d.Filename)
}

func (d FrequencyPlanDescription) proto(f fetch.Interface) (ttnpb.FrequencyPlan, error) {
	fp := ttnpb.FrequencyPlan{}

	content, err := d.content(f)
	if err != nil {
		return fp, err
	}

	if err := yaml.Unmarshal(content, &fp); err != nil {
		return fp, err
	}

	return fp, nil
}

type frequencyPlanList []FrequencyPlanDescription

func (l frequencyPlanList) get(id string) (FrequencyPlanDescription, bool) {
	for _, f := range l {
		if f.ID == id {
			return f, true
		}
	}

	return FrequencyPlanDescription{}, false
}

type queryResult struct {
	fp   ttnpb.FrequencyPlan
	err  error
	time time.Time
}

// Store of frequency plans.
type Store struct {
	// Fetcher is the fetch.Interface used to retrieve data.
	Fetcher fetch.Interface

	descriptionsMu             sync.Mutex
	descriptionsCache          frequencyPlanList
	descriptionsFetchErrorTime time.Time
	descriptionsFetchError     error

	frequencyPlansCache map[string]queryResult
	frequencyPlansMu    sync.Mutex
}

// NewStore of frequency plans.
func NewStore(fetcher fetch.Interface) *Store {
	return &Store{
		Fetcher: fetcher,

		frequencyPlansCache: map[string]queryResult{},
	}
}

func (s *Store) fetchDescriptions() (frequencyPlanList, error) {
	content, err := s.Fetcher.File("frequency-plans.yml")
	if err != nil {
		return nil, err
	}

	descriptions := frequencyPlanList{}
	if err = yaml.Unmarshal(content, &descriptions); err != nil {
		return nil, errors.NewWithCause(err, "Failed to parse the file content as a list of frequency plans")
	}

	return descriptions, nil
}

func (s *Store) descriptions() (frequencyPlanList, error) {
	s.descriptionsMu.Lock()
	defer s.descriptionsMu.Unlock()
	if s.descriptionsCache != nil {
		return s.descriptionsCache, nil
	}

	if time.Since(s.descriptionsFetchErrorTime) < yamlFetchErrorCache {
		return nil, ErrFetchFailed.NewWithCause(nil, s.descriptionsFetchError)
	}

	descriptions, err := s.fetchDescriptions()
	if err != nil {
		s.descriptionsFetchError = err
		s.descriptionsFetchErrorTime = time.Now()
		return nil, ErrFetchFailed.NewWithCause(nil, err)
	}

	s.descriptionsFetchErrorTime = time.Time{}
	s.descriptionsFetchError = nil
	s.descriptionsCache = descriptions
	return descriptions, nil
}

// getByID returns the frequency plan associated to that ID.
func (s *Store) getByID(id string) (proto ttnpb.FrequencyPlan, err error) {
	descriptions, err := s.descriptions()
	if err != nil {
		return ttnpb.FrequencyPlan{}, errors.NewWithCause(err, "Could not read the list of frequency plans filenames")
	}

	description, ok := descriptions.get(id)
	if !ok {
		return ttnpb.FrequencyPlan{}, ErrFrequencyPlanNotFound.New(errors.Attributes{
			"frequency_plan_id": id,
		})
	}

	proto, err = description.proto(s.Fetcher)
	if err != nil {
		return
	}

	if description.BaseID != "" {
		base, ok := descriptions.get(description.BaseID)
		if !ok {
			return ttnpb.FrequencyPlan{}, errors.NewWithCausef(ErrFrequencyPlanNotFound.New(errors.Attributes{
				"frequency_plan_id": description.BaseID,
			}), "Could not found the base of the frequency plan %s", id)
		}

		var baseProto ttnpb.FrequencyPlan
		baseProto, err = base.proto(s.Fetcher)
		if err != nil {
			return
		}

		proto = baseProto.Extend(proto)
	}

	return
}

// GetByID tries to retrieve the frequency plan that has the given ID, and returns an error otherwise.
func (s *Store) GetByID(id string) (ttnpb.FrequencyPlan, error) {
	if id == "" {
		return ttnpb.FrequencyPlan{}, ErrIDNotSpecified.New(nil)
	}

	s.frequencyPlansMu.Lock()
	defer s.frequencyPlansMu.Unlock()
	if cached, ok := s.frequencyPlansCache[id]; ok && cached.err == nil || time.Since(cached.time) < yamlFetchErrorCache {
		return cached.fp, cached.err
	}
	proto, err := s.getByID(id)
	s.frequencyPlansCache[id] = queryResult{
		time: time.Now(),
		fp:   proto,
		err:  err,
	}

	return proto, err
}

// GetAllIDs returns the list of IDs of the available frequency plans.
func (s *Store) GetAllIDs() ([]string, error) {
	descriptions, err := s.descriptions()
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, description := range descriptions {
		ids = append(ids, description.ID)
	}

	return ids, nil
}
