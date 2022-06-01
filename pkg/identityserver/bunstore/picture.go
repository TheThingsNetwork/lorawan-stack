// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Picture is the picture model in the database.
type Picture struct {
	bun.BaseModel `bun:"table:pictures,alias:pic"`

	Model
	SoftDelete

	Data []byte `bun:"data,type:bytea"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Picture) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func pictureToPB(m *Picture) (*ttnpb.Picture, error) {
	pb := &ttnpb.Picture{}
	if err := proto.Unmarshal(m.Data, pb); err != nil {
		return nil, err
	}
	return pb, nil
}

func pictureFromPB(_ context.Context, pb *ttnpb.Picture) (*Picture, error) {
	m := &Picture{}
	data, err := proto.Marshal(pb)
	if err != nil {
		return nil, err
	}
	m.Data = data
	return m, nil
}
