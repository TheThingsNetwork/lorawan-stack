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

package gcsv2

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/shared"
)

func (s *Server) handleFreqPlanInfo(c echo.Context) error {
	freqPlanID := c.Param(frequencyPlanIDKey)
	plan, err := s.component.FrequencyPlans.GetByID(freqPlanID)
	if err != nil {
		return err
	}
	config, err := shared.BuildSX1301Config(plan)
	if err != nil {
		return err
	}
	config.TxLUTConfigs = config.TxLUTConfigs[:0]
	return c.JSON(http.StatusOK, struct {
		SX1301Conf *shared.SX1301Config `json:"SX1301_conf"`
	}{
		SX1301Conf: config,
	})
}
