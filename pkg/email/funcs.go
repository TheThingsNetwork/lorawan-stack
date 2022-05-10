// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package email

import (
	"fmt"
	"html/template"
	"math"
	"path"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
)

const documentationBaseURL = "https://www.thethingsindustries.com/docs"

var defaultFuncs = template.FuncMap{
	"documentation_url": func(elems ...string) string {
		p := path.Join(elems...)
		return documentationBaseURL + "/" + strings.TrimPrefix(p, "/")
	},
	"relTime":  relTime,
	"enumDesc": enumDesc,
}

const (
	day   = 24 * time.Hour
	week  = 7 * day
	month = 30 * day
	year  = 12 * month
)

func relTime(d time.Duration) string {
	now := time.Now()
	return humanize.CustomRelTime(now.Add(d), now, "ago", "from now", []humanize.RelTimeMagnitude{
		{D: time.Second, Format: "now", DivBy: time.Second},
		{D: 2 * time.Second, Format: "a second %s", DivBy: 1},

		{D: time.Minute, Format: "%d seconds %s", DivBy: time.Second},
		{D: 2 * time.Minute, Format: "a minute %s", DivBy: 1},

		{D: time.Hour, Format: "%d minutes %s", DivBy: time.Minute},
		{D: 2 * time.Hour, Format: "an hour %s", DivBy: 1},

		{D: day, Format: "%d hours %s", DivBy: time.Hour},
		{D: 2 * day, Format: "a day %s", DivBy: 1},

		{D: week, Format: "%d days %s", DivBy: day},
		{D: 2 * week, Format: "a week %s", DivBy: 1},

		{D: month, Format: "%d weeks %s", DivBy: week},
		{D: 2 * month, Format: "a month %s", DivBy: 1},

		{D: year, Format: "%d months %s", DivBy: month},
		{D: 2 * year, Format: "a year %s", DivBy: 1},

		{D: math.MaxInt64, Format: "%d years %s", DivBy: year},
	})
}

func enumDesc(enum fmt.Stringer) string {
	md := i18n.Get(fmt.Sprintf("enum:%s", enum))
	if md == nil {
		return ""
	}
	return md.String()
}
