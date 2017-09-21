// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type frequencyPlanDescription struct {
	// ID to identify the frequency plan
	ID string `yaml:"id"`
	// Description in Mhz
	Description string `yaml:"description"`
	// BaseFrequency in Mhz
	BaseFrequency uint8 `yaml:"base_freq"`
	// Filename of the frequency plan within the repo
	Filename string `yaml:"file"`
}

// Store of frequency plans
type Store interface {
	GetByID(id string) (ttnpb.FrequencyPlan, error)
	GetAllIDs() []string
}

type frequencyPlanStorage struct {
	ID      string
	Content ttnpb.FrequencyPlan
}

type store map[string]ttnpb.FrequencyPlan

func (s store) GetByID(id string) (ttnpb.FrequencyPlan, error) {
	if frequencyPlan, ok := s[id]; ok {
		return frequencyPlan, nil
	}
	return ttnpb.FrequencyPlan{}, errors.New("Frequency plan not found")
}

func (s store) GetAllIDs() []string {
	ids := []string{}

	for frequencyPlanID := range s {
		ids = append(ids, frequencyPlanID)
	}

	return ids
}
