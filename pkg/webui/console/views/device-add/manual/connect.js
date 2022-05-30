// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { push } from 'connected-react-router'

import tts from '@console/api/tts'

import { selectNsConfig, selectAsConfig, selectJsConfig } from '@ttn-lw/lib/selectors/env'

import { checkFromState, mayEditApplicationDeviceKeys } from '@console/lib/feature-checks'

import { getApplicationDevEUICount, issueDevEUI } from '@console/store/actions/applications'

import {
  selectSelectedApplicationId,
  selectApplicationDevEUICount,
} from '@console/store/selectors/applications'
import { selectJoinEUIPrefixes } from '@console/store/selectors/join-server'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
  jsConfig: selectJsConfig(),
  nsConfig: selectNsConfig(),
  asConfig: selectAsConfig(),
  mayEditKeys: checkFromState(mayEditApplicationDeviceKeys, state),
  prefixes: selectJoinEUIPrefixes(state),
  createDevice: (appId, device) => tts.Applications.Devices.create(appId, device),
  applicationDevEUICounter: selectApplicationDevEUICount(state),
  getDefaultMacSettings: (freqPlan, phyVersion) =>
    tts.Ns.getDefaultMacSettings(freqPlan, phyVersion),
})

const mapDispatchToProps = dispatch => ({
  createDeviceSuccess: (appId, deviceId) =>
    dispatch(push(`/applications/${appId}/devices/${deviceId}`)),
  fetchDevEUICounter: appId => dispatch(getApplicationDevEUICount(appId)),
  issueDevEUI: appId => dispatch(attachPromise(issueDevEUI(appId))),
})

export default Manual => connect(mapStateToProps, mapDispatchToProps)(Manual)
