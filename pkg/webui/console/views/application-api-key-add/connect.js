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

import { getApplicationsRightsList } from '@console/store/actions/applications'

import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationPseudoRights,
  selectApplicationRightsError,
  selectApplicationRightsFetching,
} from '@console/store/selectors/applications'

export default ApplicationApiKeyAdd =>
  connect(
    state => ({
      appId: selectSelectedApplicationId(state),
      fetching: selectApplicationRightsFetching(state),
      error: selectApplicationRightsError(state),
      rights: selectApplicationRights(state),
      pseudoRights: selectApplicationPseudoRights(state),
    }),
    dispatch => ({
      createApiKey: api.application.apiKeys.create,
      getApplicationsRightsList: appId => dispatch(getApplicationsRightsList(appId)),
      navigateToList: appId => dispatch(replace(`/applications/${appId}/api-keys`)),
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      createApplicationApiKey: key => dispatchProps.createApiKey(stateProps.appId, key),
      navigateToList: () => dispatchProps.navigateToList(stateProps.appId),
      getApplicationsRightsList: () => dispatchProps.getApplicationsRightsList(stateProps.appId),
    }),
  )(ApplicationApiKeyAdd)
