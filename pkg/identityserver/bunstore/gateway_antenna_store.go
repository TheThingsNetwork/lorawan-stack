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
	"sort"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GatewayAntenna is the gateway antenna model in the database.
type GatewayAntenna struct {
	bun.BaseModel `bun:"table:gateway_antennas,alias:ant"`

	Model

	GatewayID string   `bun:"gateway_id,notnull"`
	Gateway   *Gateway `bun:"rel:belongs-to,join:gateway_id=id"`

	Index int `bun:"index,notnull"`

	Gain float32 `bun:"gain,nullzero"`

	Location

	// TODO: Add attributes as JSONB field.
	// Attributes map[string]string `bun:"attributes"`

	Placement int `bun:"placement,nullzero"`
}

func gatewayAntennaFromPB(pb *ttnpb.GatewayAntenna) *GatewayAntenna {
	return &GatewayAntenna{
		Gain:     pb.Gain,
		Location: locationFromPB(pb.Location),
		// TODO: Attributes field.
		Placement: int(pb.Placement),
	}
}

func gatewayAntennaToPB(m *GatewayAntenna) *ttnpb.GatewayAntenna {
	if m == nil {
		return nil
	}
	pb := &ttnpb.GatewayAntenna{
		Gain:     m.Gain,
		Location: locationToPB(m.Location),
		// TODO: Attributes field.
		Placement: ttnpb.GatewayAntennaPlacement(m.Placement),
	}
	return pb
}

// GatewayAntennaSlice is a slice of GatewayAntenna.
type GatewayAntennaSlice []*GatewayAntenna

func (a GatewayAntennaSlice) Len() int      { return len(a) }
func (a GatewayAntennaSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a GatewayAntennaSlice) Less(i, j int) bool {
	return a[i].Index < a[j].Index
}

func gatewayAntennaSliceFromPB(pbs []*ttnpb.GatewayAntenna, gatewayID string) []*GatewayAntenna {
	out := make([]*GatewayAntenna, len(pbs))
	for i, pb := range pbs {
		out[i] = gatewayAntennaFromPB(pb)
		out[i].GatewayID = gatewayID
		out[i].Index = i
	}
	return out
}

func gatewayAntennaMap(models []*GatewayAntenna) map[int]*GatewayAntenna {
	m := make(map[int]*GatewayAntenna, len(models))
	for _, model := range models {
		m[model.Index] = model
	}
	return m
}

func (s *gatewayStore) replaceGatewayAntennas(
	ctx context.Context, current []*GatewayAntenna, desired []*ttnpb.GatewayAntenna, gatewayID string,
) ([]*GatewayAntenna, error) {
	var (
		oldMap   = gatewayAntennaMap(current)
		newMap   = gatewayAntennaMap(gatewayAntennaSliceFromPB(desired, gatewayID))
		toCreate = make([]*GatewayAntenna, 0, len(newMap))
		toUpdate = make([]*GatewayAntenna, 0, len(newMap))
		toDelete = make([]*GatewayAntenna, 0, len(oldMap))
		result   = make(GatewayAntennaSlice, 0, len(newMap))
	)

	for k, v := range newMap {
		// Ignore gateway antenna that has not been updated.
		if current, ok := oldMap[k]; ok {
			delete(oldMap, k) // Don't need to delete this one.
			delete(newMap, k) // Don't need to create this one.
			if current.Gain == v.Gain && current.Location == v.Location && current.Placement == v.Placement {
				result = append(result, v)
				continue // Don't need to update this one.
			}
			v.ID = current.ID
			toUpdate = append(toUpdate, v)
			result = append(result, v)
			continue
		}
		toCreate = append(toCreate, v)
		result = append(result, v)
	}
	for _, v := range oldMap {
		toDelete = append(toDelete, v)
	}

	if len(toDelete) > 0 {
		_, err := s.DB.NewDelete().
			Model(&toDelete).
			WherePK().
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	if len(toUpdate) > 0 {
		_, err := s.DB.NewUpdate().
			Model(&toUpdate).
			Column("gain", "latitude", "longitude", "altitude", "accuracy", "placement"). // TODO: Add attributes.
			Bulk().
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	if len(toCreate) > 0 {
		_, err := s.DB.NewInsert().
			Model(&toCreate).
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
	}

	sort.Sort(result)

	return result, nil
}
