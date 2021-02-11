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

import { push } from 'connected-react-router'
import { connect } from 'react-redux'

import api from '@console/api'

import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationPseudoRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '@console/store/selectors/applications'
import { selectCollaborators } from '@console/store/selectors/collaborators'

const mapStateToProps = state => ({
  appId: selectSelectedApplicationId(state),
  collaborators: selectCollaborators(state),
  rights: selectApplicationRights(state),
  pseudoRights: selectApplicationPseudoRights(state),
  fetching: selectApplicationRightsFetching(state),
  error: selectApplicationRightsError(state),
})

const mapDispatchToProps = dispatch => ({
  redirectToList: appId => dispatch(push(`/applications/${appId}/collaborators`)),
  addCollaborator: (appId, collaborator) => api.application.collaborators.add(appId, collaborator),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  redirectToList: () => dispatchProps.redirectToList(stateProps.appId),
  addCollaborator: collaborator => dispatchProps.addCollaborator(stateProps.appId, collaborator),
})

export default ApplicationCollaboratorAdd =>
  connect(mapStateToProps, mapDispatchToProps, mergeProps)(ApplicationCollaboratorAdd)
