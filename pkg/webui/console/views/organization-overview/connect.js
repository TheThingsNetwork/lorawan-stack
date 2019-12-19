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

import { getOrganizationCollaboratorsList } from '../../store/actions/organizations'
import { getApiKeysList } from '../../store/actions/api-keys'
import {
  selectSelectedOrganization,
  selectSelectedOrganizationId,
  selectOrganizationCollaboratorsTotalCount,
  selectOrganizationCollaboratorsFetching,
} from '../../store/selectors/organizations'
import { selectApiKeysTotalCount, selectApiKeysFetching } from '../../store/selectors/api-keys'

import {
  checkFromState,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
} from '../../lib/feature-checks'

const mapStateToProps = state => {
  const orgId = selectSelectedOrganizationId(state)
  const collaboratorsTotalCount = selectOrganizationCollaboratorsTotalCount(state, { id: orgId })
  const apiKeysTotalCount = selectApiKeysTotalCount(state, { id: orgId })
  const mayViewOrganizationApiKeys = checkFromState(mayViewOrEditOrganizationApiKeys, state)
  const mayViewOrganizationCollaborators = checkFromState(
    mayViewOrEditOrganizationCollaborators,
    state,
  )
  const collaboratorsFetching =
    (mayViewOrganizationCollaborators && collaboratorsTotalCount === undefined) ||
    selectOrganizationCollaboratorsFetching(state)
  const apiKeysFetching =
    (mayViewOrganizationApiKeys && apiKeysTotalCount === undefined) || selectApiKeysFetching(state)

  return {
    orgId,
    organization: selectSelectedOrganization(state),
    collaboratorsTotalCount,
    apiKeysTotalCount,
    mayViewOrganizationApiKeys,
    mayViewOrganizationCollaborators,
    statusBarFetching: apiKeysFetching || collaboratorsFetching,
  }
}

const mapDispatchToProps = dispatch => ({
  loadData(mayViewOrganizationCollaborators, mayViewOrganizationApiKeys, orgId) {
    if (mayViewOrganizationCollaborators) dispatch(getOrganizationCollaboratorsList(orgId))
    if (mayViewOrganizationApiKeys) dispatch(getApiKeysList('organization', orgId))
  },
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  loadData: () =>
    dispatchProps.loadData(
      stateProps.mayViewOrganizationCollaborators,
      stateProps.mayViewOrganizationApiKeys,
      stateProps.orgId,
    ),
})

export default Overview =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(Overview)
