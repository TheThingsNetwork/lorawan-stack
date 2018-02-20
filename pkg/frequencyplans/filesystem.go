// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
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

	store, err := retrieveFrequencyPlans(config)
	if err != nil {
		return nil, errors.NewWithCause("Reading frequency plans from the file system failed", err)
	}

	return store, nil
}

func (config storeReadConfiguration) GetList() ([]frequencyPlanDescription, error) {
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

func (config storeReadConfiguration) GetFrequencyPlan(filename string) (ttnpb.FrequencyPlan, error) {
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
