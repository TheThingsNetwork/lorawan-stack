// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

export const getUserMe = () => (
  { type: 'GET_USER_ME' }
)

export const getUserMeFailure = () => (
  { type: 'GET_USER_ME_FAILURE' }
)

export const getUserMeSuccess = userData => (
  { type: 'GET_USER_ME_SUCCESS', userData }
)


export const logout = () => (
  { type: 'LOGOUT' }
)

export const logoutSuccess = () => (
  { type: 'LOGOUT_SUCCESS' }
)

export const logoutFailure = () => (
  { type: 'LOGOUT_FAILURE' }
)
