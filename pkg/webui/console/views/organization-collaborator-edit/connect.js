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

import withRequest from '../../../lib/components/with-request'

import { getOrganizationCollaborator } from '../../store/actions/organizations'
import {
  selectSelectedOrganizationId,
  selectOrganizationRights,
  selectOrganizationPseudoRights,
  selectOrganizationRightsFetching,
  selectOrganizationRightsError,
  selectOrganizationUserCollaborator,
  selectOrganizationOrganizationCollaborator,
  selectOrganizationCollaboratorFetching,
  selectOrganizationCollaboratorError,
} from '../../store/selectors/organizations'

import api from '../../api'

export default OrganizationCollaboratorEdit =>
  connect(
    (state, props) => {
      const orgId = selectSelectedOrganizationId(state, props)

      const { collaboratorId, collaboratorType } = props.match.params

      const collaborator =
        collaboratorType === 'user'
          ? selectOrganizationUserCollaborator(state)
          : selectOrganizationOrganizationCollaborator(state)

      const fetching =
        selectOrganizationRightsFetching(state) || selectOrganizationCollaboratorFetching(state)
      const error =
        selectOrganizationRightsError(state) || selectOrganizationCollaboratorError(state)

      return {
        collaboratorId,
        collaboratorType,
        collaborator,
        orgId,
        rights: selectOrganizationRights(state),
        pseudoRights: selectOrganizationPseudoRights(state),
        fetching,
        error,
      }
    },
    dispatch => ({
      getOrganizationCollaborator(orgId, collaboratorId, isUser) {
        dispatch(getOrganizationCollaborator(orgId, collaboratorId, isUser))
      },
      redirectToList(orgId) {
        dispatch(replace(`/organizations/${orgId}/collaborators`))
      },
      updateOrganizationCollaborator: (orgId, collaborator) =>
        api.organization.collaborators.update(orgId, collaborator),
      removeOrganizationCollaborator: (orgId, collaborator) =>
        api.organization.collaborators.remove(orgId, collaborator),
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      getOrganizationCollaborator: () =>
        dispatchProps.getOrganizationCollaborator(
          stateProps.orgId,
          stateProps.collaboratorId,
          stateProps.collaboratorType === 'user',
        ),
      redirectToList: () => dispatchProps.redirectToList(stateProps.orgId),
      updateOrganizationCollaborator: collaborator =>
        dispatchProps.updateOrganizationCollaborator(stateProps.orgId, collaborator),
      removeOrganizationCollaborator: collaborator =>
        dispatchProps.removeOrganizationCollaborator(stateProps.orgId, collaborator),
    }),
  )(
    withRequest(
      ({ getOrganizationCollaborator }) => getOrganizationCollaborator(),
      ({ fetching, collaborator }) => fetching || !Boolean(collaborator),
    )(OrganizationCollaboratorEdit),
  )
