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

package config

import (
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestUnmarshal(t *testing.T) {
	a := assertions.New(t)

	mgr := Initialize("test", "test", defaults)
	a.So(mgr, should.NotBeNil)

	mgr.Parse()
	err := mgr.mergeConfig(strings.NewReader(`file-only: 10`))
	a.So(err, should.BeNil)

	var res map[string]interface{}
	err = mgr.Unmarshal(&res)
	a.So(err, should.BeNil)
	a.So(res, should.ContainKey, "file-only")
	a.So(res["file-only"], should.Resemble, 10)
}

func TestUnmarshalKey(t *testing.T) {
	a := assertions.New(t)

	mgr := Initialize("test", "test", defaults)
	a.So(mgr, should.NotBeNil)

	mgr.Parse()
	err := mgr.mergeConfig(strings.NewReader(`file-only: 10`))
	a.So(err, should.BeNil)

	var res int
	err = mgr.UnmarshalKey("file-only", &res)
	a.So(err, should.BeNil)
	a.So(res, should.Resemble, 10)
}
