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
  checkFromState,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayViewApplicationDevices,
  mayLinkApplication,
} from '@console/lib/feature-checks'

import { getCollaboratorsList } from '@console/store/actions/collaborators'
import { getApiKeysList } from '@console/store/actions/api-keys'
import { getApplicationLink } from '@console/store/actions/link'
import { getApplicationDeviceCount } from '@console/store/actions/applications'

import { selectApiKeysTotalCount, selectApiKeysFetching } from '@console/store/selectors/api-keys'
import {
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
} from '@console/store/selectors/collaborators'
import {
  selectSelectedApplication,
  selectSelectedApplicationId,
  selectApplicationLinkIndicator,
  selectApplicationLinkFetching,
  selectApplicationDeviceCount,
  selectApplicationLastSeen,
} from '@console/store/selectors/applications'

const mapStateToProps = (state, props) => {
  const appId = selectSelectedApplicationId(state)
  const collaboratorsTotalCount = selectCollaboratorsTotalCount(state, { id: appId })
  const apiKeysTotalCount = selectApiKeysTotalCount(state)
  const devicesTotalCount = selectApplicationDeviceCount(state)
  const mayViewApplicationApiKeys = checkFromState(mayViewOrEditApplicationApiKeys, state)
  const mayViewApplicationCollaborators = checkFromState(
    mayViewOrEditApplicationCollaborators,
    state,
  )
  const mayViewApplicationLink = checkFromState(mayLinkApplication, state)
  const mayViewDevices = checkFromState(mayViewApplicationDevices, state)
  const collaboratorsFetching =
    (mayViewApplicationCollaborators && collaboratorsTotalCount === undefined) ||
    selectCollaboratorsFetching(state)
  const apiKeysFetching =
    (mayViewApplicationApiKeys && apiKeysTotalCount === undefined) || selectApiKeysFetching(state)
  const devicesFetching = mayViewDevices && devicesTotalCount === undefined

  return {
    appId,
    application: selectSelectedApplication(state),
    applicationLastSeen: selectApplicationLastSeen(state),
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
    mayViewDevices,
    appId,
  ) {
    if (mayViewApplicationCollaborators) dispatch(getCollaboratorsList('application', appId))
    if (mayViewApplicationApiKeys) dispatch(getApiKeysList('application', appId))
    if (mayViewApplicationLink) dispatch(getApplicationLink(appId))
    if (mayViewDevices) dispatch(getApplicationDeviceCount(appId))
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
      stateProps.mayViewDevices,
      stateProps.appId,
    ),
})

export default ApplicationOverview =>
  connect(mapStateToProps, mapDispatchToProps, mergeProps)(ApplicationOverview)
