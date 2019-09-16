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

package picture_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/blob"
	"go.thethings.network/lorawan-stack/pkg/identityserver/picture"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func makeCheckers(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, image.Rect(0, 0, width/2, height/2), image.NewUniform(color.Black), image.ZP, draw.Src)
	draw.Draw(img, image.Rect(0, height/2, width/2, height), image.NewUniform(color.White), image.ZP, draw.Src)
	draw.Draw(img, image.Rect(width/2, 0, width, height/2), image.NewUniform(color.White), image.ZP, draw.Src)
	draw.Draw(img, image.Rect(width/2, height/2, width, height), image.NewUniform(color.Black), image.ZP, draw.Src)
	return img
}

func TestMakeSquare(t *testing.T) {
	for _, tt := range []struct {
		Name           string
		Image          image.Image
		ExpectedBounds int
	}{
		{Name: "Large Horizontal", Image: makeCheckers(800, 600), ExpectedBounds: 500},
		{Name: "Small Horizontal", Image: makeCheckers(400, 300), ExpectedBounds: 300},
		{Name: "Large Vertical", Image: makeCheckers(600, 800), ExpectedBounds: 500},
		{Name: "Small Vertical", Image: makeCheckers(300, 400), ExpectedBounds: 300},
		{Name: "Large Square", Image: makeCheckers(800, 800), ExpectedBounds: 500},
		{Name: "Small Square", Image: makeCheckers(400, 400), ExpectedBounds: 400},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)

			var b bytes.Buffer
			png.Encode(&b, tt.Image)

			pic, err := picture.MakeSquare(&b, 500)

			a.So(err, should.BeNil)
			if a.So(pic, should.NotBeNil) && a.So(pic.Embedded, should.NotBeNil) {
				a.So(pic.Embedded.MimeType, should.Equal, "image/png")

				img, _, err := image.Decode(bytes.NewBuffer(pic.Embedded.Data))

				a.So(err, should.BeNil)
				if a.So(img, should.NotBeNil) {
					a.So(img.Bounds(), should.Resemble, image.Rect(0, 0, tt.ExpectedBounds, tt.ExpectedBounds))
					a.So(img.At(tt.ExpectedBounds/2-10, tt.ExpectedBounds/2-10), beSameColorAs, color.Black)
					a.So(img.At(tt.ExpectedBounds/2+10, tt.ExpectedBounds/2-10), beSameColorAs, color.White)
					a.So(img.At(tt.ExpectedBounds/2-10, tt.ExpectedBounds/2+10), beSameColorAs, color.White)
					a.So(img.At(tt.ExpectedBounds/2+10, tt.ExpectedBounds/2+10), beSameColorAs, color.Black)
				}
			}
		})
	}
}

func beSameColorAs(actual interface{}, expected ...interface{}) (message string) {
	actualR, actualG, actualB, actualA := actual.(color.Color).RGBA()
	expectedR, expectedG, expectedB, expectedA := expected[0].(color.Color).RGBA()
	if eq := should.Equal(actualR, expectedR); eq != "" {
		return fmt.Sprintf("R does not equal expected:\n%s", eq)
	}
	if eq := should.Equal(actualG, expectedG); eq != "" {
		return fmt.Sprintf("G does not equal expected:\n%s", eq)
	}
	if eq := should.Equal(actualB, expectedB); eq != "" {
		return fmt.Sprintf("B does not equal expected:\n%s", eq)
	}
	if eq := should.Equal(actualA, expectedA); eq != "" {
		return fmt.Sprintf("A does not equal expected:\n%s", eq)
	}
	return ""
}

func TestStore(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	var blobConfig blob.Config
	blobConfig.Provider, blobConfig.Local.Directory = "local", "."
	bucket, _ := blobConfig.GetBucket(ctx, "testdata")

	pic, err := picture.Store(ctx, bucket, "picture", &ttnpb.Picture{
		Sizes: map[uint32]string{
			800: "source.png",
		},
	}, 800, 400)

	a.So(err, should.BeNil)
	if a.So(pic, should.NotBeNil) && a.So(pic.Sizes, should.ContainKey, uint32(400)) {
		a.So(pic.Sizes[400], should.Equal, "picture/400.png")
	}
}
