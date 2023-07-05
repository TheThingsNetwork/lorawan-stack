// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package i18n_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestI18N(t *testing.T) {
	a := assertions.New(t)
	fileName := filepath.Join(os.TempDir(), fmt.Sprintf("TestI18N_%d", time.Now().Unix()))

	m1 := make(i18n.MessageDescriptorMap)
	m1.Define("not_used", "not used")

	def := m1.Define("hello_world", "hello, world")
	a.So(def.Description.Package, should.Equal, "pkg/i18n")
	a.So(def.Description.File, should.Equal, "i18n_test.go")
	def.Translations["nl"] = "hallo, wereld"
	def.Translations["ja"] = "こんにちは世界"

	a.So(def.Load(), should.BeNil)

	def.Translations["unknown"] = "hello, world"

	for lang, translation := range def.Translations {
		actual := def.Format(lang, nil)
		a.So(actual, should.Equal, translation)
	}

	err := m1.WriteFile(fileName)
	a.So(err, should.BeNil)

	m2, err := i18n.ReadFile(fileName)
	a.So(err, should.BeNil)
	a.So(m2["hello_world"].Translations, should.Resemble, def.Translations)
	a.So(m2["hello_world"].Description, should.Resemble, def.Description)

	m3 := make(i18n.MessageDescriptorMap)
	m3.Define("hello_world", "hello, beautiful world")
	m3.Define("hello_you", "hello, you")

	m3.Merge(m2)

	a.So(m3.Cleanup(), should.Contain, "not_used")
	a.So(m3.Updated(), should.Contain, "hello_world")
}

func Example() {
	i18n.Define("welcome_message", "Welcome, {name}!")

	fmt.Println(i18n.Format("welcome_message", "en", map[string]any{"name": "Alice"}))

	// Output:
	// Welcome, Alice!
}
