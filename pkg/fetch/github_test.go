// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestURLRendering(t *testing.T) {
	a := assertions.New(t)

	fetcher := FromGitHubRepository("TheThingsNetwork/device", "master", "", false).(httpFetcher)
	a.So(fetcher.baseURL, should.Equal, "https://raw.githubusercontent.com/TheThingsNetwork/device/master")

	fetcher = FromGitHubRepository("TheThingsProducts/device", "develop", "src", false).(httpFetcher)
	a.So(fetcher.baseURL, should.Equal, "https://raw.githubusercontent.com/TheThingsProducts/device/develop/src")
}
