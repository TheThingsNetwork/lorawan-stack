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

import { listModels } from '@console/store/actions/device-repository'

import {
  selectDeviceModelsByBrandId,
  selectDeviceModelsError,
  selectDeviceModelsFetching,
} from '@console/store/selectors/device-repository'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

const mapStateToProps = (state, props) => {
  const { brandId } = props

  return {
    appId: selectSelectedApplicationId(state),
    models: selectDeviceModelsByBrandId(state, brandId),
    error: selectDeviceModelsError(state),
    fetching: selectDeviceModelsFetching(state),
  }
}

const mapDispatchToProps = { listModels }

export default ModelSelect => connect(mapStateToProps, mapDispatchToProps)(ModelSelect)
