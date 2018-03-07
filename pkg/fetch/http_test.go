// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch_test

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/fetch"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestHTTP(t *testing.T) {
	a := assertions.New(t)

	// Invalid path
	{
		fetcher := fetch.FromHTTP("", false)
		_, err := fetcher.File("test")
		a.So(err, should.NotBeNil)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	serverHost := "http://server"
	content := "server content"

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/success", serverHost), httpmock.NewStringResponder(200, content))
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/fail", serverHost), httpmock.NewStringResponder(500, ""))

	fetcher := fetch.FromHTTP(serverHost, false)

	// Valid response code
	{
		receivedContent, err := fetcher.File("success")
		a.So(err, should.BeNil)
		a.So(string(receivedContent), should.Equal, content)
	}

	// Internal error response code
	{
		_, err := fetcher.File("fail")
		a.So(err, should.NotBeNil)
	}
}
