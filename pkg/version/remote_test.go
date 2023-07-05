// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package version_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

func TestCheckUpdate(t *testing.T) {
	_, ctx := test.New(t)
	const timeout = time.Second

	srv := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer srv.Close()
	sourceOpt := version.WithURLs(fmt.Sprintf("%s/tts.json", srv.URL), "https://example.com")

	for _, tc := range []struct {
		Reference    semver.Version
		Minor, Patch bool
	}{
		{
			Reference: semver.MustParse("3.14.0"), // Unreleased version
		},
		{
			Reference: semver.MustParse("3.13.2"), // Latest version
		},
		{
			Reference: semver.MustParse("3.13.1"), // Previous patch
			Patch:     true,
		},
		{
			Reference: semver.MustParse("3.12.3"), // Previous patch
			Minor:     true,
		},
	} {
		t.Run(tc.Reference.String(), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			version.ClearRecentCheckCache()
			update, err := version.CheckUpdate(ctx, sourceOpt, version.WithReference(tc.Reference))
			a.So(err, should.BeNil)

			if !tc.Minor && !tc.Patch {
				a.So(update, should.BeNil)
			} else {
				a.So(update, should.NotBeNil)
				a.So(update.Minor, should.Equal, tc.Minor)
				a.So(update.Patch, should.Equal, tc.Patch)
			}
		})
	}
}
