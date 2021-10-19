// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"encoding/json"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
	"golang.org/x/text/language"
)

func TestBundle(t *testing.T) {
	a := assertions.New(t)

	bundle, err := i18n.NewBundle(language.English)
	a.So(err, should.BeNil)

	msg, err := bundle.Define("hello", "hello, @{username}!")
	a.So(err, should.BeNil)
	a.So(msg, should.NotBeNil)
	a.So(msg.ID(), should.Equal, "hello")
	a.So(msg.String(), should.Equal, "hello, @{username}!")
	a.So(msg.Arguments(), should.Resemble, []string{"username"})

	a.So(bundle.GetAllIDs(), should.Contain, "hello")

	_, err = bundle.Define("hello", "hello message")
	a.So(err, should.NotBeNil)

	translations, err := bundle.MessagesFor(language.Dutch, true)
	a.So(err, should.BeNil)

	translationData := []byte(`{"hello":"hallo, @{username}!"}`)
	err = json.Unmarshal(translationData, translations)
	a.So(err, should.BeNil)

	formatted, err := bundle.Get("hello", bundle.MatchLanguage(language.Dutch, language.English)).Format(map[string]interface{}{
		"username": "htdvisser",
	})
	a.So(err, should.BeNil)
	a.So(formatted, should.Equal, "hallo, @htdvisser!")

	marshaled, err := json.Marshal(translations)
	a.So(err, should.BeNil)
	a.So(marshaled, should.Resemble, translationData)
}
