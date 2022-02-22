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

import {
  GET_APP_LINK_SUCCESS,
  GET_APP_LINK_FAILURE,
  UPDATE_APP_LINK_SUCCESS,
  DELETE_APP_LINK_SUCCESS,
} from '@console/store/actions/link'

const defaultProps = {
  link: undefined,
}

const getLinkSuccess = (state, { payload }) => {
  const { link = {} } = payload

  return {
    ...state,
    link,
  }
}

const getLinkFailure = (state, { payload }) => ({
  ...state,
  link: payload.link || {},
})

const updateLinkSuccess = (state, link) => {
  const newLink = { ...state.link, ...link.payload }
  return {
    ...state,
    link: newLink,
  }
}

const deleteLinkSuccess = state => ({
  ...state,
  link: {},
})

const link = (state = defaultProps, action) => {
  switch (action.type) {
    case GET_APP_LINK_SUCCESS:
      return getLinkSuccess(state, action)
    case UPDATE_APP_LINK_SUCCESS:
      return updateLinkSuccess(state, action)
    case DELETE_APP_LINK_SUCCESS:
      return deleteLinkSuccess(state)
    case GET_APP_LINK_FAILURE:
      return getLinkFailure(state, action)
    default:
      return state
  }
}

export default link
