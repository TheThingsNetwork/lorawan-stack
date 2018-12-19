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

package picture

import (
	"bytes"
	"image"
	"io"

	"github.com/disintegration/imaging"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func settings(format string) (encodingFormat imaging.Format, mimeType, extension string) {
	switch format {
	case "jpeg":
		return imaging.JPEG, "image/jpeg", "jpg"
	default:
		return imaging.PNG, "image/png", "png"
	}
}

// MakeSquare makes a square version of the image read from r, with the given
// maximum dimensions.
func MakeSquare(r io.Reader, dimensions int) (*ttnpb.Picture, error) {
	img, format, err := image.Decode(r)
	if err != nil {
		return nil, err
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
	encodedFormat, mimeType, _ := settings(format)
	var buf bytes.Buffer
	if err = imaging.Encode(&buf, img, encodedFormat); err != nil {
		return nil, err
	}
	return &ttnpb.Picture{
		Embedded: &ttnpb.Picture_Embedded{
			MimeType: mimeType,
			Data:     buf.Bytes(),
		},
	}, nil
}
