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

import withRequest from '@ttn-lw/lib/components/with-request'

import {
  checkFromState,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
} from '@console/lib/feature-checks'

import { getCollaboratorsList } from '@console/store/actions/collaborators'
import { getApiKeysList } from '@console/store/actions/api-keys'

import { selectOrganizationById } from '@console/store/selectors/organizations'
import {
  selectApiKeysTotalCount,
  selectApiKeysFetching,
  selectApiKeysError,
} from '@console/store/selectors/api-keys'
import {
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
  selectCollaboratorsError,
} from '@console/store/selectors/collaborators'

const mapStateToProps = (state, props) => {
  const apiKeysTotalCount = selectApiKeysTotalCount(state)
  const apiKeysFetching = selectApiKeysFetching(state)
  const apiKeysError = selectApiKeysError(state)
  const collaboratorsTotalCount = selectCollaboratorsTotalCount(state, { id: props.appId })
  const collaboratorsFetching = selectCollaboratorsFetching(state)
  const collaboratorsError = selectCollaboratorsError(state)

  const fetching = apiKeysFetching || collaboratorsFetching

  return {
    mayViewCollaborators: checkFromState(mayViewOrEditOrganizationCollaborators, state),
    mayViewApiKeys: checkFromState(mayViewOrEditOrganizationApiKeys, state),
    organization: selectOrganizationById(state, props.orgId),
    apiKeysTotalCount,
    apiKeysErrored: Boolean(apiKeysError),
    collaboratorsTotalCount,
    collaboratorsErrored: Boolean(collaboratorsError),
    fetching,
  }
}

const mapDispatchToProps = dispatch => ({
  loadData(mayViewCollaborators, mayViewApiKeys, orgId) {
    if (mayViewCollaborators) {
      dispatch(getCollaboratorsList('organization', orgId))
    }

    if (mayViewApiKeys) {
      dispatch(getApiKeysList('organization', orgId))
    }
  },
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  loadData: () =>
    dispatchProps.loadData(
      stateProps.mayViewCollaborators,
      stateProps.mayViewApiKeys,
      ownProps.orgId,
    ),
})

export default TitleSection =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(
    withRequest(
      ({ loadData }) => loadData(),
      () => false,
    )(TitleSection),
  )
