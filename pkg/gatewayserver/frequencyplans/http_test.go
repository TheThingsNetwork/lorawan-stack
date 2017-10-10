// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/frequencyplans"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
	yaml "gopkg.in/yaml.v2"
)

var (
	euFPID       = "EU_FREQUENCY_PLAN"
	euFPFilename = "eu_frequency_plan.yml"

	euWithLBTID       = "EU_FREQUENCY_PLAN_WITH_LBT"
	euWithLBTFilename = "eu_frequency_plan_lbt.yml"
)

func dummyFPList() string {
	return fmt.Sprintf(`- id: %s
  file: %s`, euFPID, euFPFilename)
}

func dummyFPListWithOnlyLBTExtension() string {
	return fmt.Sprintf(`- id: %s
  file: %s
  base: %s`, euWithLBTID, euWithLBTFilename, euFPID)
}

func dummyFPListWithLBTExtension() string {
	return dummyFPList() + `
` + dummyFPListWithOnlyLBTExtension()
}

func dummyEUFP() (string, error) {
	fp := ttnpb.FrequencyPlan{
		BandID: string(band.EU_863_870),
		Channels: []*ttnpb.FrequencyPlan_Channel{
			&ttnpb.FrequencyPlan_Channel{
				Frequency: 868000000,
			},
		},
	}
	out, err := yaml.Marshal(fp)
	return string(out), err
}

func dummyLBTExtension() (string, error) {
	ext := ttnpb.FrequencyPlan{
		LBT: &ttnpb.FrequencyPlan_LBTConfiguration{
			RSSITarget: 80,
		},
	}
	out, err := yaml.Marshal(ext)
	return string(out), err
}

func listURL() string {
	return gitHubFileURL("frequency-plans.yml")
}

func gitHubFileURL(filename string) string {
	return fmt.Sprintf("%s/%s", string(frequencyplans.DefaultBaseURL), filename)
}

func Example() {
	store, err := frequencyplans.RetrieveHTTPStore()
	if err != nil {
		panic(err)
	}

	frequencyPlansIDs := store.GetAllIDs()

	for _, frequencyPlanID := range frequencyPlansIDs {
		frequencyPlan, err := store.GetByID(frequencyPlanID)
		if err != nil {
			panic(err)
		}

		fmt.Println("Number of channels in frequency plan", frequencyPlanID, ": ", len(frequencyPlan.Channels))
	}
}

func ExampleBaseURIOption() {
	// A local HTTP server exposes the frequency plans list on /frequency-plans.yml, and exposes every frequency plan yml file on /<filename>
	store, err := frequencyplans.RetrieveHTTPStore(frequencyplans.BaseURIOption("http://localhost"))
	if err != nil {
		panic(err)
	}

	frequencyPlansIDs := store.GetAllIDs()

	for _, frequencyPlanID := range frequencyPlansIDs {
		frequencyPlan, err := store.GetByID(frequencyPlanID)
		if err != nil {
			panic(err)
		}

		fmt.Println("Number of channels in frequency plan", frequencyPlanID, ": ", len(frequencyPlan.Channels))
	}
}

func ExampleFileSystemRootPathOption() {
	// If all the frequency plans and the list of frequency plans are located in /workdir/frequencyplans:
	store, err := frequencyplans.ReadFileSystemStore(frequencyplans.FileSystemRootPathOption("/workdir/frequencyplans"))
	if err != nil {
		panic(err)
	}

	frequencyPlansIDs := store.GetAllIDs()

	for _, frequencyPlanID := range frequencyPlansIDs {
		frequencyPlan, err := store.GetByID(frequencyPlanID)
		if err != nil {
			panic(err)
		}

		fmt.Println("Number of channels in frequency plan", frequencyPlanID, ": ", len(frequencyPlan.Channels))
	}
}

func TestRetrieveHTTPStoreWithoutFrequencyPlans(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	frequencyPlansListURL := listURL()
	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, dummyFPList()))

	_, err := frequencyplans.RetrieveHTTPStore()
	a.So(err, should.NotBeNil)
}

func testUnknownFrequencyPlan(a *assertions.Assertion, store frequencyplans.Store) {
	_, err := store.GetByID("Unregistered Frequency Plan")
	a.So(err, should.NotBeNil)
}

func testKnownFrequencyPlan(a *assertions.Assertion, store frequencyplans.Store) {
	_, err := store.GetByID(euFPID)
	a.So(err, should.BeNil)
}

