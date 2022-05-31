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
import { push } from 'connected-react-router'

import tts from '@console/api/tts'

import { selectSelectedCollaborator } from '@ttn-lw/lib/store/selectors/collaborators'

import { getGatewaysRightsList } from '@console/store/actions/gateways'

import {
  selectSelectedGatewayId,
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsError,
} from '@console/store/selectors/gateways'

const mapStateToProps = state => ({
  gtwId: selectSelectedGatewayId(state),
  collaborators: selectSelectedCollaborator(state),
  error: selectGatewayRightsError(state),
  rights: selectGatewayRights(state),
  pseudoRights: selectGatewayPseudoRights(state),
})

const mapDispatchToProps = dispatch => ({
  getGatewaysRightsList: gtwId => dispatch(getGatewaysRightsList(gtwId)),
  redirectToList: gtwId => dispatch(push(`/gateways/${gtwId}/collaborators`)),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  getGatewaysRightsList: () => dispatchProps.getGatewaysRightsList(stateProps.gtwId),
  redirectToList: () => dispatchProps.redirectToList(stateProps.gtwId),
  addCollaborator: collaborator => tts.Gateways.Collaborators.add(stateProps.gtwId, collaborator),
})

export default GatewayCollaboratorAdd =>
  connect(mapStateToProps, mapDispatchToProps, mergeProps)(GatewayCollaboratorAdd)
