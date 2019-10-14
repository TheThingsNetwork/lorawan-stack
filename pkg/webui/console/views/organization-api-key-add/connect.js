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

import { getOrganizationsRightsList } from '../../store/actions/organizations'
import {
  selectSelectedOrganizationId,
  selectOrganizationRights,
  selectOrganizationPseudoRights,
  selectOrganizationRightsError,
  selectOrganizationRightsFetching,
} from '../../store/selectors/organizations'

import api from '../../api'

export default OrganizationApiKeyAdd =>
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
      navigateToList: orgId => dispatch(replace(`/organizations/${orgId}/api-keys`)),
      createOrganizationApiKey: api.organization.apiKeys.create,
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      getOrganizationsRightsList: () => dispatchProps.getOrganizationsRightsList(stateProps.orgId),
      navigateToList: () => dispatchProps.navigateToList(stateProps.orgId),
      createOrganizationApiKey: key =>
        dispatchProps.createOrganizationApiKey(stateProps.orgId, key),
    }),
  )(
    withRequest(
      ({ getOrganizationsRightsList }) => getOrganizationsRightsList(),
      ({ fetching, rights }) => fetching || !Boolean(rights.length),
    )(OrganizationApiKeyAdd),
  )
