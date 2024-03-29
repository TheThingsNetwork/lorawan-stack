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

import autoBind from 'auto-bind'

import Marshaler from '../util/marshaler'

class Ns {
  constructor(service) {
    this._api = service
    autoBind(this)
  }

  async generateDevAddress() {
    const result = await this._api.GenerateDevAddr()

    return Marshaler.payloadSingleResponse(result)
  }

  async getDefaultMacSettings(freqPlan, phyVersion) {
    const result = await this._api.GetDefaultMACSettings({
      routeParams: {
        frequency_plan_id: freqPlan,
        lorawan_phy_version: phyVersion,
      },
    })

    return Marshaler.payloadSingleResponse(result)
  }
}

export default Ns
