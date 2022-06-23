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

import tts from '@console/api/tts'

import withRequest from '@ttn-lw/lib/components/with-request'

import { getUsersRightsList } from '@console/store/actions/users'

import { selectUserId } from '@console/store/selectors/logout'
import {
  selectUserRights,
  selectUserRightsFetching,
  selectUserRightsError,
  selectUserPseudoRights,
} from '@console/store/selectors/users'

export default UserApiKeyAdd =>
  connect(
    state => ({
      userId: selectUserId(state),
      fetching: selectUserRightsFetching(state),
      error: selectUserRightsError(state),
      rights: selectUserRights(state),
      pseudoRights: selectUserPseudoRights(state),
    }),
    dispatch => ({
      navigateToList: () => dispatch(replace(`/user/api-keys`)),
      createUserApiKey: tts.Users.ApiKeys.create,
      getUsersRightsList: userId => dispatch(getUsersRightsList(userId)),
    }),
    (stateProps, dispatchProps, ownProps) => ({
      ...stateProps,
      ...dispatchProps,
      ...ownProps,
      getUsersRightsList: () => dispatchProps.getUsersRightsList(stateProps.userId),
      createUserApiKey: key => dispatchProps.createUserApiKey(stateProps.userId, key),
    }),
  )(
    withRequest(
      ({ userId, getUsersRightsList }) => getUsersRightsList(userId),
      ({ fetching, rights }) => fetching || rights.length === 0,
    )(UserApiKeyAdd),
  )
