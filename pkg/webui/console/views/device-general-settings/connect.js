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

import { replace } from 'connected-react-router'
import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'

import tts from '@console/api/tts'

import {
  selectIsConfig,
  selectAsConfig,
  selectJsConfig,
  selectNsConfig,
} from '@ttn-lw/lib/selectors/env'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  mayEditApplicationDeviceKeys,
  mayReadApplicationDeviceKeys,
} from '@console/lib/feature-checks'

import { updateDevice, resetDevice, resetUsedDevNonces } from '@console/store/actions/devices'
import { unclaimDevice } from '@console/store/actions/claim'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import {
  selectSelectedDevice,
  selectSelectedDeviceId,
  selectSelectedDeviceClaimable,
} from '@console/store/selectors/devices'

const mapStateToProps = state => ({
  supportsClaiming: selectSelectedDeviceClaimable(state),
  device: selectSelectedDevice(state),
  devId: selectSelectedDeviceId(state),
  appId: selectSelectedApplicationId(state),
  isConfig: selectIsConfig(),
  asConfig: selectAsConfig(),
  jsConfig: selectJsConfig(),
  nsConfig: selectNsConfig(),
  mayReadKeys: mayReadApplicationDeviceKeys.check(
    mayReadApplicationDeviceKeys.rightsSelector(state),
  ),
  mayEditKeys: mayEditApplicationDeviceKeys.check(
    mayEditApplicationDeviceKeys.rightsSelector(state),
  ),
  getDefaultMacSettings: (freqPlan, phyVersion) =>
    tts.Ns.getDefaultMacSettings(freqPlan, phyVersion),
})
const mapDispatchToProps = dispatch => ({
  ...bindActionCreators(
    {
      updateDevice: attachPromise(updateDevice),
      resetDevice: attachPromise(resetDevice),
      resetUsedDevNonces: attachPromise(resetUsedDevNonces),
      unclaimDevice: attachPromise(unclaimDevice),
    },
    dispatch,
  ),
  onDeleteSuccess: appId => dispatch(replace(`/applications/${appId}/devices`)),
  onDelete: tts.Applications.Devices.deleteById,
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  onDeleteSuccess: () => dispatchProps.onDeleteSuccess(stateProps.appId),
  resetDevice: () => dispatchProps.resetDevice(stateProps.appId, stateProps.devId),
  resetUsedDevNonces: () => dispatchProps.resetUsedDevNonces(stateProps.appId, stateProps.devId),
})

export default DeviceGeneralSettings =>
  connect(mapStateToProps, mapDispatchToProps, mergeProps)(DeviceGeneralSettings)
