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

package gatewayconfigurationserver

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

func (s *Server) handleGetFrequencyPlan(w http.ResponseWriter, r *http.Request) {
	frequencyPlanID := mux.Vars(r)["frequency_plan_id"]
	plan, err := s.component.FrequencyPlans.GetByID(frequencyPlanID)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	config, err := shared.BuildSX1301Config(plan)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	if r.Header.Get("User-Agent") == "TTNGateway" {
		// Filter out fields to reduce response size.
		config.TxLUTConfigs = nil
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		SX1301Conf *shared.SX1301Config `json:"SX1301_conf"`
	}{
		SX1301Conf: config,
	})
}
