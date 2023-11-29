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

import React, { useRef } from 'react'

import Dropdown from '@ttn-lw/components/dropdown'

import Button from '.'

export default {
  title: 'Button V2',
}

const dropdownItems = (
  <React.Fragment>
    <Dropdown.Item title="Profile Settings" icon="settings" path="/profile-settings" />
    <Dropdown.Item title="Logout" icon="power_settings_new" path="/logout" />
  </React.Fragment>
)

export const Primary = () => (
  <div style={{ textAlign: 'center' }}>
    <Button primary message="Primary" />
  </div>
)

export const WithIcon = () => (
  <div style={{ textAlign: 'center' }}>
    <Button primary icon="favorite" message="With Icon" />
  </div>
)

export const PrimayOnlyIcon = () => (
  <div style={{ textAlign: 'center' }}>
    <Button primary icon="favorite" />
  </div>
)

export const PrimayDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem' }}>
      <Button primary icon="favorite" message="Dropdown" ref={ref} dropdownItems={dropdownItems} />
    </div>
  )
}

export const PrimayOnlyIconDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem' }}>
      <Button primary icon="favorite" dropdownItems={dropdownItems} ref={ref} />
    </div>
  )
}

export const Naked = () => (
  <div style={{ textAlign: 'center' }}>
    <Button naked message="Naked" />
  </div>
)

export const NakedWithIcon = () => (
  <div style={{ textAlign: 'center' }}>
    <Button naked icon="favorite" message="Naked With Icon" />
  </div>
)

export const NakedOnlyIcon = () => (
  <div style={{ textAlign: 'center' }}>
    <Button naked icon="favorite" />
  </div>
)

export const NakedDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem' }}>
      <Button naked icon="favorite" message="Dropdown" dropdownItems={dropdownItems} ref={ref} />
    </div>
  )
}

export const NakedOnlyIconDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem' }}>
      <Button naked icon="favorite" dropdownItems={dropdownItems} ref={ref} />
    </div>
  )
}
