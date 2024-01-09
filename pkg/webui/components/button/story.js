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
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/7pBLWK4tsjoAbyJq2viMAQ/2023-console-redesign?type=design&node-id=590%3A51411&mode=design&t=Hbk2Qngeg1xqg4V3-1',
    },
  },
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
    <div style={{ textAlign: 'center', height: '6rem', paddingTop: '4rem' }}>
      <Button primary icon="favorite" message="Dropdown" ref={ref} dropdownItems={dropdownItems} />
    </div>
  )
}

export const PrimayOnlyIconDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem', paddingTop: '4rem' }}>
      <Button primary icon="favorite" dropdownItems={dropdownItems} ref={ref} />
    </div>
  )
}

export const Secondary = () => (
  <div style={{ textAlign: 'center' }}>
    <Button secondary message="Secondary" />
  </div>
)

export const SecondaryWithIcon = () => (
  <div style={{ textAlign: 'center' }}>
    <Button secondary icon="favorite" message="Secondary With Icon" />
  </div>
)

export const SecondaryOnlyIcon = () => (
  <div style={{ textAlign: 'center' }}>
    <Button secondary icon="favorite" />
  </div>
)

export const SecondaryDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem', paddingTop: '4rem' }}>
      <Button
        secondary
        icon="favorite"
        message="Dropdown"
        dropdownItems={dropdownItems}
        ref={ref}
      />
    </div>
  )
}

export const SecondaryOnlyIconDropdown = () => {
  const ref = useRef()

  return (
    <div style={{ textAlign: 'center', height: '6rem', paddingTop: '4rem' }}>
      <Button secondary icon="favorite" dropdownItems={dropdownItems} ref={ref} />
    </div>
  )
}
