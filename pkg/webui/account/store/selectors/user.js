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

const selectUserStore = state => state.user

export const selectUser = state => selectUserStore(state).user

export const selectUserId = state => {
  const user = selectUser(state)
  const { ids = {} } = user

  return ids.user_id
}

export const selectUserIsAdmin = state => {
  const user = selectUser(state)
  return user.isAdmin
}

export const selectUserName = state => selectUser(state).name

export const selectUserProfilePicture = state => selectUser(state).profile_picture

export const selectUserRights = state => selectUserStore(state).rights
