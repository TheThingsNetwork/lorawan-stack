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

export const GET_APP_LINK = 'GET_APPLICATION_LINK_REQUEST'
export const GET_APP_LINK_SUCCESS = 'GET_APPLICATION_LINK_SUCCESS'
export const GET_APP_LINK_FAILURE = 'GET_APPLICATION_LINK_FAILURE'
export const UPDATE_APP_LINK_SUCCESS = 'UPDATE_APPLICATION_LINK_SUCCESS'
export const DELETE_APP_LINK_SUCCESS = 'DELETE_APPLICATION_LINK_SUCCESS'

export const getApplicationLink = (id, meta) => (
  { type: GET_APP_LINK, id, meta }
)

export const getApplicationLinkSuccess = (link, stats, linked) => (
  { type: GET_APP_LINK_SUCCESS, link, stats, linked }
)

export const getApplicationLinkFailure = error => (
  { type: GET_APP_LINK_FAILURE, error }
)

export const updateApplicationLinkSuccess = (link, stats) => (
  { type: UPDATE_APP_LINK_SUCCESS, link, stats }
)

export const deleteApplicationLinkSuccess = () => (
  { type: DELETE_APP_LINK_SUCCESS }
)
