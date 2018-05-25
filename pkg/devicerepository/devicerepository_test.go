// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package devicerepository_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type testFetcher map[string][]byte

func (t testFetcher) File(name ...string) ([]byte, error) {
	if content, ok := t[strings.Join(name, "/")]; ok {
		return content, nil
	}

	return nil, errors.New("Not found")
}

var (
	validFetcher, invalidFetcher, emptyFetcher testFetcher

	standardUnoEncoder = "function Encoder() { return { led: 1 }}"
)

func init() {
	// Valid fetcher.
	{
		validFetcher = map[string][]byte{}
		validFetcher["brands.yml"] = []byte(`version: '3'
brands:
  thethingsproducts:
    name: The Things Products
    url: https://www.thethingsnetwork.org
    logos:
    - logo.png`)
		validFetcher["thethingsproducts/logo.png"] = []byte("image")
		validFetcher["thethingsproducts/devices.yml"] = []byte(`version: '3'
devices:
  thethingsuno:
    name: The Things Uno`)
		validFetcher["thethingsproducts/thethingsuno/versions.yml"] = []byte(`version: '3'
hardware_versions:
  standard:
    firmware_versions: [v1.0]
    photos: [standard.png]
    payload_format:
      up:
        type: grpc
        param: hosted-service:1234
      down:
        type: javascript
        param: encoder.js`)
		validFetcher["thethingsproducts/thethingsuno/standard/standard.png"] = []byte("standard-image")
		validFetcher["thethingsproducts/thethingsuno/standard/encoder.js"] = []byte(standardUnoEncoder)
	}

	// Invalid fetcher.
	{
		invalidFetcher = map[string][]byte{}
		invalidFetcher["brands.yml"] = []byte(`version: '3'
brands:
  thethingsproducts:
  - name: The Things Products
  - url: https://www.thethingsnetwork.org`)
		invalidFetcher["thethingsproducts/devices.yml"] = []byte(`version: '3'
devices:
  thethingsuno:
		name: The Things Uno`)
		invalidFetcher["thethingsproducts/thethingsuno/versions.yml"] = []byte(`version: '3'
hardware_versions:
standard:
 		   firmware_versions: [v1.0]
    photos: [standard.png]
    payload_format:
      up:
        type: grpc
        param: hosted-service:1234
      down:
        type: javascript
        param: encoder.js`)
		invalidFetcher["thethingsproducts/thethingsnode/versions.yml"] = []byte(`version: '3'
hardware_versions:
  standard:
    firmware_versions: [v1.0]
    payload_format:
      up:
        type: grpc
        param: hosted-service:1234
      down:
        type: javascript
        param: encoder.js`)
	}

	// Empty fetcher.
	{
		emptyFetcher = map[string][]byte{}
	}
}

func TestBrand(t *testing.T) {
	a := assertions.New(t)

	// Valid fetcher.
	{
		c := devicerepository.Client{Fetcher: validFetcher}
		brands, err := c.Brands()
		a.So(err, should.BeNil)

		ttp, ok := brands["thethingsproducts"]
		a.So(ok, should.BeTrue)
		a.So(ttp.Name, should.Equal, "The Things Products")
		a.So(ttp.URL, should.Equal, "https://www.thethingsnetwork.org")
		a.So(ttp.Logos, should.HaveLength, 1)

		content, err := c.BrandLogo("thethingsproducts", ttp.Logos[0])
		a.So(err, should.BeNil)
		a.So(string(content.Content), should.Equal, "image")
	}

	// Invalid fetcher.
	{
		c := devicerepository.Client{Fetcher: invalidFetcher}
		_, err := c.Brands()
		a.So(err, should.NotBeNil)
	}

	// Invalid fetcher.
	{
		c := devicerepository.Client{Fetcher: emptyFetcher}
		_, err := c.Brands()
		a.So(err, should.NotBeNil)
	}
}

