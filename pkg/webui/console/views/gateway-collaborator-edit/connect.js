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

import { getCollaborator, getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import {
  selectUserCollaborator,
  selectOrganizationCollaborator,
  selectCollaboratorFetching,
  selectCollaboratorError,
  selectCollaboratorsTotalCount,
} from '@ttn-lw/lib/store/selectors/collaborators'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsFetching,
  selectGatewayRightsError,
} from '@console/store/selectors/gateways'

const mapStateToProps = (state, props) => {
  const gtwId = selectSelectedGatewayId(state, props)

  const { collaboratorId, collaboratorType } = props.match.params

  const collaborator =
    collaboratorType === 'user'
      ? selectUserCollaborator(state)
      : selectOrganizationCollaborator(state)

  const fetching = selectGatewayRightsFetching(state) || selectCollaboratorFetching(state)
  const error = selectGatewayRightsError(state) || selectCollaboratorError(state)

  return {
    collaboratorId,
    collaboratorType,
    collaborator,
    collaboratorsTotalCount: selectCollaboratorsTotalCount(state, gtwId),
    gtwId,
    rights: selectGatewayRights(state),
    pseudoRights: selectGatewayPseudoRights(state),
    fetching,
    error,
  }
}

const mapDispatchToProps = dispatch => ({
  getCollaborator: (gtwId, collaboratorId, isUser) => {
    dispatch(getCollaborator('gateway', gtwId, collaboratorId, isUser))
    dispatch(getCollaboratorsList('gateway', gtwId))
  },
  redirectToList: gtwId => {
    dispatch(replace(`/gateways/${gtwId}/collaborators`))
  },
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  getGatewayCollaborator: () =>
    dispatchProps.getCollaborator(
      stateProps.gtwId,
      stateProps.collaboratorId,
      stateProps.collaboratorType === 'user',
    ),
  redirectToList: () => dispatchProps.redirectToList(stateProps.gtwId),
  updateCollaborator: patch => tts.Gateways.Collaborators.update(stateProps.gtwId, patch),
  removeCollaborator: collaboratorIds =>
    tts.Gateways.Collaborators.remove(stateProps.gtwId, collaboratorIds),
})

export default GatewayCollaboratorEdit =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(withRequest(({ getGatewayCollaborator }) => getGatewayCollaborator())(GatewayCollaboratorEdit))
