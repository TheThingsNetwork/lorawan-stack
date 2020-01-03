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

import { getCollaboratorsList } from '../../store/actions/collaborators'
import { selectSelectedOrganizationId } from '../../store/selectors/organizations'
import {
  selectCollaborators,
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
} from '../../store/selectors/collaborators'

export default OrganizationCollaboratorsList =>
  connect(
    state => ({
      orgId: selectSelectedOrganizationId(state),
    }),
    dispatch => ({
      getCollaboratorsList: (id, filters) =>
        dispatch(getCollaboratorsList('organization', id, filters)),
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      getCollaboratorsList: filters =>
        dispatchProps.getCollaboratorsList(stateProps.orgId, filters),
      selectTableData: state => {
        const id = { id: stateProps.orgId }

        return {
          collaborators: selectCollaborators(state, id),
          totalCount: selectCollaboratorsTotalCount(state, id),
          fetching: selectCollaboratorsFetching(state),
        }
      },
    }),
  )(OrganizationCollaboratorsList)
