// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"fmt"
	"io/ioutil"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"gopkg.in/yaml.v2"
)

// ReadFileSystemStore returns a new Store of frequency plans, based on the options given, read from the local filesystem.
//
// By default, ReadFileSystemStore reads the files in the current directory. It is possible to specify a root directory through options.
func ReadFileSystemStore(options ...ReadFileSystemStoreOption) (Store, error) {
	config := storeReadConfiguration{}

	for _, option := range options {
		option(&config)
	}

	store, err := readFileSystemStore(config)
	if err != nil {
		return nil, errors.NewWithCause("Reading frequency plans from the file system failed", err)
	}

	return store, nil
}

func readList(config storeReadConfiguration) ([]frequencyPlanDescription, error) {
	frequencyPlanListPath := config.AbsolutePath(DefaultListFilename)

	content, err := ioutil.ReadFile(frequencyPlanListPath)
	if err != nil {
		return nil, errors.NewWithCause("Reading frequency plans list failed", err)
	}

	list := []frequencyPlanDescription{}
	err = yaml.Unmarshal(content, &list)
	if err != nil {
		return nil, errors.NewWithCause("Failed to parse the file content as a list of frequency plans", err)
	}

	return list, nil
}

func readFrequencyPlan(config storeReadConfiguration, filename string) (ttnpb.FrequencyPlan, error) {
	frequencyPlanPath := config.AbsolutePath(filename)
	frequencyPlan := ttnpb.FrequencyPlan{}

	content, err := ioutil.ReadFile(frequencyPlanPath)
	if err != nil {
		return frequencyPlan, errors.NewWithCause("Reading frequency plan failed", err)
	}

	err = yaml.Unmarshal(content, &frequencyPlan)
	if err != nil {
		return frequencyPlan, errors.NewWithCause("Failed to parse the file content as a frequency plan", err)
	}

	return frequencyPlan, nil
}

func readFileSystemStore(config storeReadConfiguration) (store, error) {
	frequencyPlansInfo, err := readList(config)
	if err != nil {
		return nil, errors.NewWithCause("Fetching list of frequency plans failed", err)
	}

	frequencyPlansStorage := make(store)
	for _, frequencyPlanDescription := range frequencyPlansInfo {
		frequencyPlanContent, err := readFrequencyPlan(config, frequencyPlanDescription.Filename)
		if err != nil {
			return nil, errors.NewWithCause(fmt.Sprintf("Failed to retrieve %s frequency plan content", frequencyPlanDescription.ID), err)
		}

		frequencyPlansStorage[frequencyPlanDescription.ID] = frequencyPlanContent
	}

	return frequencyPlansStorage, nil
}
