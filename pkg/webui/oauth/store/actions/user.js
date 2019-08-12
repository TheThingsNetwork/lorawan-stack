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

export const GET_USER_ME = 'GET_OAUTH_USER_ME'
export const GET_USER_ME_SUCCESS = 'GET_OAUTH_USER_SUCCESS_ME'
export const GET_USER_ME_FAILURE = 'GET_OAUTH_USER_FAILURE_ME'
export const LOGOUT = 'LOGOUT'
export const LOGOUT_SUCCESS = 'LOGOUT_SUCCESS'
export const LOGOUT_FAILURE = 'LOGOUT_FAILURE'

export const getUserMe = () => ({ type: GET_USER_ME })

export const getUserMeSuccess = user => ({ type: GET_USER_ME_SUCCESS, user })

export const getUserMeFailure = error => ({ type: GET_USER_ME_FAILURE, error })

export const logout = () => ({ type: LOGOUT })

export const logoutSuccess = () => ({ type: LOGOUT_SUCCESS })

export const logoutFailure = error => ({ type: LOGOUT_FAILURE, error })
