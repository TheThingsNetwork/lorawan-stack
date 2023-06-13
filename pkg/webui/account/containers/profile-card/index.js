// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'
import ProfilePicture from '@ttn-lw/components/profile-picture'

import {
  selectUserName,
  selectUserId,
  selectUserProfilePicture,
} from '@account/store/selectors/user'

import style from './profile-card.styl'

const m = defineMessages({
  editProfileSettings: 'Edit profile settings',
})

const ProfileCard = () => {
  const userId = useSelector(selectUserId)
  const userName = useSelector(selectUserName)
  const profilePicture = useSelector(selectUserProfilePicture)

  return (
    <section className={style.container} data-test-id="profile-card">
      <ProfilePicture profilePicture={profilePicture} className={style.profilePicture} />
      <div className={style.panel}>
        <div className={style.name}>
          <h3>{userName || userId}</h3>
          {Boolean(userName) && <span className={style.userId}>{userId}</span>}
        </div>
        <Button.Link to="/profile-settings" icon="edit" message={m.editProfileSettings} />
      </div>
    </section>
  )
}

export default ProfileCard
