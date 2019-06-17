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
} from '../actions/link'

const defaultProps = {
  linked: false,
  link: undefined,
  stats: undefined,
}

const getLinkSuccess = function (state, { payload }) {
  const { linked, stats, link = {}} = payload

  return {
    ...state,
    linked,
    link,
    stats,
  }
}

const getLinkFailure = function (state) {
  return {
    ...state,
    link: {},
    stats: undefined,
    linked: false,
  }
}

const updateLinkSuccess = function (state, { link, stats }) {
  const newLink = { ...state.link, ...link }
  const newStats = { ...state.stats, ...stats }

  return {
    ...state,
    linked: true,
    link: newLink,
    stats: newStats,
  }
}

const deleteLinkSuccess = function (state) {
  return {
    ...state,
    linked: false,
    link: {},
    stats: undefined,
  }
}

const link = function (state = defaultProps, action) {
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
