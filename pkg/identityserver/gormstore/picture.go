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
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Picture model.
type Picture struct {
	Model
	SoftDelete // Filter on deleted_at not being NULL to clean up storage bucket.

	Data []byte `gorm:"type:BYTEA"`
}

func init() {
	registerModel(&Picture{})
}

func (p Picture) toPB() *ttnpb.Picture {
	pb := &ttnpb.Picture{}
	proto.Unmarshal(p.Data, pb)
	return pb
}

func (p *Picture) fromPB(pb *ttnpb.Picture) {
	p.Data, _ = proto.Marshal(pb)
}
