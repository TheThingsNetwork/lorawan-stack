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

import withRequest from '@ttn-lw/lib/components/with-request'

import { getApiKey } from '@console/store/actions/api-keys'
import { getUsersRightsList } from '@console/store/actions/users'

import { selectUserId } from '@console/store/selectors/user'
import {
  selectUserRights,
  selectUserPseudoRights,
  selectUserRightsError,
  selectUserRightsFetching,
} from '@console/store/selectors/users'
import {
  selectSelectedApiKey,
  selectApiKeyError,
  selectApiKeyFetching,
} from '@console/store/selectors/api-keys'

const mapStateToProps = (state, props) => {
  const { apiKeyId } = props.match.params

  const keyFetching = selectApiKeyFetching(state)
  const rightsFetching = selectUserRightsFetching(state)
  const keyError = selectApiKeyError(state)
  const rightsError = selectUserRightsError(state)

  return {
    keyId: apiKeyId,
    userId: selectUserId(state),
    apiKey: selectSelectedApiKey(state),
    rights: selectUserRights(state),
    pseudoRights: selectUserPseudoRights(state),
    fetching: keyFetching || rightsFetching,
    error: keyError || rightsError,
  }
}

const mapDispatchToProps = dispatch => ({
  deleteApiKey: api.users.apiKeys.delete,
  updateApiKey: api.users.apiKeys.update,
  deleteUserApiKeySuccess: () => dispatch(replace(`/user/api-keys`)),
  loadData: (userId, keyId) => {
    dispatch(getApiKey('users', userId, keyId))
    dispatch(getUsersRightsList(userId))
  },
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  deleteUserApiKey: () => dispatchProps.deleteApiKey(stateProps.userId, stateProps.keyId),
  updateUserApiKey: key => dispatchProps.updateApiKey(stateProps.userId, stateProps.keyId, key),
  loadData: () => dispatchProps.loadData(stateProps.userId, stateProps.keyId),
})

export default UserApiKeyEdit =>
  connect(
    mapStateToProps,
    mapDispatchToProps,
    mergeProps,
  )(
    withRequest(
      ({ loadData }) => loadData(),
      ({ fetching }) => fetching,
    )(UserApiKeyEdit),
  )
