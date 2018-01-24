// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type frequencyPlanDescription struct {
	// ID to identify the frequency plan
	ID string `yaml:"id"`
	// Description in Mhz
	Description string `yaml:"description"`
	// BaseFrequency in Mhz
	BaseFrequency uint16 `yaml:"base_freq"`
	// Filename of the frequency plan within the repo
	FPFilename string `yaml:"file"`

	// BaseID is the ID of the frequency plan that's the basis for this extended frequency plan
	BaseID string `yaml:"base,omitempty"`
}

type retrievalConfig interface {
	GetFrequencyPlan(filename string) (ttnpb.FrequencyPlan, error)
	GetList() ([]frequencyPlanDescription, error)
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

func retrieveFrequencyPlans(config retrievalConfig) (store, error) {
	frequencyPlansInfo, err := config.GetList()
	if err != nil {
		return nil, errors.NewWithCause("Failed to fetch list of frequency plans", err)
	}
	frequencyPlansExtensions := make([]frequencyPlanDescription, 0)

	frequencyPlansStorage := make(store)
	for _, description := range frequencyPlansInfo {
		if description.BaseID != "" {
			frequencyPlansExtensions = append(frequencyPlansExtensions, description)
			continue
		}

		frequencyPlanContent, err := config.GetFrequencyPlan(description.FPFilename)
		if err != nil {
			return nil, errors.NewWithCause(fmt.Sprintf("Failed to retrieve %s frequency plan content", description.ID), err)
		}

		frequencyPlansStorage[description.ID] = frequencyPlanContent
	}

	for _, extensionDescription := range frequencyPlansExtensions {
		originFrequencyPlan, ok := frequencyPlansStorage[extensionDescription.BaseID]
		if !ok {
			return nil, fmt.Errorf("Could not find original frequency plan %s for frequency plan extension %s", extensionDescription.BaseID, extensionDescription.ID)
		}

		extensionContent, err := config.GetFrequencyPlan(extensionDescription.FPFilename)
		if err != nil {
			return nil, errors.NewWithCause(fmt.Sprintf("Failed to retrieve %s frequency plan extension content", extensionDescription.ID), err)
		}

		frequencyPlansStorage[extensionDescription.ID] = originFrequencyPlan.Extend(extensionContent)
	}

	return frequencyPlansStorage, nil
}