func TestRetrieveHTTPStore(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dummyFP, err := dummyEUFP()
	a.So(err, should.BeNil)

	frequencyPlansListURL := listURL()
	europeanFPURL := gitHubFileURL(euFPFilename)
	europeanLBTExtensionURL := gitHubFileURL(euWithLBTFilename)

	lbtExtension, err := dummyLBTExtension()
	a.So(err, should.BeNil)

	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, dummyFPListWithLBTExtension()))
	httpmock.RegisterResponder("GET", europeanLBTExtensionURL, httpmock.NewStringResponder(200, lbtExtension))
	httpmock.RegisterResponder("GET", europeanFPURL, httpmock.NewStringResponder(200, dummyFP))

	store, err := frequencyplans.RetrieveHTTPStore()

	a.So(err, should.BeNil)

	testUnknownFrequencyPlan(a, store)
	testKnownFrequencyPlan(a, store)

	ids := store.GetAllIDs()
	a.So(len(ids), should.Equal, 2)

	fp, err := store.GetByID(euFPID)
	a.So(err, should.BeNil)
	a.So(fp.LBT, should.BeNil)

	extended, err := store.GetByID(euWithLBTID)
	a.So(err, should.BeNil)
	a.So(extended.LBT, should.NotBeNil)
}

func TestHTTPStoreExtensionContentNotFound(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dummyFP, err := dummyEUFP()
	a.So(err, should.BeNil)

	frequencyPlansListURL := listURL()
	europeanFPURL := gitHubFileURL(euFPFilename)

	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, dummyFPListWithLBTExtension()))
	httpmock.RegisterResponder("GET", europeanFPURL, httpmock.NewStringResponder(200, dummyFP))

	_, err = frequencyplans.RetrieveHTTPStore()

	a.So(err, should.NotBeNil)
}

func TestHTTPStoreExtensionBaseNotFound(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	frequencyPlansListURL := listURL()
	europeanLBTExtensionURL := gitHubFileURL(euWithLBTFilename)

	lbtExtension, err := dummyLBTExtension()
	a.So(err, should.BeNil)

	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, dummyFPListWithOnlyLBTExtension()))
	httpmock.RegisterResponder("GET", europeanLBTExtensionURL, httpmock.NewStringResponder(200, lbtExtension))

	_, err = frequencyplans.RetrieveHTTPStore()

	a.So(err, should.NotBeNil)
}

func TestRetrieveWithOptions(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	dummyFP, err := dummyEUFP()
	a.So(err, should.BeNil)

	dummyURI := "http://dummy.server"
	frequencyPlansListURL := fmt.Sprintf("%s/%s", dummyURI, "frequency-plans.yml")
	europeanFPURL := fmt.Sprintf("%s/%s", dummyURI, euFPFilename)

	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, dummyFPList()))
	httpmock.RegisterResponder("GET", europeanFPURL, httpmock.NewStringResponder(200, dummyFP))

	store, err := frequencyplans.RetrieveHTTPStore(frequencyplans.BaseURIOption(dummyURI))
	a.So(err, should.BeNil)

	ids := store.GetAllIDs()
	a.So(len(ids), should.Equal, 1)
	a.So(ids[0], should.Equal, euFPID)
}

func TestInvalidFrequencyPlansList(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	list := `invalid: test
		test: test`
	frequencyPlansListURL := listURL()
	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, list))

	_, err := frequencyplans.RetrieveHTTPStore()
	a.So(err, should.NotBeNil)
}

func TestInvalidFrequencyPlan(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	frequencyPlansListURL := listURL()
	europeanFPURL := gitHubFileURL(euFPFilename)

	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(200, dummyFPList()))
	httpmock.RegisterResponder("GET", europeanFPURL, httpmock.NewStringResponder(200, `    dummy`))

	_, err := frequencyplans.RetrieveHTTPStore()
	a.So(err, should.NotBeNil)
}

func TestInvalidServerResponse(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	frequencyPlansListURL := listURL()
	httpmock.RegisterResponder("GET", frequencyPlansListURL, httpmock.NewStringResponder(500, "Internal Server Error"))

	_, err := frequencyplans.RetrieveHTTPStore()
	a.So(err, should.NotBeNil)
}

type failingReadCloser struct{}

func (f failingReadCloser) Read([]byte) (int, error) { return 0, errors.New("Failing") }
func (f failingReadCloser) Close() error             { return nil }

func TestInvalidReader(t *testing.T) {
	a := assertions.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	frequencyPlansListURL := listURL()

	responder := httpmock.ResponderFromResponse(&http.Response{
		Body:       failingReadCloser{},
		StatusCode: 200,
	})

	httpmock.RegisterResponder("GET", frequencyPlansListURL, responder)

	_, err := frequencyplans.RetrieveHTTPStore()
	a.So(err, should.NotBeNil)
}
