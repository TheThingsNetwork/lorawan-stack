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

import style from './dropdown.styl'

import Dropdown from '.'

export default {
  title: 'Dropdown V2',
  component: Dropdown,
}

export const Default = () => (
  <div style={{ height: '8rem' }}>
    <Dropdown className={style.example} open>
      <Dropdown.HeaderItem title="dropdown items" />
      <Dropdown.Item title="Profile Settings" path="profile/path" icon="person" />
      <Dropdown.Item title="Admin panel" path="admin/path" icon="admin_panel_settings" />
      <hr />
      <Dropdown.Item title="Logout" path="logout/path" icon="logout" />
    </Dropdown>
  </div>
)
