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
import { replace } from 'connected-react-router'

import tts from '@console/api/tts'

import withRequest from '@ttn-lw/lib/components/with-request'

import { getApiKey } from '@console/store/actions/api-keys'

import {
  selectSelectedApplicationId,
  selectApplicationRights,
  selectApplicationPseudoRights,
  selectApplicationRightsError,
  selectApplicationRightsFetching,
} from '@console/store/selectors/applications'
import {
  selectSelectedApiKey,
  selectApiKeyError,
  selectApiKeyFetching,
} from '@console/store/selectors/api-keys'

export default ApplicationApiKeyEdit =>
  connect(
    (state, props) => {
      const { apiKeyId } = props.match.params

      const keyFetching = selectApiKeyFetching(state)
      const rightsFetching = selectApplicationRightsFetching(state)
      const keyError = selectApiKeyError(state)
      const rightsError = selectApplicationRightsError(state)

      return {
        keyId: apiKeyId,
        appId: selectSelectedApplicationId(state),
        apiKey: selectSelectedApiKey(state),
        rights: selectApplicationRights(state),
        pseudoRights: selectApplicationPseudoRights(state),
        fetching: keyFetching || rightsFetching,
        error: keyError || rightsError,
      }
    },
    dispatch => ({
      getApiKey: (appId, keyId) => {
        dispatch(getApiKey('application', appId, keyId))
      },
      deleteApiKey: tts.Applications.ApiKeys.deleteById,
      deleteSuccess: appId => dispatch(replace(`/applications/${appId}/api-keys`)),
      editApiKey: tts.Applications.ApiKeys.updateById,
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      deleteApplicationKey: key => dispatchProps.deleteApiKey(stateProps.appId, key),
      editApplicationKey: key => dispatchProps.editApiKey(stateProps.appId, stateProps.keyId, key),
      deleteSuccess: () => dispatchProps.deleteSuccess(stateProps.appId),
    }),
  )(withRequest(({ getApiKey, appId, keyId }) => getApiKey(appId, keyId))(ApplicationApiKeyEdit))
