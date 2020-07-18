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

package store

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// Secrets model.
type Secrets struct {
	Model
	SoftDelete

	Data []byte `gorm:"type:BYTEA"`
}

func init() {
	registerModel(&Secrets{})
}

func (s Secrets) toPB() *ttnpb.Secrets {
	pb := &ttnpb.Secrets{}
	pb.Unmarshal(s.Data)
	return pb
}

func (s *Secrets) fromPB(pb *ttnpb.Secrets) {
	s.Data, _ = pb.Marshal()
}
