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

import { getCollaborator } from '@console/store/actions/collaborators'

import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationPseudoRights,
  selectApplicationRightsFetching,
  selectApplicationRightsError,
} from '@console/store/selectors/applications'
import {
  selectUserCollaborator,
  selectOrganizationCollaborator,
  selectCollaboratorFetching,
  selectCollaboratorError,
} from '@console/store/selectors/collaborators'

const mapStateToProps = (state, props) => {
  const appId = selectSelectedApplicationId(state)

  const { collaboratorId, collaboratorType } = props.match.params

  const collaborator =
    collaboratorType === 'user'
      ? selectUserCollaborator(state)
      : selectOrganizationCollaborator(state)
  const fetching = selectApplicationRightsFetching(state) || selectCollaboratorFetching(state)
  const error = selectApplicationRightsError(state) || selectCollaboratorError(state)

  return {
    collaboratorId,
    collaboratorType,
    collaborator,
    appId,
    rights: selectApplicationRights(state),
    pseudoRights: selectApplicationPseudoRights(state),
    fetching,
    error,
  }
}

const mapDispatchToProps = (dispatch, ownProps) => ({
  getCollaborator(appId, collaboratorId, isUser) {
    dispatch(getCollaborator('application', appId, collaboratorId, isUser))
  },
  redirectToList(appId) {
    dispatch(replace(`/applications/${appId}/collaborators`))
  },
  updateCollaborator: (appId, patch) => api.application.collaborators.update(appId, patch),
  removeCollaborator: (appId, patch) => api.application.collaborators.update(appId, patch),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  getCollaborator: () =>
    dispatchProps.getCollaborator(
      stateProps.appId,
      stateProps.collaboratorId,
      stateProps.collaboratorType === 'user',
    ),
  redirectToList: () => dispatchProps.redirectToList(stateProps.appId),
  updateCollaborator: patch => api.application.collaborators.update(stateProps.appId, patch),
  removeCollaborator: patch => api.application.collaborators.update(stateProps.appId, patch),
})

export default ApplicationCollaboratorEdit =>
  connect(mapStateToProps, mapDispatchToProps, mergeProps)(ApplicationCollaboratorEdit)
