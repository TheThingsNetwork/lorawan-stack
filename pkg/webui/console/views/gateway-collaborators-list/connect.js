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

import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'
import {
  selectCollaborators,
  selectCollaboratorsTotalCount,
  selectCollaboratorsFetching,
  selectCollaboratorsError,
} from '@ttn-lw/lib/store/selectors/collaborators'

import { selectSelectedGatewayId } from '@console/store/selectors/gateways'

const mapStateToProps = state => ({
  gtwId: selectSelectedGatewayId(state),
  getCollaboratorsList: (id, filters) => getCollaboratorsList('gateway', id, filters),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  getCollaboratorsList: filters => stateProps.getCollaboratorsList(stateProps.gtwId, filters),
  selectTableData: state => {
    const id = { id: stateProps.orgId }

    return {
      collaborators: selectCollaborators(state, id),
      totalCount: selectCollaboratorsTotalCount(state),
      fetching: selectCollaboratorsFetching(state),
      error: selectCollaboratorsError(state),
    }
  },
})

export default GatewayCollaboratorsList =>
  connect(mapStateToProps, null, mergeProps)(GatewayCollaboratorsList)
