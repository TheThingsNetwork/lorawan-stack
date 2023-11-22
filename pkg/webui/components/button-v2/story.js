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

import Button from '.'

export default {
  title: 'Button V2',
}

export const Primary = () => (
  <div>
    <Button primary message="Primary" />
  </div>
)

export const WithIcon = () => (
  <div>
    <Button primary icon="favorite" message="With Icon" />
  </div>
)

export const PrimayOnlyIcon = () => (
  <div>
    <Button primary icon="favorite" />
  </div>
)

export const PrimayDropdown = () => (
  <div>
    <Button primary icon="favorite" message="Dropdown" withDropdown />
  </div>
)

export const PrimayOnlyIconDropdown = () => (
  <div>
    <Button primary icon="favorite" withDropdown />
  </div>
)

export const Naked = () => (
  <div>
    <Button naked message="Naked" />
  </div>
)

export const NakedWithIcon = () => (
  <div>
    <Button naked icon="favorite" message="Naked With Icon" />
  </div>
)

export const NakedOnlyIcon = () => (
  <div>
    <Button naked icon="favorite" />
  </div>
)

export const nakedDropdown = () => (
  <div>
    <Button naked icon="favorite" message="Dropdown" withDropdown />
  </div>
)

export const NakedOnlyIconDropdown = () => (
  <div>
    <Button naked icon="favorite" withDropdown />
  </div>
)
