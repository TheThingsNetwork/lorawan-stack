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

package store

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Picture model.
type Picture struct {
	Model
	SoftDelete // Filter on deleted_at not being NULL to clean up storage bucket.

	OriginalLocation string `gorm:"type:VARCHAR"`

	Original []byte `gorm:"type:BYTEA"`
	MIMEType string `gorm:"type:VARCHAR"`

	ResizedAt *time.Time
	Sizes     pq.Int64Array `gorm:"type:INT ARRAY"`
	Extension string        `gorm:"type:VARCHAR"`
}

func init() {
	registerModel(&Picture{})
}

func (p Picture) toPB() *ttnpb.Picture {
	pb := &ttnpb.Picture{
		Sizes: map[uint32]string{},
	}
	if p.OriginalLocation != "" {
		pb.Sizes[0] = p.OriginalLocation
	} else if p.ResizedAt == nil && len(p.Original) > 0 {
		pb.Embedded = &ttnpb.Picture_Embedded{
			MimeType: p.MIMEType,
			Data:     p.Original,
		}
	}
	for _, size := range p.Sizes {
		pb.Sizes[uint32(size)] = fmt.Sprintf("%s/%d.%s", p.ID, size, p.Extension)
	}
	return pb
}

func (p *Picture) fromPB(pb *ttnpb.Picture) {
	p.OriginalLocation = pb.Sizes[0]
	p.Sizes = make(pq.Int64Array, 0, len(pb.Sizes))
	for size := range pb.Sizes {
		if size != 0 {
			p.Sizes = append(p.Sizes, int64(size))
		}
	}
	if pb.Embedded != nil {
		p.Original = pb.Embedded.Data
		p.MIMEType = pb.Embedded.MimeType
	}
}
