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
import { bindActionCreators } from 'redux'
import { replace } from 'connected-react-router'

import withRequest from '@ttn-lw/lib/components/with-request'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  checkFromState,
  mayPurgeEntities,
  mayDeleteOrganization,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
} from '@console/lib/feature-checks'

import { updateOrganization, deleteOrganization } from '@console/store/actions/organizations'
import { getCollaboratorsList } from '@console/store/actions/collaborators'
import { getApiKeysList } from '@console/store/actions/api-keys'

import {
  selectSelectedOrganization,
  selectSelectedOrganizationId,
} from '@console/store/selectors/organizations'
import {
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
  selectCollaboratorsError,
} from '@console/store/selectors/collaborators'
import {
  selectApiKeysTotalCount,
  selectApiKeysFetching,
  selectApiKeysError,
} from '@console/store/selectors/api-keys'

const mapStateToProps = state => {
  const mayViewApiKeys = checkFromState(mayViewOrEditOrganizationApiKeys, state)
  const mayViewCollaborators = checkFromState(mayViewOrEditOrganizationCollaborators, state)
  const apiKeysCount = selectApiKeysTotalCount(state)
  const collaboratorsCount = selectCollaboratorsTotalCount(state)
  const mayPurgeOrg = checkFromState(mayPurgeEntities, state)
  const mayDeleteOrg = checkFromState(mayDeleteOrganization, state)

  const entitiesFetching = selectApiKeysFetching(state) || selectCollaboratorsFetching(state)
  const error = selectApiKeysError(state) || selectCollaboratorsError(state)

  const fetching =
    entitiesFetching ||
    (mayViewApiKeys && typeof apiKeysCount === undefined) ||
    (mayViewCollaborators && collaboratorsCount === undefined)
  const hasApiKeys = apiKeysCount > 0
  // Note: there is always at least one collaborator.
  const hasAddedCollaborators = collaboratorsCount > 1
  const isPristine = !hasApiKeys && !hasAddedCollaborators

  return {
    orgId: selectSelectedOrganizationId(state),
    organization: selectSelectedOrganization(state),
    mayViewApiKeys,
    mayViewCollaborators,
    fetching,
    shouldPurge: mayPurgeOrg,
    shouldConfirmDelete: !isPristine || !mayViewCollaborators || !mayViewApiKeys || Boolean(error),
    mayDeleteOrganization: mayDeleteOrg,
  }
}

const mapDispatchToProps = dispatch => ({
  ...bindActionCreators(
    {
      updateOrganization: attachPromise(updateOrganization),
      deleteOrganization: attachPromise(deleteOrganization),
      getApiKeysList,
      getCollaboratorsList,
    },
    dispatch,
  ),
  deleteOrganizationSuccess: () => dispatch(replace(`/organizations`)),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  deleteOrganization: id => dispatchProps.deleteOrganization(id, { purge: stateProps.shouldPurge }),
  loadData: () => {
    if (stateProps.mayDeleteOrganization) {
      if (stateProps.mayViewApiKeys) {
        dispatchProps.getApiKeysList('organization', stateProps.orgId)
      }

      if (stateProps.mayViewCollaborators) {
        dispatchProps.getCollaboratorsList('organization', stateProps.orgId)
      }
    }
  },
})

export default GeneralSettings =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(
    withRequest(
      ({ loadData }) => loadData(),
      ({ fetching }) => fetching,
    )(GeneralSettings),
  )
