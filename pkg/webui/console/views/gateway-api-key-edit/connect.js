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
import { replace } from 'connected-react-router'

import tts from '@console/api/tts'

import withRequest from '@ttn-lw/lib/components/with-request'

import { getApiKey } from '@console/store/actions/api-keys'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsError,
  selectGatewayRightsFetching,
} from '@console/store/selectors/gateways'
import {
  selectSelectedApiKey,
  selectApiKeyError,
  selectApiKeyFetching,
} from '@console/store/selectors/api-keys'

export default GatewayApiKeyEdit =>
  connect(
    (state, props) => {
      const apiKeyId = props.match.params.apiKeyId
      const keyFetching = selectApiKeyFetching(state)
      const rightsFetching = selectGatewayRightsFetching(state)
      const keyError = selectApiKeyError(state)
      const rightsError = selectGatewayRightsError(state)

      return {
        keyId: apiKeyId,
        gtwId: selectSelectedGatewayId(state),
        apiKey: selectSelectedApiKey(state),
        rights: selectGatewayRights(state),
        pseudoRights: selectGatewayPseudoRights(state),
        fetching: keyFetching || rightsFetching,
        error: keyError || rightsError,
      }
    },
    dispatch => ({
      getApiKey: (gtwId, keyId) => {
        dispatch(getApiKey('gateway', gtwId, keyId))
      },
      deleteApiKey: tts.Gateways.ApiKeys.deleteById,
      deleteSuccess: gtwId => dispatch(replace(`/gateways/${gtwId}/api-keys`)),
      editApiKey: tts.Gateways.ApiKeys.updateById,
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      deleteGatewayApiKey: key => dispatchProps.deleteApiKey(stateProps.gtwId, key),
      editGatewayApiKey: key => dispatchProps.editApiKey(stateProps.gtwId, stateProps.keyId, key),
      deleteSuccess: () => dispatchProps.deleteSuccess(stateProps.gtwId),
    }),
  )(withRequest(({ gtwId, keyId, getApiKey }) => getApiKey(gtwId, keyId))(GatewayApiKeyEdit))
