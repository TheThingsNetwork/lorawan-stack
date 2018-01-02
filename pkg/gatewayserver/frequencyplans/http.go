// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	yaml "gopkg.in/yaml.v2"
)

// RetrieveHTTPStore returns a new Store of frequency plans, based on the options given, fetched from a HTTP server.
//
// By default, RetrieveHTTPStore fetches its frequency plans on GitHub in the TheThingsNetwork/gateway-conf repository, in the yaml-master branch. It is possible to specify another HTTP root through options.
func RetrieveHTTPStore(options ...RetrieveHTTPStoreOption) (Store, error) {
	baseURI := DefaultBaseURL

	for _, option := range options {
		option(&baseURI)
	}

	store, err := retrieveFrequencyPlans(baseURI)
	return &store, err
}

func fetchHTTPContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.NewWithCause("HTTP request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Error with the HTTP exchange: expected successful status code, received %s status", resp.Status)
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewWithCause("Copying the HTTP response to a local buffer failed", err)
	}

	return buffer, nil
}

func (baseURI storeFetchingConfiguration) GetList() ([]frequencyPlanDescription, error) {
	list := make([]frequencyPlanDescription, 0)

	url := fmt.Sprintf("%s/%s", string(baseURI), "frequency-plans.yml")

	buffer, err := fetchHTTPContent(url)
	if err != nil {
		return nil, errors.NewWithCause("Fetching content failed", err)
	}

	err = yaml.Unmarshal(buffer, &list)
	if err != nil {
		return nil, errors.NewWithCause("Failed to parse the HTTP content as a list of frequency plans", err)
	}

	return list, nil
}

func (baseURI storeFetchingConfiguration) GetFrequencyPlan(filename string) (ttnpb.FrequencyPlan, error) {
	frequencyPlan := ttnpb.FrequencyPlan{}

	url := fmt.Sprintf("%s/%s", string(baseURI), filename)

	buffer, err := fetchHTTPContent(url)
	if err != nil {
		return frequencyPlan, errors.NewWithCause("Fetching content failed", err)
	}

	err = yaml.Unmarshal(buffer, &frequencyPlan)
	if err != nil {
		return frequencyPlan, errors.NewWithCause("Failed to parse the HTTP content as a frequency plan", err)
	}

	return frequencyPlan, nil
}
