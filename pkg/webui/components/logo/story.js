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
import { storiesOf } from '@storybook/react'

import TtsLogo from '../../assets/static/logo.svg'
import ExampleLogo from './story-logo.svg'
import Logo from '.'

storiesOf('Logo', module)
  .add('Default', () => <Logo logo={{ src: TtsLogo, alt: 'Logo' }} />)
  .add('With secondary Logo', () => (
    <Logo
      logo={{ src: TtsLogo, alt: 'Logo' }}
      secondaryLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
    />
  ))
  .add('With vertical secondary', () => (
    <Logo
      vertical
      logo={{ src: TtsLogo, alt: 'Logo' }}
      secondaryLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
    />
  ))