func TestDevice(t *testing.T) {
	a := assertions.New(t)

	// Valid fetcher.
	{
		c := devicerepository.Client{Fetcher: validFetcher}
		devices, err := c.Devices("thethingsproducts")
		a.So(err, should.BeNil)

		unoInfo, ok := devices["thethingsuno"]
		a.So(ok, should.BeTrue)
		a.So(unoInfo.Name, should.Equal, "The Things Uno")

		versions, err := c.DeviceVersions("thethingsproducts", "thethingsuno")
		a.So(err, should.BeNil)

		standard, ok := versions["standard"]
		a.So(ok, should.BeTrue)
		a.So(standard.FirmwareVersions[0], should.Equal, "v1.0")
		a.So(standard.Photos[0], should.Equal, "standard.png")
		a.So(standard.PayloadFormats.Up.Parameter, should.Equal, "hosted-service:1234")
		a.So(standard.PayloadFormats.Down.Parameter, should.Equal, standardUnoEncoder)
		a.So(standard.PayloadFormats.Up.Type, should.Equal, "grpc")
		a.So(standard.PayloadFormats.Down.Type, should.Equal, "javascript")

		image, err := c.DeviceVersionPhoto("thethingsproducts", "thethingsuno", "standard", "standard.png")
		a.So(err, should.BeNil)
		a.So(string(image.Content), should.Equal, "standard-image")
	}

	// Invalid fetcher.
	{
		c := devicerepository.Client{Fetcher: invalidFetcher}
		_, err := c.Devices("thethingsproducts")
		a.So(err, should.NotBeNil)

		_, err = c.DeviceVersions("thethingsproducts", "thethingsuno")
		a.So(err, should.NotBeNil)

		_, err = c.DeviceVersions("thethingsproducts", "thethingsnode")
		a.So(err, should.NotBeNil)
	}

	// Empty fetcher.
	{
		c := devicerepository.Client{Fetcher: emptyFetcher}
		_, err := c.Devices("thethingsproducts")
		a.So(err, should.NotBeNil)

		_, err = c.DeviceVersions("thethingsproducts", "thethingsuno")
		a.So(err, should.NotBeNil)
	}
}

func TestProtos(t *testing.T) {
	a := assertions.New(t)

	c := devicerepository.Client{Fetcher: validFetcher}

	// Brand.
	{
		brands, err := c.Brands()
		a.So(err, should.BeNil)

		ttp, ok := brands["thethingsproducts"]
		a.So(ok, should.BeTrue)

		proto := ttp.Proto()
		a.So(proto.ID, should.Equal, "thethingsproducts")
		a.So(proto.Name, should.Equal, ttp.Name)
	}

	// Model.
	{
		devices, err := c.Devices("thethingsproducts")
		a.So(err, should.BeNil)

		uno, ok := devices["thethingsuno"]
		a.So(ok, should.BeTrue)

		proto := uno.Proto()
		a.So(proto.BrandID, should.Equal, "thethingsproducts")
		a.So(proto.ModelID, should.Equal, "thethingsuno")
	}

	// Version.
	{
		versions, err := c.DeviceVersions("thethingsproducts", "thethingsuno")
		a.So(err, should.BeNil)

		std, ok := versions["standard"]
		a.So(ok, should.BeTrue)

		protos := std.Protos()
		a.So(protos, should.HaveLength, 1)
		proto := protos[0]
		a.So(proto.ModelID, should.Equal, "thethingsuno")
		a.So(proto.FirmwareVersion, should.Equal, "v1.0")
	}
}

func Example() {
	repository := devicerepository.Client{
		Fetcher: fetch.FromHTTP("https://raw.githubusercontent.com/TheThingsNetwork/devices/master", true),
	}

	brands, err := repository.Brands()
	if err != nil {
		panic(err)
	}

	for brandID, brand := range brands {
		fmt.Println("Brand:", brand.Name)
		devices, err := repository.Devices(brandID)
		if err != nil {
			panic(err)
		}

		for _, device := range devices {
			fmt.Println("\tDevice:", device.Name)
		}
	}
}
