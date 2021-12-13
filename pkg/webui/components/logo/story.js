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

import React from 'react'

import TtsLogo from '@assets/static/logo.svg'

import ExampleLogo from './story-logo.svg'
import ExampleSquareLogo from './story-logo-2.svg'

import Logo from '.'

export default {
  title: 'Logo',
}

export const Default = () => <Logo logo={{ src: TtsLogo, alt: 'Logo' }} />

export const WithSecondaryLogo = () => (
  <Logo
    logo={{ src: TtsLogo, alt: 'Logo' }}
    brandLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
  />
)

WithSecondaryLogo.story = {
  name: 'With secondary Logo',
}

export const WithSquareShapeSecondaryLogo = () => (
  <Logo
    logo={{ src: TtsLogo, alt: 'Logo' }}
    brandLogo={{ src: ExampleSquareLogo, alt: 'Secondary Logo' }}
  />
)

WithSquareShapeSecondaryLogo.story = {
  name: 'With square-shape secondary Logo',
}

export const WithSecondaryLogoVertical = () => (
  <Logo
    vertical
    logo={{ src: TtsLogo, alt: 'Logo' }}
    brandLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
  />
)

WithSecondaryLogoVertical.story = {
  name: 'With secondary logo, vertical',
}

export const WithSquareShapeSecondaryLogoVertical = () => (
  <Logo
    vertical
    logo={{ src: TtsLogo, alt: 'Logo' }}
    brandLogo={{ src: ExampleSquareLogo, alt: 'Secondary Logo' }}
  />
)

WithSquareShapeSecondaryLogoVertical.story = {
  name: 'With square-shape secondary logo, vertical',
}
