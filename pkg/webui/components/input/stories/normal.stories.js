// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import { IconSearch, IconHammer } from '@ttn-lw/components/icon'
/* eslint-disable react/prop-types */

import Input from '..'

import { Example } from './shared'

// Chosen by fair dice roll.
// Guaranteed to be random.
const generateRandom16Bytes = () => '7d85de9e99c9a2be'

export default {
  title: 'Input/Normal',
}

export const Default = () => (
  <div>
    <Example label="Username" />
    <Example label="Username" warning />
    <Example label="Username" error />
  </div>
)

export const WithPlaceholder = () => <Example placeholder="Placeholder..." />
export const WithIcon = () => <Example icon={IconSearch} />
export const WithAppend = () => <Example error append="test" />

WithIcon.story = {
  name: 'With icon',
}

export const Valid = () => <Example valid />
export const Disabled = () => <Example value="1234" disabled />
export const Readonly = () => <Example value="1234" readOnly />
export const Password = () => <Example type="password" />
export const Number = () => <Example type="number" />

export const Toggled = () => (
  <Example component={Input.Toggled} type="toggled" enabledMessage="Enabled" />
)

export const Textarea = () => <Example type="textarea" />
export const WithSpinner = () => <Example icon={IconSearch} loading />
export const Sensitive = () => <Example sensitive max={5} />

export const WithAction = () => (
  <div>
    <Example action={{ icon: IconHammer, secondary: true }} />
    <Example action={{ icon: IconHammer, secondary: true }} warning />
    <Example action={{ icon: IconHammer, secondary: true }} error />
  </div>
)

export const Generate = () => (
  <Example
    type="byte"
    component={Input.Generate}
    onGenerateValue={generateRandom16Bytes}
    min={16}
    max={16}
  />
)
