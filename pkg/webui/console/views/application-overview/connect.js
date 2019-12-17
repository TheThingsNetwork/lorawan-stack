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
  selectApplicationLinkIndicator,
  selectApplicationLinkFetching,
} from '../../store/selectors/applications'
import { selectDevicesTotalCount, selectDevicesFetching } from '../../store/selectors/devices'
import {
  getApplicationCollaboratorsList,
  getApplicationApiKeysList,
} from '../../store/actions/applications'
import { getApplicationLink } from '../../store/actions/link'

import {
  checkFromState,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayViewApplicationDevices,
  mayLinkApplication,
} from '../../lib/feature-checks'

const mapStateToProps = (state, props) => {
  const appId = selectSelectedApplicationId(state)
  const collaboratorsTotalCount = selectApplicationCollaboratorsTotalCount(state, { id: appId })
  const apiKeysTotalCount = selectApplicationApiKeysTotalCount(state, { id: appId })
  const devicesTotalCount = selectDevicesTotalCount(state)
  const mayViewApplicationApiKeys = checkFromState(mayViewOrEditApplicationApiKeys, state)
  const mayViewApplicationCollaborators = checkFromState(
    mayViewOrEditApplicationCollaborators,
    state,
  )
  const mayViewApplicationLink = checkFromState(mayLinkApplication, state)
  const mayViewDevices = checkFromState(mayViewApplicationDevices, state)
  const collaboratorsFetching =
    (mayViewApplicationCollaborators && collaboratorsTotalCount === undefined) ||
    selectApplicationCollaboratorsFetching(state)
  const apiKeysFetching =
    (mayViewApplicationApiKeys && apiKeysTotalCount === undefined) ||
    selectApplicationApiKeysFetching(state)
  const devicesFetching =
    (mayViewDevices && devicesTotalCount === undefined) || selectDevicesFetching(state)

  return {
    appId,
    application: selectSelectedApplication(state),
    collaboratorsTotalCount,
    apiKeysTotalCount,
    devicesTotalCount,
    mayViewApplicationApiKeys,
    mayViewApplicationCollaborators,
    mayViewApplicationLink,
    mayViewDevices,
    link: selectApplicationLinkIndicator(state),
    statusBarFetching:
      collaboratorsFetching ||
      apiKeysFetching ||
      devicesFetching ||
      selectApplicationLinkFetching(state),
  }
}

const mapDispatchToProps = dispatch => ({
  loadData(
    mayViewApplicationCollaborators,
    mayViewApplicationApiKeys,
    mayViewApplicationLink,
    appId,
  ) {
    if (mayViewApplicationCollaborators) dispatch(getApplicationCollaboratorsList(appId))
    if (mayViewApplicationApiKeys) dispatch(getApplicationApiKeysList(appId))
    if (mayViewApplicationLink) dispatch(getApplicationLink(appId))
  },
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  loadData: () =>
    dispatchProps.loadData(
      stateProps.mayViewApplicationCollaborators,
      stateProps.mayViewApplicationApiKeys,
      stateProps.mayViewApplicationLink,
      stateProps.appId,
    ),
})

export default ApplicationOverview =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(ApplicationOverview)
