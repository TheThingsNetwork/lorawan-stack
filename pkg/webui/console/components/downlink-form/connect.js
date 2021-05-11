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

import { connect } from 'react-redux'

import api from '@console/api'

import {
  selectSelectedApplicationId,
  selectApplicationLinkSkipPayloadCrypto,
} from '@console/store/selectors/applications'
import { selectSelectedDeviceId, selectSelectedDevice } from '@console/store/selectors/devices'

const mapStateToProps = state => {
  const appId = selectSelectedApplicationId(state)
  const devId = selectSelectedDeviceId(state)
  const device = selectSelectedDevice(state)
  const skipPayloadCrypto = selectApplicationLinkSkipPayloadCrypto(state)

  return {
    appId,
    devId,
    device,
    downlinkQueue: api.downlinkQueue,
    skipPayloadCrypto,
  }
}

export default DownlinkForm => connect(mapStateToProps)(DownlinkForm)
