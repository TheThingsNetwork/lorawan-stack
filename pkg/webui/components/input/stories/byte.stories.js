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

/* eslint-disable react/prop-types */

import React from 'react'

import Input from '..'

import { Example } from './shared'

// Chosen by fair dice roll.
// Guaranteed to be random.
const generateRandom16Bytes = () => '1c3bca1a8f3df30f'

export default {
  title: 'Input/Byte',
}

export const Byte = () => <Example type="byte" min={1} max={5} />
export const ByteReadOnly = () => (
  <Example type="byte" min={1} max={5} value="A0BF49A464" readOnly />
)

ByteReadOnly.story = {
  name: 'Byte read-only',
}

export const Sensitive = () => <Example type="byte" sensitive max={5} />

export const Generate = () => (
  <Example
    type="byte"
    component={Input.Generate}
    onGenerateValue={generateRandom16Bytes}
    min={16}
    max={16}
  />
)
