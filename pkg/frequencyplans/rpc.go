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

package frequencyplans

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type RPCServer struct {
	store *Store
}

func NewRPCServer(store *Store) *RPCServer { return &RPCServer{store: store} }

func (s *RPCServer) ListFrequencyPlans(ctx context.Context, req *ttnpb.ListFrequencyPlansRequest) (*ttnpb.ListFrequencyPlansResponse, error) {
	descriptions, err := s.store.descriptions()
	if err != nil {
		return nil, err
	}
	res := &ttnpb.ListFrequencyPlansResponse{}
	for _, desc := range descriptions {
		if req.BaseFrequency != 0 && uint16(req.BaseFrequency) != desc.BaseFrequency {
			continue
		}
		res.FrequencyPlans = append(res.FrequencyPlans, &ttnpb.FrequencyPlanDescription{
			ID:            desc.ID,
			BaseID:        desc.BaseID,
			Name:          desc.Name,
			BaseFrequency: uint32(desc.BaseFrequency),
		})
	}
	return res, nil
}
