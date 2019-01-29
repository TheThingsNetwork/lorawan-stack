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

package commands

import (
	"bytes"
	"image"
	"os"

	"github.com/disintegration/imaging"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const dimensions = 1024

func readPicture(filename string) (*ttnpb.Picture, error) {
	pictureFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer pictureFile.Close()
	img, format, err := image.Decode(pictureFile)
	if err != nil {
		return nil, err
	}
	var encodingFormat imaging.Format
	var mimeType string
	switch format {
	case "jpeg":
		encodingFormat, mimeType = imaging.JPEG, "image/jpeg"
	default:
		encodingFormat, mimeType = imaging.PNG, "image/png"
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	if width < height {
		img = imaging.CropAnchor(img, width, width, imaging.Center)
	} else if width > height {
		img = imaging.CropAnchor(img, height, height, imaging.Center)
	}
	if width > dimensions || height > dimensions {
		img = imaging.Resize(img, dimensions, dimensions, imaging.Lanczos)
	}
	var buf bytes.Buffer
	if err = imaging.Encode(&buf, img, encodingFormat); err != nil {
		return nil, err
	}
	return &ttnpb.Picture{
		Embedded: &ttnpb.Picture_Embedded{
			MimeType: mimeType,
			Data:     buf.Bytes(),
		},
	}, nil
}
