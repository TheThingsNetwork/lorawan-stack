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

package picture

import (
	"bytes"
	"context"
	"image"
	"io"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gocloud.dev/blob"
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

var errMissingOriginal = errors.DefineNotFound("original_not_found", "original picture not found")

func openOriginal(ctx context.Context, bucket *blob.Bucket, prefix string, pic *ttnpb.Picture) (image.Image, string, error) {
	var originalData io.Reader
	if pic.Embedded != nil {
		originalData = bytes.NewBuffer(pic.Embedded.Data)
	} else {
		original := pic.Sizes[0]
		if original == "" {
			var maxSize uint32
			for size := range pic.Sizes {
				if size > maxSize {
					maxSize = size
				}
			}
			original = pic.Sizes[maxSize]
		}
		if original == "" {
			return nil, "", errMissingOriginal
		}
		if strings.Contains(original, "://") {
			res, err := http.Get(original)
			if err != nil {
				return nil, "", err
			}
			originalData = res.Body
			defer res.Body.Close()
		} else {
			r, err := bucket.NewReader(ctx, path.Join(prefix, original), nil)
			if err != nil {
				return nil, "", err
			}
			originalData = r
			defer r.Close()
		}
	}
	if originalData == nil {
		return nil, "", errMissingOriginal
	}
	return image.Decode(originalData)
}

func store(ctx context.Context, bucket *blob.Bucket, key string, img image.Image, format imaging.Format, mimeType string) (err error) {
	w, err := bucket.NewWriter(ctx, key, &blob.WriterOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return err
	}
	defer func() {
		closeErr := w.Close()
		if err == nil {
			err = closeErr
		}
	}()
	err = imaging.Encode(w, img, format)
	if err != nil {
		return err
	}
	return nil
}

// Store the picture in the bucket, under the given prefix. If the picture is larger
// than the given sizes, this will also generate thumbnails and store them as well.
func Store(ctx context.Context, bucket *blob.Bucket, prefix string, pic *ttnpb.Picture, sizes ...int) (*ttnpb.Picture, error) {
	original, format, err := openOriginal(ctx, bucket, prefix, pic)
	if err != nil {
		return nil, err
	}
	applicableSizes := make([]int, 0, len(sizes)+1)
	for _, size := range sizes {
		if size > original.Bounds().Dy() {
			continue
		}
		applicableSizes = append(applicableSizes, size)
	}
	applicableSizes = append(applicableSizes, 0)
	sort.Sort(sort.Reverse(sort.IntSlice(applicableSizes)))
	imagesBySize := make(map[uint32]string, len(applicableSizes))
	img := original
	encodedFormat, mimeType, extension := settings(format)
	for _, size := range applicableSizes {
		if err = ctx.Err(); err != nil {
			return nil, err // Early exit if context canceled.
		}
		var name string
		if size == 0 {
			name = "original"
			img = original
		} else {
			name = strconv.Itoa(size)
			img = imaging.Resize(img, size, size, imaging.Lanczos)
		}
		key := path.Join(prefix, name+"."+extension)
		err = store(ctx, bucket, key, img, encodedFormat, mimeType)
		if err != nil {
			return nil, err
		}
		imagesBySize[uint32(size)] = key
	}
	return &ttnpb.Picture{
		Sizes: imagesBySize,
	}, nil
}
