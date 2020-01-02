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

import { getGatewayCollaboratorsList, getGatewayApiKeysList } from '../../store/actions/gateways'
import {
  selectSelectedGateway,
  selectSelectedGatewayId,
  selectGatewayCollaboratorsTotalCount,
  selectGatewayCollaboratorsFetching,
  selectGatewayApiKeysTotalCount,
  selectGatewayApiKeysFetching,
} from '../../store/selectors/gateways'

import {
  checkFromState,
  mayViewOrEditGatewayApiKeys,
  mayViewOrEditGatewayCollaborators,
} from '../../lib/feature-checks'

const mapStateToProps = state => {
  const gtwId = selectSelectedGatewayId(state)
  const collaboratorsTotalCount = selectGatewayCollaboratorsTotalCount(state, { id: gtwId })
  const apiKeysTotalCount = selectGatewayApiKeysTotalCount(state, { id: gtwId })
  const mayViewGatewayApiKeys = checkFromState(mayViewOrEditGatewayApiKeys, state)
  const mayViewGatewayCollaborators = checkFromState(mayViewOrEditGatewayCollaborators, state)
  const collaboratorsFetching =
    (mayViewGatewayCollaborators && collaboratorsTotalCount === undefined) ||
    selectGatewayCollaboratorsFetching(state)
  const apiKeysFetching =
    (mayViewGatewayApiKeys && apiKeysTotalCount === undefined) ||
    selectGatewayApiKeysFetching(state)

  return {
    gtwId,
    gateway: selectSelectedGateway(state),
    mayViewGatewayApiKeys,
    mayViewGatewayCollaborators,
    collaboratorsTotalCount,
    apiKeysTotalCount,
    statusBarFetching: collaboratorsFetching || apiKeysFetching,
  }
}
const mapDispatchToProps = dispatch => ({
  loadData(mayViewGatewayCollaborators, mayViewGatewayApiKeys, gtwId) {
    if (mayViewGatewayCollaborators) dispatch(getGatewayCollaboratorsList(gtwId))
    if (mayViewGatewayApiKeys) dispatch(getGatewayApiKeysList(gtwId))
  },
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  loadData: () =>
    dispatchProps.loadData(
      stateProps.mayViewGatewayCollaborators,
      stateProps.mayViewGatewayApiKeys,
      stateProps.gtwId,
    ),
})

export default GatewayOverview =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(GatewayOverview)
