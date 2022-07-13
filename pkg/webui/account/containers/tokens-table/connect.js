// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
import { bindActionCreators } from 'redux'
import { push } from 'connected-react-router'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { deleteAuthorizationToken, deleteAllTokens } from '@account/store/actions/authorizations'

import { selectUserId } from '@account/store/selectors/user'
import { selectTokenIds } from '@account/store/selectors/authorizations'

const mapStateToProps = (state, props) => ({
  userId: selectUserId(state),
  clientId: props.match.params.clientId,
  tokenIds: selectTokenIds(state),
})

const mapDispatchToProps = dispatch => ({
  ...bindActionCreators(
    {
      deleteToken: attachPromise(deleteAuthorizationToken),
      deleteAllTokens: attachPromise(deleteAllTokens),
    },
    dispatch,
  ),
  navigateToList: clientId => dispatch(push(`/client-authorizations/${clientId}`)),
})

const mergeProps = (stateProps, dispatchProps, ownProps) => ({
  ...stateProps,
  ...dispatchProps,
  ...ownProps,
  deleteToken: id => dispatchProps.deleteToken(stateProps.userId, stateProps.clientId, id),
  deleteAllTokens: ids =>
    dispatchProps.deleteAllTokens(stateProps.userId, stateProps.clientId, ids),
})

export default Tokens => connect(mapStateToProps, mapDispatchToProps, mergeProps)(Tokens)
