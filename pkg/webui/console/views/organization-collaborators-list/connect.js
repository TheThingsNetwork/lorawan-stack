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

import { getOrganizationCollaboratorsList } from '../../store/actions/organizations'
import {
  selectSelectedOrganizationId,
  selectOrganizationCollaborators,
  selectOrganizationCollaboratorsTotalCount,
  selectOrganizationCollaboratorsFetching,
} from '../../store/selectors/organizations'

export default OrganizationCollaboratorsList =>
  connect(
    state => ({
      orgId: selectSelectedOrganizationId(state),
    }),
    dispatch => ({
      getOrganizationCollaboratorsList: (id, filters) =>
        dispatch(getOrganizationCollaboratorsList(id, filters)),
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      getOrganizationCollaboratorsList: filters =>
        dispatchProps.getOrganizationCollaboratorsList(stateProps.orgId, filters),
      selectTableData: state => {
        const id = { id: stateProps.orgId }

        return {
          collaborators: selectOrganizationCollaborators(state, id),
          totalCount: selectOrganizationCollaboratorsTotalCount(state, id),
          fetching: selectOrganizationCollaboratorsFetching(state),
        }
      },
    }),
  )(OrganizationCollaboratorsList)
