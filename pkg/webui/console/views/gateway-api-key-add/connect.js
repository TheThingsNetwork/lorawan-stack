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
import { replace } from 'connected-react-router'

import api from '@console/api'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayRightsError,
  selectGatewayRightsFetching,
  selectGatewayPseudoRights,
} from '@console/store/selectors/gateways'

export default GatewayApiKeyAdd =>
  connect(
    state => ({
      gtwId: selectSelectedGatewayId(state),
      fetching: selectGatewayRightsFetching(state),
      error: selectGatewayRightsError(state),
      rights: selectGatewayRights(state),
      pseudoRights: selectGatewayPseudoRights(state),
    }),
    dispatch => ({
      createApiKey: api.gateway.apiKeys.create,
      navigateToList: gtwId => dispatch(replace(`/gateways/${gtwId}/api-keys`)),
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      createGatewayApiKey: key => dispatchProps.createApiKey(stateProps.gtwId, key),
      navigateToList: () => dispatchProps.navigateToList(stateProps.gtwId),
    }),
  )(GatewayApiKeyAdd)
