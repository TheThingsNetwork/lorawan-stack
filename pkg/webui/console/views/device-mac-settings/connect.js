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

import { bindActionCreators } from 'redux'
import { connect } from 'react-redux'

import { updateDevice } from '../../store/actions/devices'
import { attachPromise } from '../../store/actions/lib'
import { selectSelectedApplicationId } from '../../store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '../../store/selectors/devices'

const mapStateToProps = state => ({
  devId: selectSelectedDeviceId(state),
  appId: selectSelectedApplicationId(state),
  device: selectSelectedDevice(state),
})
const mapDispatchToProps = dispatch => ({
  ...bindActionCreators({ updateDevice: attachPromise(updateDevice) }, dispatch),
})

export default DeviceMacSettings =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(DeviceMacSettings)
