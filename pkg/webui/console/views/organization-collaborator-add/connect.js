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
import { push } from 'connected-react-router'

import {
  selectSelectedOrganizationId,
  selectOrganizationRights,
  selectOrganizationPseudoRights,
  selectOrganizationRightsFetching,
  selectOrganizationRightsError,
} from '../../store/selectors/organizations'
import { getOrganizationsRightsList } from '../../store/actions/organizations'

import api from '../../api'

export default OrganizationCollaboratorAdd =>
  connect(
    state => ({
      orgId: selectSelectedOrganizationId(state),
      fetching: selectOrganizationRightsFetching(state),
      error: selectOrganizationRightsError(state),
      rights: selectOrganizationRights(state),
      pseudoRights: selectOrganizationPseudoRights(state),
    }),
    dispatch => ({
      getOrganizationsRightsList: orgId => dispatch(getOrganizationsRightsList(orgId)),
      redirectToList: orgId => dispatch(push(`/organizations/${orgId}/collaborators`)),
      addOrganizationCollaborator: api.organization.collaborators.add,
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      getOrganizationsRightsList: () => dispatchProps.getOrganizationsRightsList(stateProps.orgId),
      redirectToList: () => dispatchProps.redirectToList(stateProps.orgId),
      addOrganizationCollaborator: collaborator =>
        dispatchProps.addOrganizationCollaborator(stateProps.orgId, collaborator),
    }),
  )(OrganizationCollaboratorAdd)
