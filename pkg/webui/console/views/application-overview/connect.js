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

import { connect } from 'react-redux'

import {
  selectSelectedApplication,
  selectSelectedApplicationId,
  selectApplicationCollaboratorsTotalCount,
  selectApplicationCollaboratorsFetching,
  selectApplicationApiKeysTotalCount,
  selectApplicationApiKeysFetching,
  selectSelectedApplicationDevicesTotalCount,
  selectApplicationLinkIndicator,
  selectSelectedApplicationDevicesFetching,
  selectApplicationLinkFetching,
} from '../../store/selectors/applications'
import {
  getApplicationCollaboratorsList,
  getApplicationApiKeysList,
} from '../../store/actions/applications'
import { getApplicationLink } from '../../store/actions/link'

const mapStateToProps = (state, props) => {
  const appId = selectSelectedApplicationId(state)

  return {
    appId,
    application: selectSelectedApplication(state),
    collaboratorsTotalCount: selectApplicationCollaboratorsTotalCount(state, { id: appId }),
    apiKeysTotalCount: selectApplicationApiKeysTotalCount(state, { id: appId }),
    devicesTotalCount: selectSelectedApplicationDevicesTotalCount(state),
    link: selectApplicationLinkIndicator(state),
    statusBarFetching:
      selectApplicationLinkFetching(state) ||
      selectSelectedApplicationDevicesFetching(state) ||
      selectApplicationApiKeysFetching(state) ||
      selectApplicationCollaboratorsFetching(state),
  }
}

const mapDispatchToProps = dispatch => ({
  loadData(appId) {
    dispatch(getApplicationCollaboratorsList(appId))
    dispatch(getApplicationApiKeysList(appId))
    dispatch(getApplicationLink(appId))
  },
})

export default ApplicationOverview =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
  )(ApplicationOverview)
