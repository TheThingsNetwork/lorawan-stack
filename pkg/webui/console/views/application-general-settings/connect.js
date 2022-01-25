// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { connect as withConnect } from 'react-redux'
import { replace } from 'connected-react-router'
import { bindActionCreators } from 'redux'

import withRequest from '@ttn-lw/lib/components/with-request'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import pipe from '@ttn-lw/lib/pipe'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import {
  checkFromState,
  mayEditBasicApplicationInfo,
  mayDeleteApplication,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayPurgeEntities,
  mayViewApplicationLink,
} from '@console/lib/feature-checks'

import { updateApplication, deleteApplication } from '@console/store/actions/applications'
import { updateApplicationLink, getApplicationLink } from '@console/store/actions/link'
import { getCollaboratorsList } from '@console/store/actions/collaborators'
import { getApiKeysList } from '@console/store/actions/api-keys'
import { getPubsubsList } from '@console/store/actions/pubsubs'
import { getWebhooksList } from '@console/store/actions/webhooks'

import {
  selectWebhooksTotalCount,
  selectWebhooksFetching,
  selectWebhooksError,
} from '@console/store/selectors/webhooks'
import {
  selectPubsubsTotalCount,
  selectPubsubsFetching,
  selectPubsubsError,
} from '@console/store/selectors/pubsubs'
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
import {
  selectApplicationLink,
  selectSelectedApplication,
  selectSelectedApplicationId,
} from '@console/store/selectors/applications'

const mapStateToProps = state => {
  const mayViewApiKeys = checkFromState(mayViewOrEditApplicationApiKeys, state)
  const mayViewCollaborators = checkFromState(mayViewOrEditApplicationCollaborators, state)
  const mayPurgeApp = checkFromState(mayPurgeEntities, state)
  const mayDeleteApp = checkFromState(mayDeleteApplication, state)
  const mayViewLink = checkFromState(mayViewApplicationLink, state)
  const applicationLink = selectApplicationLink(state)
  const apiKeysCount = selectApiKeysTotalCount(state)
  const collaboratorsCount = selectCollaboratorsTotalCount(state)
  const webhooksCount = selectWebhooksTotalCount(state)
  const pubsubsCount = selectPubsubsTotalCount(state)

  const entitiesFetching =
    selectApiKeysFetching(state) ||
    selectCollaboratorsFetching(state) ||
    selectPubsubsFetching(state) ||
    selectWebhooksFetching(state)
  const error =
    selectApiKeysError(state) ||
    selectCollaboratorsError(state) ||
    selectPubsubsError(state) ||
    selectWebhooksError(state)

  const fetching =
    entitiesFetching ||
    (mayViewApiKeys && typeof apiKeysCount === undefined) ||
    (mayViewCollaborators && collaboratorsCount === undefined) ||
    typeof collaboratorsCount === undefined ||
    typeof pubsubsCount === undefined
  const hasIntegrations = webhooksCount > 0 || pubsubsCount > 0
  const hasApiKeys = apiKeysCount > 0
  // Note: there is always at least one collaborator.
  const hasAddedCollaborators = collaboratorsCount > 1
  const isPristine = !hasApiKeys && !hasAddedCollaborators && !hasIntegrations
  return {
    appId: selectSelectedApplicationId(state),
    application: selectSelectedApplication(state),
    mayViewApiKeys,
    mayViewCollaborators,
    mayViewLink,
    fetching,
    mayPurge: mayPurgeApp,
    shouldConfirmDelete: !isPristine || !mayViewCollaborators || !mayViewApiKeys || Boolean(error),
    mayDeleteApplication: mayDeleteApp,
    link: applicationLink,
  }
}

const mapDispatchToProps = dispatch => ({
  ...bindActionCreators(
    {
      updateApplicationLink: attachPromise(updateApplicationLink),
      updateApplication: attachPromise(updateApplication),
      deleteApplication: attachPromise(deleteApplication),
      getApiKeysList,
      getCollaboratorsList,
      getWebhooksList,
      getPubsubsList,
      getApplicationLink,
    },
    dispatch,
  ),
  getLink: (id, selector) => dispatch(getApplicationLink(id, selector)),
  onDeleteSuccess: () => dispatch(replace(`/applications`)),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  deleteApplication: (id, purge = false) => dispatchProps.deleteApplication(id, { purge }),
  loadData: () => {
    if (stateProps.mayDeleteApplication) {
      if (stateProps.mayViewApiKeys) {
        dispatchProps.getApiKeysList('application', stateProps.appId)
      }

      if (stateProps.mayViewLink) {
        dispatchProps.getApplicationLink(stateProps.appId, 'skip_payload_crypto')
      }

      if (stateProps.mayViewCollaborators) {
        dispatchProps.getCollaboratorsList('application', stateProps.appId)
      }

      dispatchProps.getWebhooksList(stateProps.appId)
      dispatchProps.getPubsubsList(stateProps.appId)
    }
  },
})

const addHocs = pipe(
  withConnect(mapStateToProps, mapDispatchToProps, mergeProps),
  withFeatureRequirement(mayEditBasicApplicationInfo, {
    redirect: ({ appId }) => `/applications/${appId}`,
  }),
  withRequest(
    ({ loadData }) => loadData(),
    ({ fetching }) => fetching,
  ),
)

export default ApplicationGeneralSettings => addHocs(ApplicationGeneralSettings)
