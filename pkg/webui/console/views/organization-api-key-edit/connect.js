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
import { replace } from 'connected-react-router'

import api from '@console/api'

import withRequest from '@ttn-lw/lib/components/with-request'

import { getApiKey } from '@console/store/actions/api-keys'

import {
  selectSelectedOrganizationId,
  selectOrganizationRights,
  selectOrganizationPseudoRights,
  selectOrganizationRightsError,
  selectOrganizationRightsFetching,
} from '@console/store/selectors/organizations'
import {
  selectSelectedApiKey,
  selectApiKeyError,
  selectApiKeyFetching,
} from '@console/store/selectors/api-keys'

export default OrganizationApiKeyEdit =>
  connect(
    function (state, props) {
      const { apiKeyId } = props.match.params

      const keyFetching = selectApiKeyFetching(state)
      const rightsFetching = selectOrganizationRightsFetching(state)
      const keyError = selectApiKeyError(state)
      const rightsError = selectOrganizationRightsError(state)

      return {
        keyId: apiKeyId,
        orgId: selectSelectedOrganizationId(state),
        apiKey: selectSelectedApiKey(state),
        rights: selectOrganizationRights(state),
        pseudoRights: selectOrganizationPseudoRights(state),
        fetching: keyFetching || rightsFetching,
        error: keyError || rightsError,
      }
    },
    dispatch => ({
      getApiKey(orgId, keyId) {
        dispatch(getApiKey('organization', orgId, keyId))
      },
      deleteOrganizationApiKeySuccess: orgId =>
        dispatch(replace(`/organizations/${orgId}/api-keys`)),
      deleteOrganizationApiKey: api.organization.apiKeys.delete,
      updateOrganizationApiKey: api.organization.apiKeys.update,
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      deleteOrganizationApiKeySuccess: () =>
        dispatchProps.deleteOrganizationApiKeySuccess(stateProps.orgId),
      deleteOrganizationApiKey: () =>
        dispatchProps.deleteOrganizationApiKey(stateProps.orgId, stateProps.keyId),
      updateOrganizationApiKey: key =>
        dispatchProps.updateOrganizationApiKey(stateProps.orgId, stateProps.keyId, key),
    }),
  )(
    withRequest(
      ({ getApiKey, orgId, keyId }) => getApiKey(orgId, keyId),
      ({ fetching, apiKey }) => fetching || !Boolean(apiKey),
    )(OrganizationApiKeyEdit),
  )
